package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	dl "plas_static/pkg/dealer"
	dt "plas_static/pkg/detector"
	ex "plas_static/pkg/executor"
	"plas_static/pkg/gurobi"
	lt "plas_static/pkg/latency"
	sn "plas_static/pkg/snapshotter"
	"strings"
	"sync"
	"time"
)

func main() {

	clientset, err := dl.GetKubernetesClient()
	if err != nil {
		fmt.Println("Error creating Kubernetes client:", err)
		return
	}

	config, err := dl.ReadConfig("config.yml")
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return
	}

	endpoints := dl.EndpointsParcer(*config)

	http.HandleFunc("/endpoints", func(w http.ResponseWriter, r *http.Request) {
		dl.EndpointsHoster(w, r, endpoints)
	})
	port := 5000 // Port, change if needed
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("HTTP server listening on%s\n", addr)

	// --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
	var mu sync.Mutex
	var SumNetworkTopology []lt.SumNetworkTopology

	go func() {
		for {
			time.Sleep(10 * time.Second)

			newSumNetworkTopology := lt.NetworkTopologyParser(endpoints)

			// for _, sumNetworkTopology := range newSumNetworkTopology {
			// 	for nodeInfo, ip := range sumNetworkTopology.NodeInfo {
			// 		fmt.Printf("For node: %s, IP: %s:\n", nodeInfo, ip)
			// 		for _, latencyNodesPair := range sumNetworkTopology.LatencyNodesPair {
			// 			fmt.Printf("Latency between %s and %s: %.2f ms\n", latencyNodesPair.Node1, latencyNodesPair.Node2, latencyNodesPair.Latency)
			// 		}
			// 		for _, latencyNodeUserPair := range sumNetworkTopology.LatencyNodeUserPair {
			// 			fmt.Printf("Latency between %s and user %s: %.2f ms\n", latencyNodeUserPair.Node, latencyNodeUserPair.User, latencyNodeUserPair.Latency)
			// 		}
			// 	}
			// }

			mu.Lock()
			SumNetworkTopology = newSumNetworkTopology
			mu.Unlock()

			time.Sleep(30 * time.Second)
		}
	}()

	go func() {

		for {

			// For each deployment in the config file
			for _, deploymentInConfig := range config.Deployments {
				deployment := deploymentInConfig

				go func() {

					mu.Lock()
					sumNetworkTopology := SumNetworkTopology
					mu.Unlock()

					NodeSnap := sn.Snapshotter(clientset, config, sumNetworkTopology, deployment.DeploymentName, deployment.Namespace, deployment.ServiceTime)

					sn.NodeSnapPrinter(NodeSnap, *config, deployment)

					// Convert the NodeSnap to JSON format
					jsonNodeSnap := gurobi.NodeSnapToJSON(NodeSnap, config)

					// Generate a JSON file from the JSONNodeSnap
					gurobi.GenerateJSONFile(jsonNodeSnap)

					// Compare the RTT results to the SLO, and print the results
					SLOViolationLogs, sumUserRTT, violation := dt.SLOViolationDetector(NodeSnap, deployment.SLO)

					vioRatio := float64(violation) / float64(sumUserRTT)

					// judge the violation ratio and decide whether to deploy the placement plan
					if vioRatio > 0 && vioRatio < deployment.SLOVioToleration {
						fmt.Printf("Warining:%d%% of user violate the SLO, violation rate: %.2f is within the toleration range %.2f, \n", violation, vioRatio*100, deployment.SLOVioToleration)
						fmt.Printf("SLO violation logs: %+v\n", SLOViolationLogs)
					} else if vioRatio > deployment.SLOVioToleration {
						fmt.Printf("SLO violation logs: %+v\n", SLOViolationLogs)
						fmt.Printf("SLO violation ratio: %.2f is violate, \n", vioRatio)
						// If the SLO is violated, generate a placement plan
						gurobi.RunGurobi()

						// Change here the format of the placement plan for ease of reading
						PlacementPlan := ex.JSONToPlacementPlan()

						// Print the placement plan
						ex.PlacementPlanPrinter(PlacementPlan)

						fmt.Print("Do you want to deploy the placement plan? (yes/no): ")
						reader := bufio.NewReader(os.Stdin)
						input, _ := reader.ReadString('\n')
						input = strings.TrimSpace(input)

						// Deploy the placement plan with the user's permission
						if input == "yes" {
							ex.ExecutePlacementPlan(clientset, PlacementPlan)
							fmt.Println("Placement plan is executed")
						} else {
							fmt.Println("Placement plan is not executed")
						}
					} else {
						fmt.Printf("SLO violation ratio: %.2f is within the toleration range %.2f, \n", vioRatio, deployment.SLOVioToleration)
					}

				}()
			}
			time.Sleep(60 * time.Second)
		}
	}()

	http.ListenAndServe(addr, nil)

}

// User connecting to the master node and get the node IP for him/her to connect to.

// 1 Step: Verify whether model is reflecting the reality correctly. Generate the placement plan and see whether the RTT for users are all belong the SLO,
// and whether the RTT is the same as what we calculated.
