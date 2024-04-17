package executor

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type PlacementPlan struct {
	NodeReplicas   map[string]int      `json:"node_replicas"`
	SiteUserMap    map[string][]string `json:"site_user_map"`
	DeploymentName string              `json:"deployment_name"`
}

func JSONToPlacementPlan() PlacementPlan {
	// Get all files in the current directory
	files, err := os.ReadDir("../../data/placement/")
	if err != nil {
		log.Fatal(err)
	}

	// Filter JSON files and sort them in reverse order by name
	var jsonFiles []os.DirEntry
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "placement_plan_") && strings.HasSuffix(file.Name(), ".json") {
			jsonFiles = append(jsonFiles, file)
		}
	}
	sort.Slice(jsonFiles, func(i, j int) bool {
		return jsonFiles[i].Name() > jsonFiles[j].Name()
	})

	if len(jsonFiles) == 0 {
		log.Fatal("No JSON files found")
	}

	// Open the most recent file
	file, err := os.Open(filepath.Join("./data/placement/", jsonFiles[0].Name()))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read the file into a byte slice
	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal the JSON into a PlacementPlan struct
	var plan PlacementPlan
	if err := json.Unmarshal(bytes, &plan); err != nil {
		log.Fatal(err)
	}
	// Print the plan
	return plan
}

func PlacementPlanPrinter(plan PlacementPlan) {
	fmt.Println("---------------------------------")
	fmt.Println("New Placement Plan:")
	fmt.Printf("Deployment: %s\n", plan.DeploymentName)
	for node, replicas := range plan.NodeReplicas {
		fmt.Printf("Node: %s should have %d replicas\n", node, replicas)
	}
	for site, users := range plan.SiteUserMap {
		fmt.Printf("Site: %s, Users: %v\n", site, users)
	}
}
