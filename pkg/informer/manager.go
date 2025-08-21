package informer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/sjson"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	_ "modernc.org/sqlite"
)

// ClusterConnection represents a Kubernetes cluster connection
type ClusterConnection struct {
	ID        string                                                    `json:"id"`
	Name      string                                                    `json:"name"`
	Config    *rest.Config                                              `json:"-"`
	Client    dynamic.Interface                                         `json:"-"`
	Factory   dynamicinformer.DynamicSharedInformerFactory              `json:"-"`
	Informers map[schema.GroupVersionResource]cache.SharedIndexInformer `json:"-"`
	Context   string                                                    `json:"context"`
	Server    string                                                    `json:"server"`
	Status    string                                                    `json:"status"` // connected, disconnected, error
	LastError string                                                    `json:"lastError,omitempty"`
	IsPinned  bool                                                      `json:"isPinned"`
	mu        sync.RWMutex
}

// ResourceVersionStore manages persistent storage of resource versions
type ResourceVersionStore struct {
	storePath string
	data      map[string]map[string]string // clusterID -> GVR -> resourceVersion
	mu        sync.RWMutex
}

// SensitiveResourceConfig defines which resources and fields are sensitive
type SensitiveResourceConfig struct {
	// Map of GroupKind key to list of sensitive field paths
	Resources map[string][]string `json:"resources"`
}

type SensitiveGroupKind struct {
	Group string `json:"group"`
	Kind  string `json:"kind"`
}

// Key returns the map key for this GroupKind
func (gk SensitiveGroupKind) Key() string {
	return gk.Group + "/" + gk.Kind
}

// DatabaseCache manages resource caching using SQLite
type DatabaseCache struct {
	db              *sql.DB
	cacheDir        string
	mu              sync.RWMutex
	sensitiveConfig *SensitiveResourceConfig
}

// RedactedResource represents a resource with sensitive data redacted
type RedactedResource struct {
	Original        *unstructured.Unstructured `json:"original"`
	Redacted        *unstructured.Unstructured `json:"redacted"`
	SensitiveFields []string                   `json:"sensitiveFields"`
	IsFullyRedacted bool                       `json:"isFullyRedacted"`
}

