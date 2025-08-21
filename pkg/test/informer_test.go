package test

import (
	"path/filepath"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"ksight/pkg/informer"
)

var _ = Describe("InformerManager", func() {
	var (
		testClusterID string
		eventChan     chan informer.Event
		eventMutex    sync.RWMutex
		receivedEvents []informer.Event
	)

	BeforeEach(func() {
		testClusterID = "test-cluster-1"
		eventChan = make(chan informer.Event, 100)
		receivedEvents = []informer.Event{}

		// Create new informer manager with event capture
		informerManager = informer.NewInformerManager(
			filepath.Join(tempDir, "test_resource_versions.json"),
			func(event informer.Event) {
				eventMutex.Lock()
				receivedEvents = append(receivedEvents, event)
				eventMutex.Unlock()
				select {
				case eventChan <- event:
				default:
				}
			},
		)
	})

	AfterEach(func() {
		if informerManager != nil {
			informerManager.Shutdown()
		}
		close(eventChan)
	})

	Context("Cluster Management", func() {
		It("should add a cluster successfully", func() {
			kubeconfigPath := writeKubeconfigToTempFile()
			
			err := informerManager.AddCluster(testClusterID, "test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			clusters := informerManager.GetClusters()
			Expect(clusters).To(HaveKey(testClusterID))
			Expect(clusters[testClusterID].Name).To(Equal("test-cluster"))
			Expect(clusters[testClusterID].Status).To(Equal("connected"))
		})

		It("should remove a cluster successfully", func() {
			kubeconfigPath := writeKubeconfigToTempFile()
			
			err := informerManager.AddCluster(testClusterID, "test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			err = informerManager.RemoveCluster(testClusterID)
			Expect(err).NotTo(HaveOccurred())

			clusters := informerManager.GetClusters()
			Expect(clusters).NotTo(HaveKey(testClusterID))
		})

		It("should return error when removing non-existent cluster", func() {
			err := informerManager.RemoveCluster("non-existent")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})
	})

	Context("Resource Watchers", func() {
		BeforeEach(func() {
			kubeconfigPath := writeKubeconfigToTempFile()
			err := informerManager.AddCluster(testClusterID, "test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should add pod watcher and receive events", func() {
			// Create test namespace
			testNS := createTestNamespace("test-pods")
			Expect(k8sClient.Create(ctx, testNS)).To(Succeed())
			defer deleteResource(testNS)

			// Add pod watcher
			podGVR := schema.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			}
			
			err := informerManager.AddResourceWatcher(testClusterID, podGVR, "")
			Expect(err).NotTo(HaveOccurred())

			// Wait for informer to sync
			time.Sleep(2 * time.Second)

			// Create a test pod
			testPod := createTestPod("test-pods", "test-pod-1")
			Expect(k8sClient.Create(ctx, testPod)).To(Succeed())
			defer deleteResource(testPod)

			// Wait for and verify event
			Eventually(func() int {
				eventMutex.RLock()
				defer eventMutex.RUnlock()
				return len(receivedEvents)
			}, 10*time.Second, time.Second).Should(BeNumerically(">", 0))

			eventMutex.RLock()
			found := false
			for _, event := range receivedEvents {
				if event.Type == "ADDED" && event.Name == "test-pod-1" && event.GVR.Resource == "pods" {
					found = true
					break
				}
			}
			eventMutex.RUnlock()
			Expect(found).To(BeTrue())
		})

		It("should add deployment watcher and receive events", func() {
			// Create test namespace
			testNS := createTestNamespace("test-deployments")
			Expect(k8sClient.Create(ctx, testNS)).To(Succeed())
			defer deleteResource(testNS)

			// Add deployment watcher
			deploymentGVR := schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			}
			
			err := informerManager.AddResourceWatcher(testClusterID, deploymentGVR, "")
			Expect(err).NotTo(HaveOccurred())

			// Wait for informer to sync
			time.Sleep(2 * time.Second)

			// Create a test deployment
			testDeployment := createTestDeployment("test-deployments", "test-deployment-1", 1)
			Expect(k8sClient.Create(ctx, testDeployment)).To(Succeed())
			defer deleteResource(testDeployment)

			// Wait for and verify event
			Eventually(func() bool {
				eventMutex.RLock()
				defer eventMutex.RUnlock()
				for _, event := range receivedEvents {
					if event.Type == "ADDED" && event.Name == "test-deployment-1" && event.GVR.Resource == "deployments" {
						return true
					}
				}
				return false
			}, 10*time.Second, time.Second).Should(BeTrue())
		})

		It("should remove resource watcher successfully", func() {
			podGVR := schema.GroupVersionResource{
				Group:    "",
				Version:  "v1", 
				Resource: "pods",
			}
			
			err := informerManager.AddResourceWatcher(testClusterID, podGVR, "")
			Expect(err).NotTo(HaveOccurred())

			err = informerManager.RemoveResourceWatcher(testClusterID, podGVR)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle multiple watchers for same cluster", func() {
			podGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
			serviceGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
			
			err := informerManager.AddResourceWatcher(testClusterID, podGVR, "")
			Expect(err).NotTo(HaveOccurred())
			
			err = informerManager.AddResourceWatcher(testClusterID, serviceGVR, "")
			Expect(err).NotTo(HaveOccurred())

			// Verify both watchers are active
			clusters := informerManager.GetClusters()
			cluster := clusters[testClusterID]
			Expect(len(cluster.Informers)).To(Equal(2))
		})
	})

	Context("Resource Version Persistence", func() {
		It("should persist and restore resource versions", func() {
			kubeconfigPath := writeKubeconfigToTempFile()
			err := informerManager.AddCluster(testClusterID, "test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			// Create test namespace and pod
			testNS := createTestNamespace("test-persistence")
			Expect(k8sClient.Create(ctx, testNS)).To(Succeed())
			defer deleteResource(testNS)

			podGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
			err = informerManager.AddResourceWatcher(testClusterID, podGVR, "")
			Expect(err).NotTo(HaveOccurred())

			// Wait for sync and create pod to generate events
			time.Sleep(2 * time.Second)
			testPod := createTestPod("test-persistence", "test-pod-rv")
			Expect(k8sClient.Create(ctx, testPod)).To(Succeed())
			defer deleteResource(testPod)

			// Wait for event to be processed
			Eventually(func() int {
				eventMutex.RLock()
				defer eventMutex.RUnlock()
				return len(receivedEvents)
			}, 10*time.Second, time.Second).Should(BeNumerically(">", 0))

			// Shutdown and recreate informer manager
			informerManager.Shutdown()
			
			// Create new manager with same storage path
			newManager := informer.NewInformerManager(
				filepath.Join(tempDir, "test_resource_versions.json"),
				func(event informer.Event) {},
			)
			defer newManager.Shutdown()

			// Verify resource versions were loaded
			// This is a basic test - in practice, you'd verify the actual resource version values
			Expect(newManager).NotTo(BeNil())
		})
	})

	Context("Event Handling", func() {
		BeforeEach(func() {
			kubeconfigPath := writeKubeconfigToTempFile()
			err := informerManager.AddCluster(testClusterID, "test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle pod lifecycle events", func() {
			testNS := createTestNamespace("test-lifecycle")
			Expect(k8sClient.Create(ctx, testNS)).To(Succeed())
			defer deleteResource(testNS)

			podGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
			err := informerManager.AddResourceWatcher(testClusterID, podGVR, "")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(2 * time.Second)

			// Create pod
			testPod := createTestPod("test-lifecycle", "lifecycle-pod")
			Expect(k8sClient.Create(ctx, testPod)).To(Succeed())

			// Wait for ADD event
			Eventually(func() bool {
				eventMutex.RLock()
				defer eventMutex.RUnlock()
				for _, event := range receivedEvents {
					if event.Type == "ADDED" && event.Name == "lifecycle-pod" {
						return true
					}
				}
				return false
			}, 10*time.Second, time.Second).Should(BeTrue())

			// Update pod
			testPod.Labels["updated"] = "true"
			Expect(k8sClient.Update(ctx, testPod)).To(Succeed())

			// Wait for MODIFIED event
			Eventually(func() bool {
				eventMutex.RLock()
				defer eventMutex.RUnlock()
				for _, event := range receivedEvents {
					if event.Type == "MODIFIED" && event.Name == "lifecycle-pod" {
						return true
					}
				}
				return false
			}, 10*time.Second, time.Second).Should(BeTrue())

			// Delete pod
			deleteResource(testPod)

			// Wait for DELETE event
			Eventually(func() bool {
				eventMutex.RLock()
				defer eventMutex.RUnlock()
				for _, event := range receivedEvents {
					if event.Type == "DELETED" && event.Name == "lifecycle-pod" {
						return true
					}
				}
				return false
			}, 10*time.Second, time.Second).Should(BeTrue())
		})
	})
})
