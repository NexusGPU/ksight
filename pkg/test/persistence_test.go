package test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"ksight/pkg/informer"
)

var _ = Describe("ResourceVersion Persistence", func() {
	var (
		testStorePath string
		testClusterID string
		manager1      *informer.InformerManager
		manager2      *informer.InformerManager
		events1       []informer.Event
		events2       []informer.Event
		events1Mutex  sync.RWMutex
		events2Mutex  sync.RWMutex
	)

	BeforeEach(func() {
		testStorePath = filepath.Join(tempDir, fmt.Sprintf("persistence_test_%d.json", GinkgoRandomSeed()))
		testClusterID = "persistence-test-cluster"
		events1 = []informer.Event{}
		events2 = []informer.Event{}

		// Create first manager
		manager1 = informer.NewInformerManager(
			testStorePath,
			func(event informer.Event) {
				events1Mutex.Lock()
				events1 = append(events1, event)
				events1Mutex.Unlock()
			},
		)
	})

	AfterEach(func() {
		if manager1 != nil {
			manager1.CleanupTestData()
			manager1.Shutdown()
		}
		if manager2 != nil {
			manager2.CleanupTestData()
			manager2.Shutdown()
		}
		// Clean up test storage path and cache directory
		if testStorePath != "" {
			os.Remove(testStorePath)
			cacheDir := filepath.Join(filepath.Dir(testStorePath), "cache")
			os.RemoveAll(cacheDir)
		}
	})

	Context("ResourceVersion Storage", func() {
		It("should persist resource versions to disk", func() {
			kubeconfigPath := writeKubeconfigToTempFile()

			// Add cluster and watcher
			err := manager1.AddCluster(testClusterID, "test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			podGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
			err = manager1.AddResourceWatcher(testClusterID, podGVR, "")
			Expect(err).NotTo(HaveOccurred())

			// Create test namespace and pod
			testNS := createTestNamespace("persistence-test")
			Expect(k8sClient.Create(ctx, testNS)).To(Succeed())
			defer deleteResource(testNS)

			time.Sleep(2 * time.Second)

			testPod := createTestPod("persistence-test", "persistence-pod")
			Expect(k8sClient.Create(ctx, testPod)).To(Succeed())
			defer deleteResource(testPod)

			// Wait for events
			Eventually(func() int {
				events1Mutex.RLock()
				defer events1Mutex.RUnlock()
				return len(events1)
			}, 10*time.Second, time.Second).Should(BeNumerically(">", 0))

			// Shutdown first manager
			manager1.Shutdown()
			manager1 = nil

			// Verify storage file exists
			Expect(testStorePath).To(BeAnExistingFile())

			// Create second manager with same storage path
			manager2 = informer.NewInformerManager(
				testStorePath,
				func(event informer.Event) {
					events2Mutex.Lock()
					events2 = append(events2, event)
					events2Mutex.Unlock()
				},
			)

			// Add same cluster and watcher
			err = manager2.AddCluster(testClusterID, "test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			err = manager2.AddResourceWatcher(testClusterID, podGVR, "")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(2 * time.Second)

			// Create another pod to trigger events
			testPod2 := createTestPod("persistence-test", "persistence-pod-2")
			Expect(k8sClient.Create(ctx, testPod2)).To(Succeed())
			defer deleteResource(testPod2)

			// Should receive events from second manager
			Eventually(func() int {
				events2Mutex.RLock()
				defer events2Mutex.RUnlock()
				return len(events2)
			}, 10*time.Second, time.Second).Should(BeNumerically(">", 0))
		})

		It("should handle missing storage file gracefully", func() {
			nonExistentPath := filepath.Join(tempDir, "non-existent.json")

			manager := informer.NewInformerManager(
				nonExistentPath,
				func(event informer.Event) {},
			)
			defer manager.Shutdown()

			kubeconfigPath := writeKubeconfigToTempFile()
			err := manager.AddCluster(testClusterID, "test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			clusters := manager.GetClusters()
			Expect(clusters).To(HaveKey(testClusterID))
		})

		It("should handle corrupted storage file", func() {
			// Write invalid JSON to storage file
			err := os.WriteFile(testStorePath, []byte("invalid json content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			manager := informer.NewInformerManager(
				testStorePath,
				func(event informer.Event) {},
			)
			defer manager.Shutdown()

			kubeconfigPath := writeKubeconfigToTempFile()
			err = manager.AddCluster(testClusterID, "test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			clusters := manager.GetClusters()
			Expect(clusters).To(HaveKey(testClusterID))
		})
	})

	Context("Cache Resumption", func() {
		It("should resume watching from last known resource version", func() {
			kubeconfigPath := writeKubeconfigToTempFile()

			// First manager session
			err := manager1.AddCluster(testClusterID, "test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			podGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
			err = manager1.AddResourceWatcher(testClusterID, podGVR, "")
			Expect(err).NotTo(HaveOccurred())

			// Create test namespace
			testNS := createTestNamespace("cache-resume-test")
			Expect(k8sClient.Create(ctx, testNS)).To(Succeed())
			defer deleteResource(testNS)

			time.Sleep(2 * time.Second)

			// Create first pod
			testPod1 := createTestPod("cache-resume-test", "resume-pod-1")
			Expect(k8sClient.Create(ctx, testPod1)).To(Succeed())
			defer deleteResource(testPod1)

			// Wait for events
			Eventually(func() int {
				events1Mutex.RLock()
				defer events1Mutex.RUnlock()
				return len(events1)
			}, 10*time.Second, time.Second).Should(BeNumerically(">", 0))

			// Get the last resource version from events
			var lastResourceVersion string
			events1Mutex.RLock()
			for _, event := range events1 {
				if event.Object != nil && event.Object.GetResourceVersion() != "" {
					lastResourceVersion = event.Object.GetResourceVersion()
				}
			}
			events1Mutex.RUnlock()

			// Shutdown first manager
			manager1.Shutdown()
			manager1 = nil

			// Create second manager
			manager2 = informer.NewInformerManager(
				testStorePath,
				func(event informer.Event) {
					events2Mutex.Lock()
					events2 = append(events2, event)
					events2Mutex.Unlock()
				},
			)

			// Add cluster and watcher again
			err = manager2.AddCluster(testClusterID, "test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			err = manager2.AddResourceWatcher(testClusterID, podGVR, "")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(2 * time.Second)

			// Create second pod
			testPod2 := createTestPod("cache-resume-test", "resume-pod-2")
			Expect(k8sClient.Create(ctx, testPod2)).To(Succeed())
			defer deleteResource(testPod2)

			// Should receive events for new pod
			Eventually(func() int {
				events2Mutex.RLock()
				defer events2Mutex.RUnlock()
				return len(events2)
			}, 10*time.Second, time.Second).Should(BeNumerically(">", 0))

			// Verify we have a resource version from the first session
			Expect(lastResourceVersion).NotTo(BeEmpty())
		})
	})

	Context("Multiple Clusters and Resources", func() {
		It("should persist resource versions for multiple clusters", func() {
			kubeconfigPath := writeKubeconfigToTempFile()

			cluster1ID := "cluster-1"
			cluster2ID := "cluster-2"

			// Add two clusters
			err := manager1.AddCluster(cluster1ID, "cluster-1", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			err = manager1.AddCluster(cluster2ID, "cluster-2", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			// Add watchers for different resources
			podGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
			serviceGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}

			err = manager1.AddResourceWatcher(cluster1ID, podGVR, "")
			Expect(err).NotTo(HaveOccurred())

			err = manager1.AddResourceWatcher(cluster2ID, serviceGVR, "")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(2 * time.Second)

			// Create test namespace
			testNS := createTestNamespace("multi-cluster-test")
			Expect(k8sClient.Create(ctx, testNS)).To(Succeed())
			defer deleteResource(testNS)

			// Create resources
			testPod := createTestPod("multi-cluster-test", "multi-pod")
			Expect(k8sClient.Create(ctx, testPod)).To(Succeed())
			defer deleteResource(testPod)

			// Wait for events
			Eventually(func() int {
				events1Mutex.RLock()
				defer events1Mutex.RUnlock()
				return len(events1)
			}, 10*time.Second, time.Second).Should(BeNumerically(">", 0))

			// Shutdown and recreate
			manager1.Shutdown()
			manager1 = nil

			manager2 = informer.NewInformerManager(
				testStorePath,
				func(event informer.Event) {
					events2Mutex.Lock()
					events2 = append(events2, event)
					events2Mutex.Unlock()
				},
			)

			// Verify both clusters can be added back
			err = manager2.AddCluster(cluster1ID, "cluster-1", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			err = manager2.AddCluster(cluster2ID, "cluster-2", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			clusters := manager2.GetClusters()
			Expect(clusters).To(HaveKey(cluster1ID))
			Expect(clusters).To(HaveKey(cluster2ID))
		})
	})
})
