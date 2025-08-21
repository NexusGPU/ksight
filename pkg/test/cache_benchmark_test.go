package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ksight/pkg/informer"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Resource counts for benchmark
const (
	POD_COUNT        = 100000
	NODE_COUNT       = 10000
	SERVICE_COUNT    = 10000
	CONFIGMAP_COUNT  = 1000
	DEPLOYMENT_COUNT = 1000
)

// BenchmarkDatabaseCache benchmarks the SQLite cache performance
func BenchmarkDatabaseCache(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "ksight-benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cacheDir := filepath.Join(tempDir, "cache")
	os.MkdirAll(cacheDir, 0755)

	// Create database cache
	cache, err := informer.NewDatabaseCache(cacheDir, informer.DefaultSensitiveConfig)
	if err != nil {
		b.Fatalf("Failed to create database cache: %v", err)
	}
	defer cache.Close()

	b.ResetTimer()
	b.Run("LoadInitialData", func(b *testing.B) {
		benchmarkLoadInitialData(b, cache)
	})

	b.Run("SingleEventWrites", func(b *testing.B) {
		benchmarkSingleEventWrites(b, cache)
	})

	b.Run("BatchEventWrites", func(b *testing.B) {
		benchmarkBatchEventWrites(b, cache)
	})

	b.Run("CacheQueries", func(b *testing.B) {
		benchmarkCacheQueries(b, cache)
	})
}

// benchmarkLoadInitialData tests loading large amounts of initial data
func benchmarkLoadInitialData(b *testing.B, cache *informer.DatabaseCache) {
	clusterID := "benchmark-cluster"
	
	// Generate test resources
	pods := generatePods(POD_COUNT)
	nodes := generateNodes(NODE_COUNT)
	services := generateServices(SERVICE_COUNT)
	configmaps := generateConfigMaps(CONFIGMAP_COUNT)
	deployments := generateDeployments(DEPLOYMENT_COUNT)

	totalResources := len(pods) + len(nodes) + len(services) + len(configmaps) + len(deployments)
	b.Logf("Loading %d total resources: %d pods, %d nodes, %d services, %d configmaps, %d deployments",
		totalResources, len(pods), len(nodes), len(services), len(configmaps), len(deployments))

	start := time.Now()

	// Store all resources
	for _, pod := range pods {
		cache.StoreResource(clusterID, schema.GroupVersionResource{Version: "v1", Resource: "pods"}, pod)
	}
	podsTime := time.Since(start)

	nodeStart := time.Now()
	for _, node := range nodes {
		cache.StoreResource(clusterID, schema.GroupVersionResource{Version: "v1", Resource: "nodes"}, node)
	}
	nodesTime := time.Since(nodeStart)

	serviceStart := time.Now()
	for _, service := range services {
		cache.StoreResource(clusterID, schema.GroupVersionResource{Version: "v1", Resource: "services"}, service)
	}
	servicesTime := time.Since(serviceStart)

	configmapStart := time.Now()
	for _, cm := range configmaps {
		cache.StoreResource(clusterID, schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}, cm)
	}
	configmapsTime := time.Since(configmapStart)

	deploymentStart := time.Now()
	for _, deploy := range deployments {
		cache.StoreResource(clusterID, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, deploy)
	}
	deploymentsTime := time.Since(deploymentStart)

	totalTime := time.Since(start)

	// Performance metrics
	b.Logf("=== LOAD PERFORMANCE METRICS ===")
	b.Logf("Pods (%d):        %v (%.2f μs/resource)", POD_COUNT, podsTime, float64(podsTime.Microseconds())/float64(POD_COUNT))
	b.Logf("Nodes (%d):       %v (%.2f μs/resource)", NODE_COUNT, nodesTime, float64(nodesTime.Microseconds())/float64(NODE_COUNT))
	b.Logf("Services (%d):    %v (%.2f μs/resource)", SERVICE_COUNT, servicesTime, float64(servicesTime.Microseconds())/float64(SERVICE_COUNT))
	b.Logf("ConfigMaps (%d):  %v (%.2f μs/resource)", CONFIGMAP_COUNT, configmapsTime, float64(configmapsTime.Microseconds())/float64(CONFIGMAP_COUNT))
	b.Logf("Deployments (%d): %v (%.2f μs/resource)", DEPLOYMENT_COUNT, deploymentsTime, float64(deploymentsTime.Microseconds())/float64(DEPLOYMENT_COUNT))
	b.Logf("TOTAL (%d):       %v (%.2f μs/resource)", totalResources, totalTime, float64(totalTime.Microseconds())/float64(totalResources))

	// Get cache stats
	stats, err := cache.GetCacheStats()
	if err == nil {
		b.Logf("=== CACHE STATISTICS ===")
		b.Logf("Total resources: %d", stats["total"])
		b.Logf("Sensitive resources: %d", stats["sensitive"])
	}
}

