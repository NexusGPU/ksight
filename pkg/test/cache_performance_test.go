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

func BenchmarkSQLitePerformance(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "ksight-perf-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cacheDir := filepath.Join(tempDir, "cache")
	os.MkdirAll(cacheDir, 0755)

	cache, err := informer.NewDatabaseCache(cacheDir, informer.DefaultSensitiveConfig)
	if err != nil {
		b.Fatalf("Failed to create database cache: %v", err)
	}
	defer cache.Close()

	b.Run("LoadPerformanceTest", func(b *testing.B) {
		testLargeDatasetPerformance(b, cache)
	})

	b.Run("SingleWrites", func(b *testing.B) {
		testSingleWrites(b, cache)
	})
}

func testLargeDatasetPerformance(b *testing.B, cache *informer.DatabaseCache) {
	clusterID := "perf-cluster"

	// Test with the specified counts
	counts := []struct {
		name    string
		count   int
		gvr     schema.GroupVersionResource
		genFunc func(int) *unstructured.Unstructured
	}{
		{"Pods", 100000, schema.GroupVersionResource{Version: "v1", Resource: "pods"}, generateSimplePod},
		{"Nodes", 10000, schema.GroupVersionResource{Version: "v1", Resource: "nodes"}, generateSimpleNode},
		{"Services", 10000, schema.GroupVersionResource{Version: "v1", Resource: "services"}, generateSimpleService},
		{"ConfigMaps", 1000, schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}, generateSimpleConfigMap},
		{"Deployments", 1000, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, generateSimpleDeployment},
	}

	totalResources := 0
	for _, c := range counts {
		totalResources += c.count
	}

	b.Logf("=== PERFORMANCE TEST: %d Total Resources ===", totalResources)

	overallStart := time.Now()

	for _, c := range counts {
		b.Logf("Testing %s (%d resources)...", c.name, c.count)
		start := time.Now()

		for i := 0; i < c.count; i++ {
			resource := c.genFunc(i)
			cache.StoreResource(clusterID, c.gvr, resource)
		}

		duration := time.Since(start)
		avgMicros := float64(duration.Microseconds()) / float64(c.count)

		b.Logf("%s: %v total, %.2f μs/resource, %.0f resources/sec",
			c.name, duration, avgMicros, float64(c.count)/duration.Seconds())
	}

	totalDuration := time.Since(overallStart)
	totalAvgMicros := float64(totalDuration.Microseconds()) / float64(totalResources)

	b.Logf("=== OVERALL RESULTS ===")
	b.Logf("Total time: %v", totalDuration)
	b.Logf("Average per resource: %.2f μs", totalAvgMicros)
	b.Logf("Resources per second: %.0f", float64(totalResources)/totalDuration.Seconds())

	// Get final cache stats
	stats, _ := cache.GetCacheStats()
	b.Logf("=== CACHE STATS ===")
	b.Logf("Total cached: %d", stats["total"])
	b.Logf("Sensitive: %d", stats["sensitive"])
}

func testSingleWrites(b *testing.B, cache *informer.DatabaseCache) {
	clusterID := "single-write-cluster"
	podGVR := schema.GroupVersionResource{Version: "v1", Resource: "pods"}

	// Test single pod write performance
	b.Run("PodWrite", func(b *testing.B) {
		b.ResetTimer()
		var totalDuration time.Duration

		for i := 0; i < b.N; i++ {
			pod := generateSimplePod(i)
			start := time.Now()
			cache.StoreResource(clusterID, podGVR, pod)
			duration := time.Since(start)
			totalDuration += duration
		}

		avgDuration := totalDuration / time.Duration(b.N)
		b.Logf("Average single write: %v", avgDuration)
	})

	// Test sensitive resource write (Secret)
	secretGVR := schema.GroupVersionResource{Version: "v1", Resource: "secrets"}
	b.Run("SecretWrite", func(b *testing.B) {
		b.ResetTimer()
		var totalDuration time.Duration

		for i := 0; i < b.N; i++ {
			secret := generateSimpleSecret(i)
			start := time.Now()
			cache.StoreResource(clusterID, secretGVR, secret)
			duration := time.Since(start)
			totalDuration += duration
		}

		avgDuration := totalDuration / time.Duration(b.N)
		b.Logf("Average secret write (with redaction): %v", avgDuration)
	})
}

// Simple resource generators that avoid deep copy issues
func generateSimplePod(i int) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]any{
				"name":            fmt.Sprintf("pod-%d", i),
				"namespace":       "default",
				"resourceVersion": fmt.Sprintf("%d", time.Now().UnixNano()),
				"labels": map[string]any{
					"app": "benchmark",
				},
			},
			"spec": map[string]any{
				"containers": []any{
					map[string]any{
						"name":  "app",
						"image": "nginx:latest",
					},
				},
			},
			"status": map[string]any{
				"phase": "Running",
			},
		},
	}
}

func generateSimpleNode(i int) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Node",
			"metadata": map[string]any{
				"name":            fmt.Sprintf("node-%d", i),
				"resourceVersion": fmt.Sprintf("%d", time.Now().UnixNano()),
			},
			"status": map[string]any{
				"conditions": []any{
					map[string]any{
						"type":   "Ready",
						"status": "True",
					},
				},
			},
		},
	}
}

func generateSimpleService(i int) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]any{
				"name":            fmt.Sprintf("service-%d", i),
				"namespace":       "default",
				"resourceVersion": fmt.Sprintf("%d", time.Now().UnixNano()),
			},
			"spec": map[string]any{
				"selector": map[string]any{
					"app": fmt.Sprintf("app-%d", i),
				},
				"type": "ClusterIP",
			},
		},
	}
}

func generateSimpleConfigMap(i int) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]any{
				"name":            fmt.Sprintf("configmap-%d", i),
				"namespace":       "default",
				"resourceVersion": fmt.Sprintf("%d", time.Now().UnixNano()),
			},
			"data": map[string]any{
				"config": "key=value",
			},
		},
	}
}

func generateSimpleDeployment(i int) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"name":            fmt.Sprintf("deployment-%d", i),
				"namespace":       "default",
				"resourceVersion": fmt.Sprintf("%d", time.Now().UnixNano()),
			},
			"spec": map[string]any{
				"replicas": "3",
				"selector": map[string]any{
					"matchLabels": map[string]any{
						"app": fmt.Sprintf("app-%d", i),
					},
				},
			},
		},
	}
}

func generateSimpleSecret(i int) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]any{
				"name":            fmt.Sprintf("secret-%d", i),
				"namespace":       "default",
				"resourceVersion": fmt.Sprintf("%d", time.Now().UnixNano()),
			},
			"type": "Opaque",
			"data": map[string]any{
				"password": "c2VjcmV0", // base64 encoded "secret"
				"username": "YWRtaW4=", // base64 encoded "admin"
			},
		},
	}
}
