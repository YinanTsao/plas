package gurobi

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	dl "plas_static/pkg/dealer"
	sn "plas_static/pkg/snapshotter"
	"time"
)

type JSONNodeSnap struct {
	Nodes           []string           `json:"nodes"`
	Users           []string           `json:"users"`
	Capacities      map[string]int     `json:"capacities"`
	LatencyNodeUser map[string]float64 `json:"latency_node_user"`
	RPS             float64            `json:"rps"`
	ServiceRate     float64            `json:"service_rate"`
	SLO             float64            `json:"slo"`
	DeploymentName  string             `json:"deployment_name"`
	OptiPref        int                `json:"opti_pref"`
}

// !!!!!!!!!!FOR EACH DEPLOYMENT !!!!!!!!!!!
// The result will be going through all the Deployment in the cluster, as defined in main.go.
// For each deployment, the function will take a snapshot of the cluster, and then convert the snapshot to JSON format.

// converting the type NodeSnap to json format as data.json file
func NodeSnapToJSON(NodeSnap []sn.NodeSnap, config *dl.Config) JSONNodeSnap {
	var jsonNodeSnap JSONNodeSnap

	jsonNodeSnap.Capacities = make(map[string]int)
	jsonNodeSnap.LatencyNodeUser = make(map[string]float64)

	// for each NodeSnap, fill Nodes from NodeName
	for _, snap := range NodeSnap {

		// Initialize the map
		jsonNodeSnap.Nodes = append(jsonNodeSnap.Nodes, snap.NodeName)
		jsonNodeSnap.Users = config.Users
		jsonNodeSnap.Capacities[snap.NodeName] = snap.AvaiSlots

		// Initialize the inner map
		for _, latency := range snap.LatencyToUsers {
			// for each latency, fill LatencyNodeUser from LatencyToUsers
			key := fmt.Sprintf("('%s', '%s')", snap.NodeName, latency.UserIP)
			jsonNodeSnap.LatencyNodeUser[key] = latency.Latency // in milliseconds
		} // 2 decimal places

		for _, deployment := range snap.Deployments {
			// for each deployment, fill SLO from SLO in config and convert service time to service rate
			for _, deplodeployment := range config.Deployments {
				if deployment.DeploymentName == deplodeployment.DeploymentName {
					jsonNodeSnap.DeploymentName = deployment.DeploymentName
					jsonNodeSnap.SLO = float64(deplodeployment.SLO)
					jsonNodeSnap.ServiceRate = 1 / deplodeployment.ServiceTime
					jsonNodeSnap.RPS = deplodeployment.RPSPerUser
					jsonNodeSnap.OptiPref = deplodeployment.OptiPref
				}
			}
		}
	}
	return jsonNodeSnap
}

// GenerateJSONFile creates a JSON file from the JSONNodeSnap
func GenerateJSONFile(jsonNodeSnap JSONNodeSnap) {
	// Marshal the JSONNodeSnap into JSON
	jsonData, err := json.Marshal(jsonNodeSnap)
	if err != nil {
		log.Fatalf("Error marshalling JSONNodeSnap to JSON: %v", err)
	}

	// add timestamp to the file name
	t := time.Now()
	timestamp := t.Format("20060102150405")
	fileName := fmt.Sprintf("./data/snapshot/fodder_%s.json", timestamp)

	// Create the file
	file, err := os.Create(fileName)
	fmt.Println("fodder.json created")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	// Write the JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		log.Fatalf("Error writing JSON to file: %v", err)
	}
}
