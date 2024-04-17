package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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

func GetEndpoints(url string) (Endpoints, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Endpoints{}, fmt.Errorf("failed to get endpoints: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Endpoints{}, fmt.Errorf("failed to read response body: %v", err)
	}

	var endpoints Endpoints
	err = json.Unmarshal(body, &endpoints)
	if err != nil {
		return Endpoints{}, fmt.Errorf("failed to unmarshal endpoints: %v", err)
	}

	return endpoints, nil
}

func NetworkTopoHandler(w http.ResponseWriter, r *http.Request, topology NetworkTopology) {

	jsonData, err := json.Marshal(topology)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
