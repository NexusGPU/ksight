package allocation

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func Detect() {
	var kubeconfig *string

	if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
		kubeconfig = flag.String("kubeconfig", kubeconfigEnv, "absolute path to the kubeconfig file")
	} else if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating clientset: %v\n", err)
		os.Exit(1)
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing pods: %v\n", err)
		os.Exit(1)
	}

	totalGPUs := int64(0)
	foundGPUPods := false

	fmt.Println("Pod Name | Node Name | GPU Request Count")
	fmt.Println("---------|-----------|------------------")

	for _, pod := range pods.Items {
		gpuCount := int64(0)
		hasGPU := false
		for _, container := range pod.Spec.Containers {
			if limits := container.Resources.Limits; limits != nil {
				if gpuQuantity, exists := limits["nvidia.com/gpu"]; exists {
					hasGPU = true
					gpuValue := gpuQuantity.Value()
					gpuCount += gpuValue
				}
			}
		}

		if hasGPU {
			foundGPUPods = true
			requestGPUCount := int64(0)
			for _, container := range pod.Spec.Containers {
				if requests := container.Resources.Requests; requests != nil {
					if gpuQuantity, exists := requests["nvidia.com/gpu"]; exists {
						requestGPUCount += gpuQuantity.Value()
					}
				}
			}

			fmt.Printf("%s | %s | %d\n", pod.Name, pod.Spec.NodeName, requestGPUCount)
			totalGPUs += gpuCount
		}
	}

	if !foundGPUPods {
		fmt.Println("No pods with GPU limits found.")
	}

	fmt.Println("---------|-----------|------------------")
	fmt.Printf("Total GPU limits: %d\n", totalGPUs)
}
