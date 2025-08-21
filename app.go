package main

import (
	"context"
	"fmt"

	"ksight/pkg/service"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// App struct
type App struct {
	ctx            context.Context
	clusterService *service.ClusterService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.clusterService = service.NewClusterService(ctx)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Cluster Management Methods

// AddCluster adds a new cluster connection
func (a *App) AddCluster(name, kubeconfig, context string) (string, error) {
	return a.clusterService.AddCluster(name, kubeconfig, context)
}

// RemoveCluster removes a cluster connection
func (a *App) RemoveCluster(clusterID string) error {
	return a.clusterService.RemoveCluster(clusterID)
}

// GetClusters returns all cluster connections
func (a *App) GetClusters() map[string]service.ClusterInfo {
	return a.clusterService.GetClusters()
}

// ToggleClusterPin toggles the pinned state of a cluster
func (a *App) ToggleClusterPin(clusterID string) error {
	return a.clusterService.ToggleClusterPin(clusterID)
}

// Resource Watcher Methods

// AddResourceWatcher adds a resource watcher for a cluster
func (a *App) AddResourceWatcher(clusterID, group, version, resource, namespace string) error {
	request := service.ResourceWatchRequest{
		ClusterID: clusterID,
		Group:     group,
		Version:   version,
		Resource:  resource,
		Namespace: namespace,
	}
	return a.clusterService.AddResourceWatcher(request)
}

// RemoveResourceWatcher removes a resource watcher
func (a *App) RemoveResourceWatcher(clusterID, group, version, resource string) error {
	request := service.ResourceWatchRequest{
		ClusterID: clusterID,
		Group:     group,
		Version:   version,
		Resource:  resource,
	}
	return a.clusterService.RemoveResourceWatcher(request)
}

// GetResourceTypes returns available resource types for a cluster
func (a *App) GetResourceTypes(clusterID string) ([]schema.GroupVersionResource, error) {
	return a.clusterService.GetResourceTypes(clusterID)
}

// Kubeconfig Management Methods

// LoadKubeconfigFromFile loads kubeconfig from file path
func (a *App) LoadKubeconfigFromFile(filePath string) (string, error) {
	return a.clusterService.LoadKubeconfigFromFile(filePath)
}

// SaveKubeconfigToFile saves kubeconfig content to file
func (a *App) SaveKubeconfigToFile(content, fileName string) (string, error) {
	return a.clusterService.SaveKubeconfigToFile(content, fileName)
}

// GetKubeconfigFiles returns list of saved kubeconfig files
func (a *App) GetKubeconfigFiles() ([]string, error) {
	return a.clusterService.GetKubeconfigFiles()
}

// WatchDefaultKubeconfig watches the default ~/.kube directory for changes
func (a *App) WatchDefaultKubeconfig() error {
	return a.clusterService.WatchDefaultKubeconfig()
}

// Shutdown gracefully shuts down the app
func (a *App) Shutdown() {
	if a.clusterService != nil {
		a.clusterService.Shutdown()
	}
}