// InformerManager manages dynamic informers for multiple clusters
type InformerManager struct {
	clusters     map[string]*ClusterConnection
	store        *ResourceVersionStore
	dbCache      *DatabaseCache
	eventHandler func(event Event)
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// DefaultSensitiveConfig is the default configuration for sensitive resources
var DefaultSensitiveConfig = &SensitiveResourceConfig{
	Resources: map[string][]string{
		"/Secret": {
			"data",
			"stringData",
		},
		"external-secrets.io/SecretStore": {
			"spec.provider",
			"spec.auth",
		},
		"external-secrets.io/ClusterSecretStore": {
			"spec.provider",
			"spec.auth",
		},
		"bitnami.com/SealedSecret": {
			"spec.encryptedData",
		},
		"cert-manager.io/Certificate": {
			"spec.privateKey",
			"spec.keystores",
		},
	},
}

// Event represents a resource event from informers
type Event struct {
	Type      string                      `json:"type"` // ADDED, MODIFIED, DELETED
	ClusterID string                      `json:"clusterId"`
	GVR       schema.GroupVersionResource `json:"gvr"`
	Namespace string                      `json:"namespace"`
	Name      string                      `json:"name"`
	Object    *unstructured.Unstructured  `json:"object"`
	OldObject *unstructured.Unstructured  `json:"oldObject,omitempty"`
	Timestamp time.Time                   `json:"timestamp"`
}

// NewInformerManager creates a new informer manager
func NewInformerManager(storePath string, eventHandler func(Event)) *InformerManager {
	ctx, cancel := context.WithCancel(context.Background())

	store := &ResourceVersionStore{
		storePath: storePath,
		data:      make(map[string]map[string]string),
	}
	store.load()

	// Create database cache in same directory as version store
	cacheDir := filepath.Join(filepath.Dir(storePath), "cache")
	os.MkdirAll(cacheDir, 0755)

	dbCache, err := NewDatabaseCache(cacheDir, DefaultSensitiveConfig)
	if err != nil {
		// Log error but continue without cache
		fmt.Printf("Warning: Failed to initialize database cache: %v\n", err)
	}

	return &InformerManager{
		clusters:     make(map[string]*ClusterConnection),
		store:        store,
		dbCache:      dbCache,
		eventHandler: eventHandler,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// AddCluster adds a new cluster connection
func (im *InformerManager) AddCluster(id, name, kubeconfig, context string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Parse kubeconfig
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		// Try loading from file path
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return fmt.Errorf("failed to parse kubeconfig: %w", err)
		}
	}

	// Create dynamic client
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Create informer factory
	factory := dynamicinformer.NewDynamicSharedInformerFactory(client, 30*time.Second)

	cluster := &ClusterConnection{
		ID:        id,
		Name:      name,
		Config:    config,
		Client:    client,
		Factory:   factory,
		Informers: make(map[schema.GroupVersionResource]cache.SharedIndexInformer),
		Context:   context,
		Server:    config.Host,
		Status:    "connected",
	}

	im.clusters[id] = cluster

	// Initialize store for this cluster if not exists
	if _, exists := im.store.data[id]; !exists {
		im.store.data[id] = make(map[string]string)
	}

	return nil
}

// RemoveCluster removes a cluster connection and stops all its informers
func (im *InformerManager) RemoveCluster(id string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	cluster, exists := im.clusters[id]
	if !exists {
		return fmt.Errorf("cluster %s not found", id)
	}

	cluster.mu.Lock()
	defer cluster.mu.Unlock()

	// Stop all informers for this cluster
	for gvr := range cluster.Informers {
		// Stop the informer (this will be handled by factory stop)
		delete(cluster.Informers, gvr)
	}

	// Stop the factory
	cluster.Factory.Shutdown()

	// Remove from clusters map
	delete(im.clusters, id)

	// Clean up store data for this cluster
	delete(im.store.data, id)
	im.store.save()

	return nil
}

// AddResourceWatcher adds a watcher for a specific GVK
func (im *InformerManager) AddResourceWatcher(clusterID string, gvr schema.GroupVersionResource, namespace string) error {
	im.mu.RLock()
	cluster, exists := im.clusters[clusterID]
	im.mu.RUnlock()

	if !exists {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.mu.Lock()
	defer cluster.mu.Unlock()

	// Check if informer already exists
	if _, exists := cluster.Informers[gvr]; exists {
		return nil // Already watching
	}

	// Get last known resource version for this GVR (for future use)
	// lastResourceVersion := im.store.getResourceVersion(clusterID, gvr.String())

	// Test API access first to handle 401 errors
	if err := im.testAPIAccess(cluster, gvr, namespace); err != nil {
		cluster.Status = "error"
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Unauthorized") {
			cluster.LastError = fmt.Sprintf("Unauthorized access to %s: %v", gvr.String(), err)
			return fmt.Errorf("unauthorized access to %s in cluster %s: %w", gvr.String(), clusterID, err)
		}
		cluster.LastError = fmt.Sprintf("Failed to access %s: %v", gvr.String(), err)
		return err
	}

	// Create and configure informer
	informer := cluster.Factory.ForResource(gvr).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			im.handleEvent("ADDED", clusterID, gvr, obj, nil)
		},
		UpdateFunc: func(oldObj, newObj any) {
			im.handleEvent("MODIFIED", clusterID, gvr, newObj, oldObj)
		},
		DeleteFunc: func(obj any) {
			im.handleEvent("DELETED", clusterID, gvr, obj, nil)
		},
	})

	// Store the informer
	cluster.Informers[gvr] = informer

	// Start the informer if not already started
	go cluster.Factory.Start(im.ctx.Done())

	// Wait for cache sync with timeout in a separate goroutine
	go func() {
		ctx, cancel := context.WithTimeout(im.ctx, 30*time.Second)
		defer cancel()

		if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
			cluster.mu.Lock()
			cluster.Status = "error"
			cluster.LastError = "failed to sync cache for " + gvr.String()
			cluster.mu.Unlock()
		}
	}()

	return nil
}

