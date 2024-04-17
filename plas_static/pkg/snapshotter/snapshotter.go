package snapshotter

import (
	"fmt"
	dl "plas_static/pkg/dealer"
	lt "plas_static/pkg/latency"
	prom "plas_static/pkg/metrics"

	"k8s.io/client-go/kubernetes"
)

// All time unit in the struct is in milliseconds
type NodeSnap struct {
	NodeName       string
	CPUCores       int
	RAMGB          float64
	TotalSlots     int // a slot is 1C1G
	UnusedCPUs     float64
	UnusedRAMGB    float64
	AvaiSlots      int
	Deployments    []NodeDeploymentSnap
	LatencyToUsers []LatencyToUsers
}

type LatencyToUsers struct {
	UserIP  string
	Latency float64
}

type NodeDeploymentSnap struct {
	DeploymentName string
	Pod            []PodSnap
	InstaneCount   int
	SumRequestRate float64
	ResponseTime   float64
	RTT            []RTTResult
}

type PodSnap struct {
	PodName        string
	AvgRequestRate float64
	Node           string
}

type RTTResult struct {
	UserIP string
	RTT    float64
}

// Snapper measures the CPU and RAM metrics of each node and the average request rate of each pod, and estimates the response time for the instances of a deployment
func Snapshotter(clientset *kubernetes.Clientset, config *dl.Config, SumNetworkTopology []lt.SumNetworkTopology, deploymentName, namespace string, serviceTime float64) []NodeSnap {
	prometheusURL_sys := "http://130.104.229.12:30090"
	prometheusURL_app := "http://130.104.229.12:30091"

	NodeMetricInfo, _ := prom.PromQueryNodeMetric(prometheusURL_sys)
	PlacementPods, _ := dl.GetDeploymentPlacementInfo(clientset, deploymentName, namespace)

	var Nodesnap []NodeSnap
	// For each node, get the CPU and RAM metrics
	for _, node := range NodeMetricInfo {

		var nodeSnap NodeSnap
		var nodeDeploymentSnap NodeDeploymentSnap
		nodeSnap.NodeName = node.Name
		nodeSnap.CPUCores = node.CPUCores
		nodeSnap.RAMGB = node.RAMGB
		nodeSnap.UnusedCPUs = node.UnusedCPUs
		nodeSnap.UnusedRAMGB = node.UnusedRAMGB

		// Calculate the total number of slots: 1C1G
		if int(node.CPUCores) <= int(node.RAMGB) {
			nodeSnap.TotalSlots = int(node.CPUCores)
		} else {
			nodeSnap.TotalSlots = int(node.RAMGB)
		}

		// Calculate the number of available slots: 1C1G
		if int(node.UnusedCPUs) <= int(node.UnusedRAMGB) {
			nodeSnap.AvaiSlots = int(node.UnusedCPUs)
		} else {
			nodeSnap.AvaiSlots = int(node.UnusedRAMGB)
		}

		// Measure the latency between the node and the users
		// nodeIP, _ := dl.GetNodeIP(clientset, node.Name)
		userIPs := config.GetUserIPs()
		for _, user := range userIPs {
			for _, networkTopology := range SumNetworkTopology {
				for _, latencyNodeUserPair := range networkTopology.LatencyNodeUserPair {
					if latencyNodeUserPair.Node == node.Name && latencyNodeUserPair.User == user {
						var latencyToUsers LatencyToUsers
						latencyToUsers.UserIP = user
						latencyToUsers.Latency = latencyNodeUserPair.Latency // Already in ms
						nodeSnap.LatencyToUsers = append(nodeSnap.LatencyToUsers, latencyToUsers)
					}
				}
			}

			// latency, _ := lt.MeasureLatency(user, nodeIP)
			// var latencyToUsers LatencyToUsers
			// latencyToUsers.UserIP = user
			// latencyToUsers.Latency = float64(latency) / 1e6 // in milliseconds
			// nodeSnap.LatencyToUsers = append(nodeSnap.LatencyToUsers, latencyToUsers)
		}

		// For each node, get the pods that belong to the deployment
		for nodeName, pods := range PlacementPods {
			//-----------------------------------
			// If the node name matches the node name in the placement info
			// if nodeName == node.Name {
			// 	var sumRequestRate float64
			// 	// For each pod, get the average request rate
			// 	for _, pod := range pods {
			// 		var podSnap PodSnap
			// 		podSnap.PodName = pod
			// 		avgRequestRate, _ := prom.PromQueryAvgRequestRate(prometheusURL_app, deploymentName)
			// 		fmt.Println("pod:", pod, "avgRequestRate:", avgRequestRate)
			// 		podSnap.AvgRequestRate = avgRequestRate

			// 		// Sum the request rates
			// 		sumRequestRate += avgRequestRate
			// 		podSnap.Node = nodeName
			// 		nodeDeploymentSnap.Pod = append(nodeDeploymentSnap.Pod, podSnap)
			// 	}
			// --------------Leave for later, openfaas doesn't support request rate for pod level

			// Store the sum of the request rates
			nodeDeploymentSnap.DeploymentName = deploymentName
			PodCounts := dl.CountPodsOnNodes(PlacementPods)
			// nodeDeploymentSnap.SumRequestRate = sumRequestRate
			nodeDeploymentSnap.InstaneCount = PodCounts[nodeName]

			funcRequestRate, _ := prom.PromQueryAvgRequestRate(prometheusURL_app, deploymentName)
			funcReplicasNum, _ := prom.PromQueryReplicas(prometheusURL_app, deploymentName)
			RequestRatePerPod := funcRequestRate / float64(funcReplicasNum)
			fmt.Printf("Func Request Rate: %.2f, Func Replicas: %d, Request Rate Per Pod: %.2f\n", funcRequestRate, funcReplicasNum, RequestRatePerPod)

			if nodeName == node.Name {
				var sumRequestRate float64
				// For each pod, get the average request rate
				for _, pod := range pods {
					var podSnap PodSnap
					podSnap.PodName = pod
					podSnap.AvgRequestRate = RequestRatePerPod

					podSnap.Node = nodeName
					nodeDeploymentSnap.Pod = append(nodeDeploymentSnap.Pod, podSnap)
				}

				nodeDeploymentSnap.SumRequestRate = float64(PodCounts[nodeName]) * RequestRatePerPod
				// Estimate the response time
				ResponseTime := ResponseTimeEstimator(serviceTime, PodCounts[nodeName], sumRequestRate)
				nodeDeploymentSnap.ResponseTime = ResponseTime // in ms

				//--------------------------------------------------------- Above is good

				// For each user, measure the latency between the user and the node
				// But only for the one deployed with the deployment, not for all the nodes, change it!!!
				for _, latency := range nodeSnap.LatencyToUsers {
					// Calculate the RTT
					RTT := RoundTripTimeEstimator(ResponseTime, latency.Latency)
					var rttResult RTTResult
					rttResult.UserIP = latency.UserIP
					rttResult.RTT = RTT // in milliseconds
					// Append the RTT result to the deployment snap
					nodeDeploymentSnap.RTT = append(nodeDeploymentSnap.RTT, rttResult)
				}

				// Append the deployment snap to the node snap
				nodeSnap.Deployments = append(nodeSnap.Deployments, nodeDeploymentSnap)
			}
		}
		Nodesnap = append(Nodesnap, nodeSnap)
	}
	return Nodesnap
}

