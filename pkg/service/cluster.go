package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"ksight/pkg/informer"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// EventEmitter interface for abstracting event emission
type EventEmitter interface {
	Emit(event string, data interface{})
}

// WailsEventEmitter implements EventEmitter using Wails runtime
type WailsEventEmitter struct {
	ctx context.Context
}

func (w *WailsEventEmitter) Emit(event string, data interface{}) {
	runtime.EventsEmit(w.ctx, event, data)
}

// MockEventEmitter implements EventEmitter for testing
type MockEventEmitter struct{}

func (m *MockEventEmitter) Emit(event string, data interface{}) {
	// Mock implementation - could log or store events for testing
}

// detectEnvironment returns appropriate EventEmitter based on context
func detectEnvironment(ctx context.Context) EventEmitter {
	// Check if we're in a Wails context by looking for specific context values
	if ctx.Value("wails") != nil {
		return &WailsEventEmitter{ctx: ctx}
	}
	// Default to mock for test environments
	return &MockEventEmitter{}
}

// ClusterService handles cluster management and resource watching
type ClusterService struct {
	ctx             context.Context
	informerManager *informer.InformerManager
	dataDir         string
	eventEmitter    EventEmitter
}

// ClusterInfo represents cluster information for frontend
type ClusterInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Context     string `json:"context"`
	Server      string `json:"server"`
	Status      string `json:"status"`
	LastError   string `json:"lastError,omitempty"`
	IsPinned    bool   `json:"isPinned"`
}

// ResourceWatchRequest represents a request to watch resources
type ResourceWatchRequest struct {
	ClusterID   string `json:"clusterId"`
	Group       string `json:"group"`
	Version     string `json:"version"`
	Resource    string `json:"resource"`
	Namespace   string `json:"namespace,omitempty"`
}

// NewClusterService creates a new cluster service
func NewClusterService(ctx context.Context) *ClusterService {
	// Get user data directory
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".ksight")
	os.MkdirAll(dataDir, 0755)

	cs := &ClusterService{
		ctx:          ctx,
		dataDir:      dataDir,
		eventEmitter: detectEnvironment(ctx),
	}

	// Create informer manager with event handler
	manager := informer.NewInformerManager(
		filepath.Join(dataDir, "resource_versions.json"),
		func(event informer.Event) {
			// Emit event to frontend
			cs.eventEmitter.Emit("resource:event", event)
		},
	)

	cs.informerManager = manager
	return cs
}

// AddCluster adds a new cluster connection
func (cs *ClusterService) AddCluster(name, kubeconfig, context string) (string, error) {
	// Generate cluster ID
	clusterID := fmt.Sprintf("cluster_%d", len(cs.informerManager.GetClusters())+1)
	
	err := cs.informerManager.AddCluster(clusterID, name, kubeconfig, context)
	if err != nil {
		return "", err
	}

	// Emit cluster added event
	clusters := cs.GetClusters()
	cs.eventEmitter.Emit("cluster:added", clusters[clusterID])

	return clusterID, nil
}

// RemoveCluster removes a cluster connection
func (cs *ClusterService) RemoveCluster(clusterID string) error {
	err := cs.informerManager.RemoveCluster(clusterID)
	if err != nil {
		return err
	}

	// Emit cluster removed event
	cs.eventEmitter.Emit("cluster:removed", clusterID)

	return nil
}

// GetClusters returns all cluster connections
func (cs *ClusterService) GetClusters() map[string]ClusterInfo {
	clusters := cs.informerManager.GetClusters()
	result := make(map[string]ClusterInfo)

	for id, cluster := range clusters {
		result[id] = ClusterInfo{
			ID:        cluster.ID,
			Name:      cluster.Name,
			Context:   cluster.Context,
			Server:    cluster.Server,
			Status:    cluster.Status,
			LastError: cluster.LastError,
			IsPinned:  cluster.IsPinned,
		}
	}

	return result
}