// RemoveResourceWatcher removes a watcher for a specific GVK
func (im *InformerManager) RemoveResourceWatcher(clusterID string, gvr schema.GroupVersionResource) error {
	im.mu.RLock()
	cluster, exists := im.clusters[clusterID]
	im.mu.RUnlock()

	if !exists {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.mu.Lock()
	defer cluster.mu.Unlock()

	// Check if informer exists
	if _, exists := cluster.Informers[gvr]; !exists {
		return fmt.Errorf("watcher for %s not found", gvr.String())
	}

	// Remove the informer
	delete(cluster.Informers, gvr)

	// Note: We can't actually stop individual informers in the factory,
	// but removing from our map will stop event handling

	return nil
}

// GetClusters returns all cluster connections
func (im *InformerManager) GetClusters() map[string]*ClusterConnection {
	im.mu.RLock()
	defer im.mu.RUnlock()

	result := make(map[string]*ClusterConnection)
	for id, cluster := range im.clusters {
		result[id] = cluster
	}
	return result
}

// handleEvent processes informer events and forwards them to the event handler
func (im *InformerManager) handleEvent(eventType, clusterID string, gvr schema.GroupVersionResource, obj, oldObj any) {
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return
	}

	var unstructuredOldObj *unstructured.Unstructured
	if oldObj != nil {
		if old, ok := oldObj.(*unstructured.Unstructured); ok {
			unstructuredOldObj = old
		}
	}

	// Update resource version in store
	resourceVersion := unstructuredObj.GetResourceVersion()
	if resourceVersion != "" {
		im.store.setResourceVersion(clusterID, gvr.String(), resourceVersion)

		// Cache the resource in database (non-blocking)
		if im.dbCache != nil {
			go func() {
				if err := im.dbCache.StoreResource(clusterID, gvr, unstructuredObj); err != nil {
					// Log error but don't fail the event processing
					fmt.Printf("Warning: Failed to cache resource: %v\\n", err)
				}
			}()
		}
	}

	// Create event with redacted object for sensitive resources
	var eventObj *unstructured.Unstructured
	if im.dbCache != nil && im.dbCache.isSensitiveResource(gvr, unstructuredObj) {
		eventObj = im.dbCache.redactSensitiveFields(unstructuredObj)
	} else {
		eventObj = unstructuredObj
	}

	event := Event{
		Type:      eventType,
		ClusterID: clusterID,
		GVR:       gvr,
		Namespace: unstructuredObj.GetNamespace(),
		Name:      unstructuredObj.GetName(),
		Object:    eventObj,
		OldObject: unstructuredOldObj,
		Timestamp: time.Now(),
	}

	if im.eventHandler != nil {
		im.eventHandler(event)
	}
}

// Shutdown stops all informers and saves state
func (im *InformerManager) Shutdown() {
	im.cancel()

	im.mu.Lock()
	defer im.mu.Unlock()

	for _, cluster := range im.clusters {
		cluster.Factory.Shutdown()
	}

	im.store.save()
}

// CleanupTestData removes all cache files for testing purposes
func (im *InformerManager) CleanupTestData() {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Clear in-memory store
	im.store.data = make(map[string]map[string]string)

	// Close and remove database cache
	if im.dbCache != nil {
		im.dbCache.Close()
		if im.dbCache.cacheDir != "" {
			os.RemoveAll(im.dbCache.cacheDir)
		}
	}

	// Remove version store file
	if im.store.storePath != "" {
		os.Remove(im.store.storePath)
	}
}

// ResourceVersionStore methods

func (rvs *ResourceVersionStore) load() {
	rvs.mu.Lock()
	defer rvs.mu.Unlock()

	if _, err := os.Stat(rvs.storePath); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(rvs.storePath)
	if err != nil {
		return
	}

	json.Unmarshal(data, &rvs.data)
}

func (rvs *ResourceVersionStore) save() {
	rvs.mu.RLock()
	defer rvs.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(rvs.storePath)
	os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(rvs.data, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(rvs.storePath, data, 0644)
}

func (rvs *ResourceVersionStore) getResourceVersion(clusterID, gvr string) string {
	rvs.mu.RLock()
	defer rvs.mu.RUnlock()

	if cluster, exists := rvs.data[clusterID]; exists {
		return cluster[gvr]
	}
	return ""
}

func (rvs *ResourceVersionStore) setResourceVersion(clusterID, gvr, version string) {
	rvs.mu.Lock()
	defer rvs.mu.Unlock()

	if _, exists := rvs.data[clusterID]; !exists {
		rvs.data[clusterID] = make(map[string]string)
	}

	rvs.data[clusterID][gvr] = version

	// Save periodically (could be optimized with a background goroutine)
	go rvs.save()
}

// DatabaseCache needs to be accessible for benchmarks
type DatabaseCacheInterface interface {
	StoreResource(clusterID string, gvr schema.GroupVersionResource, resource *unstructured.Unstructured) error
	GetResource(clusterID string, gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, bool, error)
	LoadResources(clusterID string, gvr schema.GroupVersionResource) ([]*unstructured.Unstructured, string, error)
	GetCacheStats() (map[string]int, error)
	Close() error
}

// NewDatabaseCache creates a new database cache
func NewDatabaseCache(cacheDir string, config *SensitiveResourceConfig) (*DatabaseCache, error) {
	// Use unique database file to avoid conflicts during parallel tests
	dbPath := filepath.Join(cacheDir, "resource_cache.db")

	db, err := sql.Open("sqlite", dbPath+"?_busy_timeout=10000&_journal_mode=MEMORY")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	cache := &DatabaseCache{
		db:              db,
		cacheDir:        cacheDir,
		sensitiveConfig: config,
	}

	if err := cache.initTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database tables: %w", err)
	}

	return cache, nil
}