// benchmarkSingleEventWrites tests individual event write performance
func benchmarkSingleEventWrites(b *testing.B, cache *informer.DatabaseCache) {
	clusterID := "benchmark-cluster"
	podGVR := schema.GroupVersionResource{Version: "v1", Resource: "pods"}

	b.Run("PodEvents", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pod := generateSinglePod(fmt.Sprintf("benchmark-pod-%d", i))
			start := time.Now()
			cache.StoreResource(clusterID, podGVR, pod)
			duration := time.Since(start)
			if i == 0 {
				b.Logf("Single pod write: %v", duration)
			}
		}
	})

	b.Run("SensitiveResourceEvents", func(b *testing.B) {
		secretGVR := schema.GroupVersionResource{Version: "v1", Resource: "secrets"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			secret := generateSingleSecret(fmt.Sprintf("benchmark-secret-%d", i))
			start := time.Now()
			cache.StoreResource(clusterID, secretGVR, secret)
			duration := time.Since(start)
			if i == 0 {
				b.Logf("Single secret write (sensitive): %v", duration)
			}
		}
	})
}

// benchmarkBatchEventWrites tests batch write performance
func benchmarkBatchEventWrites(b *testing.B, cache *informer.DatabaseCache) {
	clusterID := "benchmark-cluster"
	
	b.Run("1000PodBatch", func(b *testing.B) {
		pods := generatePods(1000)
		podGVR := schema.GroupVersionResource{Version: "v1", Resource: "pods"}
		
		b.ResetTimer()
		start := time.Now()
		for _, pod := range pods {
			cache.StoreResource(clusterID, podGVR, pod)
		}
		duration := time.Since(start)
		b.Logf("1000 pod batch write: %v (%.2f μs/resource)", duration, float64(duration.Microseconds())/1000.0)
	})
}

// benchmarkCacheQueries tests query performance
func benchmarkCacheQueries(b *testing.B, cache *informer.DatabaseCache) {
	clusterID := "benchmark-cluster"
	podGVR := schema.GroupVersionResource{Version: "v1", Resource: "pods"}
	
	// Pre-populate with some data
	pods := generatePods(1000)
	for _, pod := range pods {
		cache.StoreResource(clusterID, podGVR, pod)
	}

	b.Run("GetResource", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			podName := fmt.Sprintf("test-pod-%d", i%1000)
			start := time.Now()
			_, _, err := cache.GetResource(clusterID, podGVR, "default", podName)
			duration := time.Since(start)
			if err == nil && i == 0 {
				b.Logf("Single resource query: %v", duration)
			}
		}
	})

	b.Run("LoadResources", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()
			resources, _, err := cache.LoadResources(clusterID, podGVR)
			duration := time.Since(start)
			if err == nil && i == 0 {
				b.Logf("Load all resources (%d): %v", len(resources), duration)
			}
		}
	})
}

// Helper functions to generate test resources

func generatePods(count int) []*unstructured.Unstructured {
	pods := make([]*unstructured.Unstructured, count)
	for i := 0; i < count; i++ {
		pods[i] = generateSinglePod(fmt.Sprintf("test-pod-%d", i))
	}
	return pods
}

func generateSinglePod(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":            name,
				"namespace":       "default",
				"resourceVersion": fmt.Sprintf("%d", time.Now().UnixNano()),
				"uid":             fmt.Sprintf("pod-uid-%s", name),
				"labels": map[string]interface{}{
					"app":     "benchmark-app",
					"version": "v1.0.0",
				},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "app-container",
						"image": "nginx:1.20",
						"ports": []interface{}{
							map[string]interface{}{
								"containerPort": "8080",
								"name":          "http",
							},
						},
						"env": []interface{}{
							map[string]interface{}{
								"name":  "ENV_VAR",
								"value": "test-value",
							},
						},
						"resources": map[string]interface{}{
							"requests": map[string]interface{}{
								"cpu":    "100m",
								"memory": "128Mi",
							},
							"limits": map[string]interface{}{
								"cpu":    "500m",
								"memory": "512Mi",
							},
						},
					},
				},
				"restartPolicy": "Always",
			},
			"status": map[string]interface{}{
				"phase": "Running",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
					},
				},
			},
		},
	}
}