// NodeSnapPrinter prints the node snap, for test purposes
func NodeSnapPrinter(NodeSnap []NodeSnap, config dl.Config, deployment dl.DeploymentConfig) {
	fmt.Println("---------------------------------------------")
	fmt.Printf("                Deployment:%s               \n", deployment.DeploymentName)
	fmt.Println("---------------------------------------------")
	for _, snaps := range NodeSnap {
		fmt.Printf("Node: %s, CPU cores: %d, RAM: %.2fGB, unused CPU: %.2f, unused RAM: %.2f GB, TotalSlot: %d, AvailableSlot: %d\n", snaps.NodeName, snaps.CPUCores, snaps.RAMGB, snaps.UnusedCPUs, snaps.UnusedRAMGB, snaps.TotalSlots, snaps.AvaiSlots)
		for _, latency := range snaps.LatencyToUsers {
			fmt.Printf("User: %s, Latency: %.2f ms\n", latency.UserIP, latency.Latency)
		}
		for _, deployment := range snaps.Deployments {
			fmt.Printf("# Pods: %d, Sum request rate: %.2f req/s, Response time: %.2f ms\n", deployment.InstaneCount, deployment.SumRequestRate, deployment.ResponseTime)
			for _, pod := range deployment.Pod {
				fmt.Printf("Pod: %s, Avg request rate: %.2f req/s\n", pod.PodName, pod.AvgRequestRate)
			}
			for _, rtt := range deployment.RTT {
				fmt.Printf("User: %s, RTT: %.2f ms\n", rtt.UserIP, rtt.RTT)
			}
		}
	}
}