// initTables creates the necessary database tables
func (dc *DatabaseCache) initTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS resource_cache (
		uid VARCHAR(36) PRIMARY KEY,
		cluster_id VARCHAR(255) NOT NULL,
		gvr VARCHAR(255) NOT NULL,
		namespace VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		resource_version VARCHAR(16) NOT NULL,
		data TEXT NOT NULL,
		is_sensitive BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(cluster_id, gvr, namespace, name)
	);
	CREATE INDEX IF NOT EXISTS idx_updated_at ON resource_cache(updated_at);
	`

	_, err := dc.db.Exec(schema)
	return err
}

func (dc *DatabaseCache) StoreResource(clusterID string, gvr schema.GroupVersionResource, resource *unstructured.Unstructured) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Always remove last-applied-configuration annotation if present
	annotations := resource.GetAnnotations()
	if annotations != nil {
		delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
	}

	// Check if resource is sensitive
	isSensitive := dc.isSensitiveResource(gvr, resource)

	var dataToStore []byte
	var err error

	if isSensitive {
		// Store redacted version in cache
		redacted := dc.redactSensitiveFields(resource)
		dataToStore, err = json.Marshal(redacted)
	} else {
		dataToStore, err = json.Marshal(resource)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal resource: %w", err)
	}

	query := `
		INSERT OR REPLACE INTO resource_cache 
		(uid, cluster_id, gvr, namespace, name, resource_version, data, is_sensitive, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err = dc.db.Exec(query,
		string(resource.GetUID()),
		clusterID,
		gvr.String(),
		resource.GetNamespace(),
		resource.GetName(),
		resource.GetResourceVersion(),
		string(dataToStore),
		isSensitive,
	)

	return err
}

// GetResource retrieves a resource from cache
func (dc *DatabaseCache) GetResource(clusterID string, gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, bool, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	query := `
		SELECT data, is_sensitive FROM resource_cache 
		WHERE cluster_id = ? AND gvr = ? AND namespace = ? AND name = ?
	`

	var data string
	var isSensitive bool
	err := dc.db.QueryRow(query, clusterID, gvr.String(), namespace, name).Scan(&data, &isSensitive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, err
	}

	var resource unstructured.Unstructured
	if err := json.Unmarshal([]byte(data), &resource); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal resource: %w", err)
	}

	return &resource, isSensitive, nil
}

// LoadInitialData loads cached resource data from database for faster startup
func (im *InformerManager) LoadInitialData(clusterID string, gvr schema.GroupVersionResource) ([]*unstructured.Unstructured, string, error) {
	if im.dbCache == nil {
		return nil, "", nil
	}

	return im.dbCache.LoadResources(clusterID, gvr)
}

// LoadResources loads all resources for a specific GVR from cache
func (dc *DatabaseCache) LoadResources(clusterID string, gvr schema.GroupVersionResource) ([]*unstructured.Unstructured, string, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	// Only load cache that's not too old (24 hours)
	query := `
		SELECT data, resource_version FROM resource_cache 
		WHERE cluster_id = ? AND gvr = ?
		ORDER BY updated_at DESC
	`

	rows, err := dc.db.Query(query, clusterID, gvr.String())
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var resources []*unstructured.Unstructured
	var latestResourceVersion string

	for rows.Next() {
		var data string
		var resourceVersion string

		if err := rows.Scan(&data, &resourceVersion); err != nil {
			continue // Skip invalid entries
		}

		var resource unstructured.Unstructured
		if err := json.Unmarshal([]byte(data), &resource); err != nil {
			continue // Skip invalid entries
		}

		resources = append(resources, &resource)
		if latestResourceVersion == "" || resourceVersion > latestResourceVersion {
			latestResourceVersion = resourceVersion
		}
	}

	return resources, latestResourceVersion, nil
}