// ToggleClusterPin toggles the pinned state of a cluster
func (cs *ClusterService) ToggleClusterPin(clusterID string) error {
	clusters := cs.informerManager.GetClusters()
	cluster, exists := clusters[clusterID]
	if !exists {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.IsPinned = !cluster.IsPinned

	// Emit cluster updated event
	clusterInfo := ClusterInfo{
		ID:        cluster.ID,
		Name:      cluster.Name,
		Context:   cluster.Context,
		Server:    cluster.Server,
		Status:    cluster.Status,
		LastError: cluster.LastError,
		IsPinned:  cluster.IsPinned,
	}
	cs.eventEmitter.Emit("cluster:updated", clusterInfo)

	return nil
}

// AddResourceWatcher adds a resource watcher for a cluster
func (cs *ClusterService) AddResourceWatcher(request ResourceWatchRequest) error {
	gvr := schema.GroupVersionResource{
		Group:    request.Group,
		Version:  request.Version,
		Resource: request.Resource,
	}

	err := cs.informerManager.AddResourceWatcher(request.ClusterID, gvr, request.Namespace)
	if err != nil {
		return err
	}

	// Emit watcher added event
	cs.eventEmitter.Emit("watcher:added", request)

	return nil
}

// RemoveResourceWatcher removes a resource watcher
func (cs *ClusterService) RemoveResourceWatcher(request ResourceWatchRequest) error {
	gvr := schema.GroupVersionResource{
		Group:    request.Group,
		Version:  request.Version,
		Resource: request.Resource,
	}

	err := cs.informerManager.RemoveResourceWatcher(request.ClusterID, gvr)
	if err != nil {
		return err
	}

	// Emit watcher removed event
	cs.eventEmitter.Emit("watcher:removed", request)

	return nil
}

// LoadKubeconfigFromFile loads kubeconfig from file path
func (cs *ClusterService) LoadKubeconfigFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read kubeconfig file: %w", err)
	}

	return string(data), nil
}

// SaveKubeconfigToFile saves kubeconfig content to file
func (cs *ClusterService) SaveKubeconfigToFile(content, fileName string) (string, error) {
	kubeconfigDir := filepath.Join(cs.dataDir, "kubeconfigs")
	os.MkdirAll(kubeconfigDir, 0755)

	filePath := filepath.Join(kubeconfigDir, fileName)
	err := os.WriteFile(filePath, []byte(content), 0600)
	if err != nil {
		return "", fmt.Errorf("failed to save kubeconfig: %w", err)
	}

	return filePath, nil
}

// GetKubeconfigFiles returns list of saved kubeconfig files
func (cs *ClusterService) GetKubeconfigFiles() ([]string, error) {
	kubeconfigDir := filepath.Join(cs.dataDir, "kubeconfigs")
	
	entries, err := os.ReadDir(kubeconfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// WatchDefaultKubeconfig watches the default ~/.kube directory for changes
func (cs *ClusterService) WatchDefaultKubeconfig() error {
	// This would implement file system watching for ~/.kube directory
	// For now, we'll just return nil as this requires additional dependencies
	// like fsnotify which should be added to go.mod
	return nil
}

// GetResourceTypes returns available resource types for a cluster
func (cs *ClusterService) GetResourceTypes(clusterID string) ([]schema.GroupVersionResource, error) {
	clusters := cs.informerManager.GetClusters()
	cluster, exists := clusters[clusterID]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", clusterID)
	}

	// Create discovery client from config
	// This is a simplified implementation - in practice, you'd want to
	// cache this information and refresh it periodically
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cluster.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}
	
	resourceLists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("failed to discover resources: %w", err)
	}

	var gvrs []schema.GroupVersionResource
	for _, resourceList := range resourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			continue
		}

		for _, resource := range resourceList.APIResources {
			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: resource.Name,
			}
			gvrs = append(gvrs, gvr)
		}
	}

	return gvrs, nil
}

// Shutdown gracefully shuts down the service
func (cs *ClusterService) Shutdown() {
	cs.informerManager.Shutdown()
}