func generateNodes(count int) []*unstructured.Unstructured {
	nodes := make([]*unstructured.Unstructured, count)
	for i := 0; i < count; i++ {
		nodes[i] = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Node",
				"metadata": map[string]interface{}{
					"name":            fmt.Sprintf("node-%d", i),
					"resourceVersion": fmt.Sprintf("%d", time.Now().UnixNano()),
					"labels": map[string]interface{}{
						"kubernetes.io/os":   "linux",
						"kubernetes.io/arch": "amd64",
						"node-type":          "worker",
					},
				},
				"spec": map[string]interface{}{
					"podCIDR": fmt.Sprintf("10.244.%d.0/24", i%256),
				},
				"status": map[string]interface{}{
					"capacity": map[string]interface{}{
						"cpu":    "4",
						"memory": "8Gi",
						"pods":   "110",
					},
					"allocatable": map[string]interface{}{
						"cpu":    "3800m",
						"memory": "7.5Gi",
						"pods":   "110",
					},
					"conditions": []interface{}{
						map[string]interface{}{
							"type":   "Ready",
							"status": "True",
						},
					},
				},
			},
		}
	}
	return nodes
}

func generateServices(count int) []*unstructured.Unstructured {
	services := make([]*unstructured.Unstructured, count)
	for i := 0; i < count; i++ {
		services[i] = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name":            fmt.Sprintf("service-%d", i),
					"namespace":       "default",
					"resourceVersion": fmt.Sprintf("%d", time.Now().UnixNano()),
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"app": fmt.Sprintf("app-%d", i),
					},
					"ports": []interface{}{
						map[string]interface{}{
							"name":       "http",
							"port":       "80",
							"targetPort": "8080",
							"protocol":   "TCP",
						},
					},
					"type": "ClusterIP",
				},
			},
		}
	}
	return services
}

func generateConfigMaps(count int) []*unstructured.Unstructured {
	configmaps := make([]*unstructured.Unstructured, count)
	for i := 0; i < count; i++ {
		configmaps[i] = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":            fmt.Sprintf("configmap-%d", i),
					"namespace":       "default",
					"resourceVersion": fmt.Sprintf("%d", time.Now().UnixNano()),
				},
				"data": map[string]interface{}{
					"app.properties": fmt.Sprintf("app.name=myapp-%d\napp.version=1.0.0\ndebug=false", i),
					"config.yaml":    "server:\n  port: 8080\n  host: 0.0.0.0",
					"script.sh":      "#!/bin/bash\necho 'Hello World'\nexit 0",
				},
			},
		}
	}
	return configmaps
}

func generateDeployments(count int) []*unstructured.Unstructured {
	deployments := make([]*unstructured.Unstructured, count)
	for i := 0; i < count; i++ {
		deployments[i] = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":            fmt.Sprintf("deployment-%d", i),
					"namespace":       "default",
					"resourceVersion": fmt.Sprintf("%d", time.Now().UnixNano()),
				},
				"spec": map[string]interface{}{
					"replicas": "3",
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app": fmt.Sprintf("app-%d", i),
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": fmt.Sprintf("app-%d", i),
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "app",
									"image": "nginx:1.20",
									"ports": []interface{}{
										map[string]interface{}{
											"containerPort": "80",
										},
									},
								},
							},
						},
					},
				},
				"status": map[string]interface{}{
					"replicas":      "3",
					"readyReplicas": "3",
					"conditions": []interface{}{
						map[string]interface{}{
							"type":   "Available",
							"status": "True",
						},
					},
				},
			},
		}
	}
	return deployments
}

func generateSingleSecret(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":            name,
				"namespace":       "default",
				"resourceVersion": fmt.Sprintf("%d", time.Now().UnixNano()),
			},
			"type": "Opaque",
			"data": map[string]interface{}{
				"username": "YWRtaW4=", // base64 encoded "admin"
				"password": "MWYyZDFlMmU2N2Rm", // base64 encoded "1f2d1e2e67df"
				"api-key":  "c2VjcmV0LWFwaS1rZXk=", // base64 encoded "secret-api-key"
			},
		},
	}
}