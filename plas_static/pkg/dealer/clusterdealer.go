package dealer

import (
	"context"
	"os"
	"path/filepath"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type NodeInfo struct {
	Name string
	IP   string
}

var PlacementPods map[string][]string // Node name to pod names

func GetKubernetesClient() (*kubernetes.Clientset, error) {
	// Load kubeconfig from the default location or use an in-cluster config
	// Modify if needed when the config file is not at the default path
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// Create the Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// Get all the nodes name and its IP on the cluster
func GetNodes(clientset *kubernetes.Clientset) ([]NodeInfo, error) {
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var nodeInfoList []NodeInfo
	for _, node := range nodes.Items {
		var nodeIP string
		var nodeName string

		for _, address := range node.Status.Addresses {
			if address.Type == "InternalIP" {
				nodeIP = address.Address
			} else if address.Type == "Hostname" {
				nodeName = address.Address
			}
		}

		if nodeName != "" {
			nodeInfoList = append(nodeInfoList, NodeInfo{Name: nodeName, IP: nodeIP})
		}
	}
	return nodeInfoList, nil
}

// Get the node's IP address by its name
func GetNodeIP(clientset *kubernetes.Clientset, nodeName string) (string, error) {
	node, err := clientset.CoreV1().Nodes().Get(context.Background(), nodeName, v1.GetOptions{})
	if err != nil {
		return "", err
	}

	for _, address := range node.Status.Addresses {
		if address.Type == "InternalIP" {
			return address.Address, nil
		}
	}
	return "", nil
}

func GetDeploymentPlacementInfo(clientset *kubernetes.Clientset, deploymentName, namespace string) (map[string][]string, error) {
	placementPods := make(map[string][]string)

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), deploymentName, v1.GetOptions{})
	if err != nil {
		return placementPods, err
	}

	podList, err := clientset.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{
		LabelSelector: v1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		return placementPods, err
	}

	for _, pod := range podList.Items {
		placementPods[pod.Spec.NodeName] = append(placementPods[pod.Spec.NodeName], pod.Name)
	}

	return placementPods, nil
}

func CountPodsOnNodes(PlacementPods map[string][]string) map[string]int {
	podCounts := make(map[string]int)

	for nodeName, pods := range PlacementPods {
		podCounts[nodeName] = len(pods)
	}

	return podCounts
}
