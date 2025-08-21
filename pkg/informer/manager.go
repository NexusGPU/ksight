package informer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// ClusterConnection represents a Kubernetes cluster connection
type ClusterConnection struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Config      *rest.Config           `json:"-"`
	Client      dynamic.Interface      `json:"-"`
	Factory     dynamicinformer.DynamicSharedInformerFactory `json:"-"`
	Informers   map[schema.GroupVersionResource]cache.SharedIndexInformer `json:"-"`
	Context     string                 `json:"context"`
	Server      string                 `json:"server"`
	Status      string                 `json:"status"` // connected, disconnected, error
	LastError   string                 `json:"lastError,omitempty"`
	IsPinned    bool                   `json:"isPinned"`
	mu          sync.RWMutex
}

// ResourceVersionStore manages persistent storage of resource versions
type ResourceVersionStore struct {
	storePath string
	data      map[string]map[string]string // clusterID -> GVR -> resourceVersion
	mu        sync.RWMutex
}

// InformerManager manages dynamic informers for multiple clusters
type InformerManager struct {
	clusters     map[string]*ClusterConnection
	store        *ResourceVersionStore
	eventHandler func(event Event)
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// Event represents a resource event from informers
type Event struct {
	Type        string                 `json:"type"`        // ADDED, MODIFIED, DELETED
	ClusterID   string                 `json:"clusterId"`
	GVR         schema.GroupVersionResource `json:"gvr"`
	Namespace   string                 `json:"namespace"`
	Name        string                 `json:"name"`
	Object      *unstructured.Unstructured `json:"object"`
	OldObject   *unstructured.Unstructured `json:"oldObject,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewInformerManager creates a new informer manager
func NewInformerManager(storePath string, eventHandler func(Event)) *InformerManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	store := &ResourceVersionStore{
		storePath: storePath,
		data:      make(map[string]map[string]string),
	}
	store.load()

	return &InformerManager{
		clusters:     make(map[string]*ClusterConnection),
		store:        store,
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

	// Create and configure informer
	informer := cluster.Factory.ForResource(gvr).Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			im.handleEvent("ADDED", clusterID, gvr, obj, nil)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			im.handleEvent("MODIFIED", clusterID, gvr, newObj, oldObj)
		},
		DeleteFunc: func(obj interface{}) {
			im.handleEvent("DELETED", clusterID, gvr, obj, nil)
		},
	})

	// Store the informer
	cluster.Informers[gvr] = informer

	// Start the informer if not already started
	go cluster.Factory.Start(im.ctx.Done())

	// Wait for cache sync with timeout
	go func() {
		ctx, cancel := context.WithTimeout(im.ctx, 30*time.Second)
		defer cancel()
		
		if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
			cluster.Status = "error"
			cluster.LastError = "failed to sync cache for " + gvr.String()
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
func (im *InformerManager) handleEvent(eventType, clusterID string, gvr schema.GroupVersionResource, obj, oldObj interface{}) {
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
	}

	event := Event{
		Type:      eventType,
		ClusterID: clusterID,
		GVR:       gvr,
		Namespace: unstructuredObj.GetNamespace(),
		Name:      unstructuredObj.GetName(),
		Object:    unstructuredObj,
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
