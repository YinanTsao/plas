package executor

import (
	"context"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ExecutePlacementPlan(clientset *kubernetes.Clientset, plan PlacementPlan) {
	// Get the current deployment
	deployment, err := clientset.AppsV1().Deployments("default").Get(context.Background(), plan.DeploymentName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// Update the number of replicas for each node
	for nodeName, replicaCount := range plan.NodeReplicas {
		// Get the current node
		node, err := clientset.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
		if err != nil {
			log.Fatal(err)
		}

		// Update the node's labels to specify the number of replicas
		node.Labels["replicas"] = fmt.Sprintf("%d", replicaCount)

		// Update the node
		_, err = clientset.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}

	// Update the deployment's node selector to select nodes with the correct number of replicas
	deployment.Spec.Template.Spec.NodeSelector = map[string]string{
		"replicas": fmt.Sprintf("%d", len(plan.NodeReplicas)),
	}

	// Update the deployment
	_, err = clientset.AppsV1().Deployments("default").Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// Delete pods on other nodes
	pods, err := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, pod := range pods.Items {
		if _, ok := plan.NodeReplicas[pod.Spec.NodeName]; !ok {
			err := clientset.CoreV1().Pods("default").Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