// testAPIAccess tests if we can access the API for a specific GVR
func (im *InformerManager) testAPIAccess(cluster *ClusterConnection, gvr schema.GroupVersionResource, namespace string) error {
	// Try to list resources to test permissions
	listOptions := metav1.ListOptions{Limit: 1}

	if namespace != "" {
		_, err := cluster.Client.Resource(gvr).Namespace(namespace).List(context.TODO(), listOptions)
		return err
	} else {
		_, err := cluster.Client.Resource(gvr).List(context.TODO(), listOptions)
		return err
	}
}

// Close closes the database connection
func (dc *DatabaseCache) Close() error {
	if dc.db != nil {
		return dc.db.Close()
	}
	return nil
}

// isSensitiveResource checks if a resource is considered sensitive
func (dc *DatabaseCache) isSensitiveResource(gvr schema.GroupVersionResource, resource *unstructured.Unstructured) bool {
	if dc.sensitiveConfig == nil || dc.sensitiveConfig.Resources == nil {
		return false
	}

	// Create key from GVR group and resource kind
	key := gvr.Group + "/" + resource.GetKind()

	// Check if this resource type is configured as sensitive
	_, exists := dc.sensitiveConfig.Resources[key]
	return exists
}

// redactSensitiveFields creates a copy of the resource with sensitive fields redacted
func (dc *DatabaseCache) redactSensitiveFields(resource *unstructured.Unstructured) *unstructured.Unstructured {
	if dc.sensitiveConfig == nil || dc.sensitiveConfig.Resources == nil {
		return resource.DeepCopy()
	}

	// Create a deep copy to avoid modifying the original
	redacted := resource.DeepCopy()

	// Get the GVR from the resource
	gvk := resource.GroupVersionKind()
	key := gvk.Group + "/" + gvk.Kind

	// Get field paths for this specific resource type
	fieldPaths, exists := dc.sensitiveConfig.Resources[key]
	if !exists {
		return redacted
	}

	// Redact sensitive field paths for this resource type
	for _, fieldPath := range fieldPaths {
		dc.redactFieldPath(redacted.Object, fieldPath)
	}

	return redacted
}

// redactFieldPath redacts a specific field path in the object using sjson library
func (dc *DatabaseCache) redactFieldPath(obj map[string]any, fieldPath string) {
	// Convert object to JSON
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return // Skip if marshaling fails
	}

	// Convert array notation from [*] to proper sjson format
	// sjson doesn't support [*] wildcard, so we need to handle arrays differently
	if strings.Contains(fieldPath, "[*]") {
		dc.redactArrayFieldPath(obj, fieldPath)
		return
	}

	// Use sjson to set the field to redacted value
	modifiedJSON, err := sjson.Set(string(jsonBytes), fieldPath, "<redacted>")
	if err != nil {
		return // Skip if setting fails
	}

	// Unmarshal back to the object
	var result map[string]any
	if err := json.Unmarshal([]byte(modifiedJSON), &result); err != nil {
		return // Skip if unmarshaling fails
	}

	// Copy the modified fields back to the original object
	for k, v := range result {
		obj[k] = v
	}
}

// redactArrayFieldPath handles field paths with [*] array notation
func (dc *DatabaseCache) redactArrayFieldPath(obj map[string]any, fieldPath string) {
	// Split the path at [*] to handle array wildcards
	parts := strings.Split(fieldPath, "[*]")
	if len(parts) != 2 {
		return // Invalid format
	}

	arrayPath := parts[0]
	remainderPath := strings.TrimPrefix(parts[1], ".")

	// Navigate to the array field
	current := obj
	for _, part := range strings.Split(arrayPath, ".") {
		if part == "" {
			continue
		}
		if next, ok := current[part]; ok {
			if nextMap, ok := next.(map[string]any); ok {
				current = nextMap
			} else {
				return // Path doesn't lead to an object
			}
		} else {
			return // Path doesn't exist
		}
	}

	// Find the array field
	lastPart := arrayPath[strings.LastIndex(arrayPath, ".")+1:]
	if arrayPath == lastPart {
		// No dots in path, array is at root level
		if arrayField, ok := obj[lastPart]; ok {
			dc.redactArrayElements(arrayField, remainderPath)
		}
	} else {
		// Array is nested
		parentPath := arrayPath[:strings.LastIndex(arrayPath, ".")]
		parent := obj
		for _, part := range strings.Split(parentPath, ".") {
			if part == "" {
				continue
			}
			if next, ok := parent[part].(map[string]any); ok {
				parent = next
			} else {
				return
			}
		}
		if arrayField, ok := parent[lastPart]; ok {
			dc.redactArrayElements(arrayField, remainderPath)
		}
	}
}

