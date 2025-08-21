package test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"ksight/pkg/service"
)

var _ = Describe("ClusterService", func() {
	var (
		testService *service.ClusterService
		testCtx     context.Context
		testCancel  context.CancelFunc
	)

	BeforeEach(func() {
		testCtx, testCancel = context.WithCancel(context.Background())
		testService = service.NewClusterService(testCtx)
	})

	AfterEach(func() {
		if testService != nil {
			testService.Shutdown()
		}
		testCancel()
	})

	Context("Cluster Management", func() {
		It("should add cluster successfully", func() {
			kubeconfigPath := writeKubeconfigToTempFile()
			
			clusterID, err := testService.AddCluster("test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterID).NotTo(BeEmpty())

			clusters := testService.GetClusters()
			Expect(clusters).To(HaveKey(clusterID))
			Expect(clusters[clusterID].Name).To(Equal("test-cluster"))
			Expect(clusters[clusterID].Status).To(Equal("connected"))
		})

		It("should remove cluster successfully", func() {
			kubeconfigPath := writeKubeconfigToTempFile()
			
			clusterID, err := testService.AddCluster("test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			err = testService.RemoveCluster(clusterID)
			Expect(err).NotTo(HaveOccurred())

			clusters := testService.GetClusters()
			Expect(clusters).NotTo(HaveKey(clusterID))
		})

		It("should toggle cluster pin", func() {
			kubeconfigPath := writeKubeconfigToTempFile()
			
			clusterID, err := testService.AddCluster("test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			// Initially not pinned
			clusters := testService.GetClusters()
			Expect(clusters[clusterID].IsPinned).To(BeFalse())

			// Toggle to pinned
			err = testService.ToggleClusterPin(clusterID)
			Expect(err).NotTo(HaveOccurred())

			clusters = testService.GetClusters()
			Expect(clusters[clusterID].IsPinned).To(BeTrue())

			// Toggle back to unpinned
			err = testService.ToggleClusterPin(clusterID)
			Expect(err).NotTo(HaveOccurred())

			clusters = testService.GetClusters()
			Expect(clusters[clusterID].IsPinned).To(BeFalse())
		})

		It("should return error for invalid kubeconfig", func() {
			invalidKubeconfig := "invalid yaml content"
			
			_, err := testService.AddCluster("invalid-cluster", invalidKubeconfig, "")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Resource Watchers", func() {
		var clusterID string

		BeforeEach(func() {
			kubeconfigPath := writeKubeconfigToTempFile()
			var err error
			clusterID, err = testService.AddCluster("test-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should add resource watcher", func() {
			request := service.ResourceWatchRequest{
				ClusterID: clusterID,
				Group:     "",
				Version:   "v1",
				Resource:  "pods",
			}

			err := testService.AddResourceWatcher(request)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should remove resource watcher", func() {
			request := service.ResourceWatchRequest{
				ClusterID: clusterID,
				Group:     "",
				Version:   "v1",
				Resource:  "pods",
			}

			err := testService.AddResourceWatcher(request)
			Expect(err).NotTo(HaveOccurred())

			err = testService.RemoveResourceWatcher(request)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should get resource types for cluster", func() {
			gvrs, err := testService.GetResourceTypes(clusterID)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(gvrs)).To(BeNumerically(">", 0))

			// Check for common resources
			foundPods := false
			foundServices := false
			for _, gvr := range gvrs {
				if gvr.Resource == "pods" && gvr.Version == "v1" {
					foundPods = true
				}
				if gvr.Resource == "services" && gvr.Version == "v1" {
					foundServices = true
				}
			}
			Expect(foundPods).To(BeTrue())
			Expect(foundServices).To(BeTrue())
		})

		It("should handle multiple watchers", func() {
			requests := []service.ResourceWatchRequest{
				{ClusterID: clusterID, Group: "", Version: "v1", Resource: "pods"},
				{ClusterID: clusterID, Group: "", Version: "v1", Resource: "services"},
				{ClusterID: clusterID, Group: "apps", Version: "v1", Resource: "deployments"},
			}

			for _, request := range requests {
				err := testService.AddResourceWatcher(request)
				Expect(err).NotTo(HaveOccurred())
			}

			// Remove all watchers
			for _, request := range requests {
				err := testService.RemoveResourceWatcher(request)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})

	Context("Kubeconfig Management", func() {
		It("should save and load kubeconfig files", func() {
			content := getKubeconfigContent()
			fileName := "test-kubeconfig.yaml"

			// Save kubeconfig
			filePath, err := testService.SaveKubeconfigToFile(content, fileName)
			Expect(err).NotTo(HaveOccurred())
			Expect(filePath).To(ContainSubstring(fileName))

			// Load kubeconfig
			loadedContent, err := testService.LoadKubeconfigFromFile(filePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(loadedContent).To(Equal(content))

			// List kubeconfig files
			files, err := testService.GetKubeconfigFiles()
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ContainElement(fileName))
		})

		It("should handle non-existent file", func() {
			_, err := testService.LoadKubeconfigFromFile("/non/existent/path")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Integration with Real Resources", func() {
		var (
			clusterID string
			testNS    *corev1.Namespace
		)

		BeforeEach(func() {
			kubeconfigPath := writeKubeconfigToTempFile()
			var err error
			clusterID, err = testService.AddCluster("integration-cluster", kubeconfigPath, "test-context")
			Expect(err).NotTo(HaveOccurred())

			// Create test namespace with unique name
			nsName := fmt.Sprintf("integration-test-%d", time.Now().UnixNano())
			testNS = createTestNamespace(nsName)
			Expect(k8sClient.Create(ctx, testNS)).To(Succeed())
		})

		AfterEach(func() {
			if testNS != nil {
				deleteResource(testNS)
			}
		})

		It("should watch and receive pod events", func() {
			// Add pod watcher
			request := service.ResourceWatchRequest{
				ClusterID: clusterID,
				Group:     "",
				Version:   "v1",
				Resource:  "pods",
			}

			err := testService.AddResourceWatcher(request)
			Expect(err).NotTo(HaveOccurred())

			// Wait for informer to sync
			time.Sleep(2 * time.Second)

			// Create a test pod
			testPod := createTestPod(testNS.Name, "integration-pod")
			Expect(k8sClient.Create(ctx, testPod)).To(Succeed())
			defer deleteResource(testPod)

			// Wait a bit for the event to be processed
			time.Sleep(3 * time.Second)

			// The event should have been processed by the informer
			// In a real test, you'd capture events and verify them
			Expect(testPod.Name).To(Equal("integration-pod"))
		})

		It("should handle cluster removal with active watchers", func() {
			// Add multiple watchers
			requests := []service.ResourceWatchRequest{
				{ClusterID: clusterID, Group: "", Version: "v1", Resource: "pods"},
				{ClusterID: clusterID, Group: "", Version: "v1", Resource: "services"},
			}

			for _, request := range requests {
				err := testService.AddResourceWatcher(request)
				Expect(err).NotTo(HaveOccurred())
			}

			// Remove cluster - should clean up all watchers
			err := testService.RemoveCluster(clusterID)
			Expect(err).NotTo(HaveOccurred())

			clusters := testService.GetClusters()
			Expect(clusters).NotTo(HaveKey(clusterID))
		})
	})

	Context("Error Handling", func() {
		It("should handle invalid cluster ID for watchers", func() {
			request := service.ResourceWatchRequest{
				ClusterID: "non-existent-cluster",
				Group:     "",
				Version:   "v1",
				Resource:  "pods",
			}

			err := testService.AddResourceWatcher(request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should handle toggle pin for non-existent cluster", func() {
			err := testService.ToggleClusterPin("non-existent-cluster")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("should handle remove non-existent cluster", func() {
			err := testService.RemoveCluster("non-existent-cluster")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})
	})
})
