package test

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"ksight/pkg/informer"
	"ksight/pkg/service"
)

var (
	cfg       *rest.Config
	k8sClient client.Client
	clientset *kubernetes.Clientset
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc

	// Test instances
	informerManager *informer.InformerManager
	clusterService  *service.ClusterService
	tempDir         string
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	if os.Getenv("DEBUG_MODE") == "true" {
		SetDefaultEventuallyTimeout(10 * time.Minute)
	} else {
		SetDefaultEventuallyTimeout(10 * time.Second)
	}
	SetDefaultEventuallyPollingInterval(200 * time.Millisecond)
	SetDefaultConsistentlyDuration(5 * time.Second)
	SetDefaultConsistentlyPollingInterval(250 * time.Millisecond)

	RunSpecs(t, "KSight Backend Test Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{},
		ErrorIfCRDPathMissing: false,
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s"),
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	scheme := runtime.NewScheme()
	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = appsv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	clientset, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(clientset).NotTo(BeNil())

	// Create temporary directory for test data
	tempDir, err = os.MkdirTemp("", "ksight-test-*")
	Expect(err).NotTo(HaveOccurred())

	// Initialize test instances
	informerManager = informer.NewInformerManager(
		filepath.Join(tempDir, "resource_versions.json"),
		func(event informer.Event) {
			// Event handler for tests - we'll capture events in individual tests
		},
	)

	clusterService = service.NewClusterService(ctx)
})

var _ = AfterSuite(func() {
	cancel()

	if informerManager != nil {
		informerManager.Shutdown()
	}

	if clusterService != nil {
		clusterService.Shutdown()
	}

	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())

	// Clean up temp directory
	if tempDir != "" {
		os.RemoveAll(tempDir)
	}
})

// Helper functions for tests

func createTestPod(namespace, name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "test-app",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "nginx:latest",
				},
			},
		},
	}
}

func createTestDeployment(namespace, name string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "test-deployment",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-deployment",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test-deployment",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
}

func createTestNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func waitForResource(obj client.Object, timeout time.Duration) {
	Eventually(func() error {
		return k8sClient.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	}, timeout, time.Second).Should(Succeed())
}

func deleteResource(obj client.Object) {
	// Delete the resource
	err := k8sClient.Delete(ctx, obj)
	if client.IgnoreNotFound(err) != nil {
		Expect(err).NotTo(HaveOccurred())
	}
	
	// Wait for the resource to be fully deleted
	Eventually(func() bool {
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(obj), obj)
		return client.IgnoreNotFound(err) == nil
	}, 30*time.Second, time.Second).Should(BeTrue())
}

func getKubeconfigContent() string {
	// Generate kubeconfig content for the test cluster
	var userConfig string
	if cfg.BearerToken != "" {
		userConfig = fmt.Sprintf("token: %s", cfg.BearerToken)
	} else if len(cfg.CertData) > 0 && len(cfg.KeyData) > 0 {
		userConfig = fmt.Sprintf(`client-certificate-data: %s
    client-key-data: %s`,
			base64.StdEncoding.EncodeToString(cfg.CertData),
			base64.StdEncoding.EncodeToString(cfg.KeyData))
	} else {
		// Fallback to insecure for test environment
		userConfig = "token: test-token"
	}

	return fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: %s
    server: %s
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    %s
`,
		base64.StdEncoding.EncodeToString(cfg.CAData),
		cfg.Host,
		userConfig,
	)
}

func writeKubeconfigToTempFile() string {
	// Create a temporary kubeconfig file
	tmpFile, err := os.CreateTemp(tempDir, "kubeconfig-*.yaml")
	Expect(err).NotTo(HaveOccurred())
	defer tmpFile.Close()

	// Write kubeconfig content to the file
	kubeconfigContent := getKubeconfigContent()
	_, err = tmpFile.WriteString(kubeconfigContent)
	Expect(err).NotTo(HaveOccurred())

	return tmpFile.Name()
}
