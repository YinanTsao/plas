package metrics

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type NodeMetricInfo struct {
	Name        string
	CPUCores    int
	RAMGB       float64
	UnusedCPUs  float64
	UnusedRAMGB float64
}

func PromQueryNodeMetric(prometheusURL string) ([]NodeMetricInfo, error) {

	client, err := api.NewClient(api.Config{Address: prometheusURL})
	if err != nil {
		log.Fatalf("Error creating Prometheus client: %v\n", err)
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cpuQuery := `count(node_cpu_seconds_total{mode="idle"}) by (instance)`
	memQuery := `node_memory_MemTotal_bytes / 1024 / 1024 / 1024`

	// cpuUsage := `100 * avg(1 - rate(node_cpu_seconds_total{mode="idle"}[1m])) by (instance)`

	usedCPUQuery := `rate(node_cpu_seconds_total{mode="idle"}[1m])`
	unusedRAMQuery := `node_memory_MemAvailable_bytes / 1024 / 1024 / 1024`

	cpuResult, _, err := v1api.Query(ctx, cpuQuery, time.Now())
	if err != nil {
		log.Fatalf("Error querying Prometheus for CPU cores: %v\n", err)
	}

	memResult, _, err := v1api.Query(ctx, memQuery, time.Now())
	if err != nil {
		log.Fatalf("Error querying Prometheus for RAM: %v\n", err)
	}

	usedCPUResult, _, err := v1api.Query(ctx, usedCPUQuery, time.Now())
	if err != nil {
		log.Fatalf("Error querying Prometheus for unused CPU: %v\n", err)
	}

	unusedRAMResult, _, err := v1api.Query(ctx, unusedRAMQuery, time.Now())
	if err != nil {
		log.Fatalf("Error querying Prometheus for unused RAM: %v\n", err)
	}

	// Create a map of node names to their metrics
	nodes := make(map[string]*NodeMetricInfo)

	for _, sample := range cpuResult.(model.Vector) {
		nodeName := string(sample.Metric["instance"])
		nodes[nodeName] = &NodeMetricInfo{Name: nodeName, CPUCores: int(sample.Value)}
	}

	for _, sample := range memResult.(model.Vector) {
		nodeName := string(sample.Metric["instance"])
		if node, ok := nodes[nodeName]; ok {
			node.RAMGB = float64(sample.Value)
		}
	}

	for _, sample := range usedCPUResult.(model.Vector) {
		nodeName := string(sample.Metric["instance"])
		if node, ok := nodes[nodeName]; ok {
			node.UnusedCPUs = float64(nodes[nodeName].CPUCores) - float64(sample.Value)
		}
	}

	for _, sample := range unusedRAMResult.(model.Vector) {
		nodeName := string(sample.Metric["instance"])
		if node, ok := nodes[nodeName]; ok {
			node.UnusedRAMGB = float64(sample.Value)
		}
	}

	// Convert map to slice
	nodeSlice := make([]NodeMetricInfo, 0, len(nodes))
	for _, node := range nodes {
		nodeSlice = append(nodeSlice, *node)
	}

	return nodeSlice, nil

}

func PromQueryAvgRequestRate(prometheusURL, deploymentName string) (float64, error) {

	client, err := api.NewClient(api.Config{Address: prometheusURL})
	if err != nil {
		log.Fatalf("Error creating Prometheus client: %v\n", err)
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//here we have to use the deployment name to get the request rate for one function in OpenFaaS since it doesn't
	//really offer the request rate for the pod level.
	//We will have to divide the request rate by the number of replicas to get the request rate for each pod,
	//Luckily, the gateway of OpenFaaS will distribute the requests evenly among the replicas using round-robin
	query := fmt.Sprintf(`sum (rate ( gateway_function_invocation_total{function_name='%s.openfaas-fn'} [1m]))`, deploymentName)
	// query := fmt.Sprintf(`sum (rate ( gateway_function_invocation_total{function_name='%s.openfaas-fn'} [1m]))`, podName)
	result, _, err := v1api.Query(ctx, query, time.Now())
	if err != nil {
		log.Fatalf("Error querying Prometheus for average request rate: %v\n", err)
	}

	var avgRequestRate float64
	if vec, ok := result.(model.Vector); ok && len(vec) > 0 {
		avgRequestRate = float64(vec[0].Value)
	}

	return avgRequestRate, nil
}

func PromQueryReplicas(prometheusURL, deploymentName string) (int64, error) {

	client, err := api.NewClient(api.Config{Address: prometheusURL})
	if err != nil {
		log.Fatalf("Error creating Prometheus client: %v\n", err)
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := fmt.Sprintf(`gateway_service_count{function_name="%s.openfaas-fn"}`, deploymentName)
	result, _, err := v1api.Query(ctx, query, time.Now())
	if err != nil {
		log.Fatalf("Error querying Prometheus for replicas: %v\n", err)
	}

	var replicas int64
	if vec, ok := result.(model.Vector); ok && len(vec) > 0 {
		replicas = int64(vec[0].Value)
	}

	return replicas, nil
}
