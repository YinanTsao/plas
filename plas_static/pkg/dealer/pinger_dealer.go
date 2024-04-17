package dealer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Endpoints struct {
	Node map[string]string
	User []string
}

type NetworkTopology struct {
	NodeInfo       map[string]string
	ToNodesLatency map[string]float64
	ToUserLatency  map[string]float64
}

func EndpointsParcer(config Config) Endpoints {
	clientset, err := GetKubernetesClient()
	if err != nil {
		fmt.Printf("Failed to get Kubernetes client: %v\n", err)
		return Endpoints{}
	}

	nodes, err := GetNodes(clientset)
	if err != nil {
		fmt.Printf("Failed to get nodes: %v\n", err)
		return Endpoints{}
	}

	endpoints := Endpoints{
		Node: make(map[string]string),
		User: []string{},
	}

	for _, node := range nodes {
		endpoints.Node[node.Name] = node.IP
	}

	endpoints.User = append(endpoints.User, config.Users...)

	return endpoints
}

func EndpointsHoster(w http.ResponseWriter, r *http.Request, endpoints Endpoints) {

	jsonData, err := json.Marshal(endpoints)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func NetworkTopologyGetter(url string) (NetworkTopology, error) {
	resp, err := http.Get(url)
	if err != nil {
		return NetworkTopology{}, fmt.Errorf("failed to get networktopology: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NetworkTopology{}, fmt.Errorf("failed to read response body: %v", err)
	}

	var topology NetworkTopology
	err = json.Unmarshal(body, &topology)
	if err != nil {
		return NetworkTopology{}, fmt.Errorf("failed to unmarshal networktopology: %v", err)
	}

	return topology, nil
}