// redactArrayElements redacts fields in all elements of an array
func (dc *DatabaseCache) redactArrayElements(arrayField any, remainderPath string) {
	if arraySlice, ok := arrayField.([]any); ok {
		for _, element := range arraySlice {
			if elementMap, ok := element.(map[string]any); ok {
				if remainderPath == "" {
					// Redact the entire element
					for k := range elementMap {
						elementMap[k] = "<redacted>"
					}
				} else {
					// Redact specific field in element
					dc.redactFieldPath(elementMap, remainderPath)
				}
			}
		}
	}
}

// GetOriginalResource retrieves the original (non-redacted) resource from API server
func (im *InformerManager) GetOriginalResource(clusterID string, gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
	im.mu.RLock()
	cluster, exists := im.clusters[clusterID]
	im.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("cluster %s not found", clusterID)
	}

	// Always fetch from API server for sensitive resources
	if namespace != "" {
		return cluster.Client.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	} else {
		return cluster.Client.Resource(gvr).Get(context.TODO(), name, metav1.GetOptions{})
	}
}

// GetResourceWithSensitivityInfo gets a resource and indicates if it's sensitive
func (im *InformerManager) GetResourceWithSensitivityInfo(clusterID string, gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, bool, error) {
	// First try to get from cache
	if im.dbCache != nil {
		cached, isSensitive, err := im.dbCache.GetResource(clusterID, gvr, namespace, name)
		if err != nil {
			return nil, false, err
		}

		if cached != nil {
			return cached, isSensitive, nil
		}
	}

	// Fallback to API server
	original, err := im.GetOriginalResource(clusterID, gvr, namespace, name)
	if err != nil {
		return nil, false, err
	}

	// Check if sensitive and return redacted version if needed
	if im.dbCache != nil && im.dbCache.isSensitiveResource(gvr, original) {
		redacted := im.dbCache.redactSensitiveFields(original)
		return redacted, true, nil
	}

	return original, false, nil
}

// CleanOldCache removes cache entries older than the specified duration
func (dc *DatabaseCache) CleanOldCache(maxAge time.Duration) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	query := `DELETE FROM resource_cache WHERE updated_at < datetime('now', '-' || ? || ' seconds')`
	_, err := dc.db.Exec(query, int(maxAge.Seconds()))
	return err
}

// GetCacheStats returns statistics about the cache
func (dc *DatabaseCache) GetCacheStats() (map[string]int, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	stats := make(map[string]int)

	// Total resources
	var total int
	err := dc.db.QueryRow("SELECT COUNT(*) FROM resource_cache").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Sensitive resources
	var sensitive int
	err = dc.db.QueryRow("SELECT COUNT(*) FROM resource_cache WHERE is_sensitive = 1").Scan(&sensitive)
	if err != nil {
		return nil, err
	}
	stats["sensitive"] = sensitive

	// Resources by cluster
	rows, err := dc.db.Query("SELECT cluster_id, COUNT(*) FROM resource_cache GROUP BY cluster_id")
	if err != nil {
		return stats, nil
	}
	defer rows.Close()

	for rows.Next() {
		var clusterID string
		var count int
		if err := rows.Scan(&clusterID, &count); err == nil {
			stats["cluster_"+clusterID] = count
		}
	}

	return stats, nil
}

// GetCacheStats returns cache statistics from the informer manager
func (im *InformerManager) GetCacheStats() (map[string]int, error) {
	if im.dbCache == nil {
		return map[string]int{}, fmt.Errorf("cache not available")
	}
	return im.dbCache.GetCacheStats()
}

// CleanOldCache removes old cache entries
func (im *InformerManager) CleanOldCache(maxAge time.Duration) error {
	if im.dbCache == nil {
		return fmt.Errorf("cache not available")
	}
	return im.dbCache.CleanOldCache(maxAge)
}
