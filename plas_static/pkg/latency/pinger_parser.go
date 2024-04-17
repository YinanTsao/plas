package latency

import (
	dl "plas_static/pkg/dealer"
)

type SumNetworkTopology struct {
	NodeInfo            map[string]string
	LatencyNodesPair    []LatencyNodesPair
	LatencyNodeUserPair []LatencyNodeUserPair
}

type LatencyNodesPair struct {
	Node1   string
	Node2   string
	Latency float64
}

type LatencyNodeUserPair struct {
	Node    string
	User    string
	Latency float64
}

func NetworkTopologyParser(endpoints dl.Endpoints) []SumNetworkTopology {
	var sumNetworkTopologies []SumNetworkTopology

	for node, obj_ip := range endpoints.Node {

		url := "http://" + obj_ip + ":3300/networktopo"
		networktopo, _ := dl.NetworkTopologyGetter(url)

		if obj_ip == networktopo.NodeInfo[node] {

			var sumNetworkTopology SumNetworkTopology
			nodeInfo := make(map[string]string)

			nodeInfo[node] = obj_ip
			sumNetworkTopology.NodeInfo = nodeInfo

			for node2 := range networktopo.ToNodesLatency {
				sumNetworkTopology.LatencyNodesPair = append(sumNetworkTopology.LatencyNodesPair, LatencyNodesPair{Node1: node, Node2: node2, Latency: networktopo.ToNodesLatency[node2]})
				// fmt.Printf("Latency between %s and %s: %f\n", node, node2, networktopo.ToNodesLatency[node2])
			}
			for userIP := range networktopo.ToUserLatency {
				sumNetworkTopology.LatencyNodeUserPair = append(sumNetworkTopology.LatencyNodeUserPair, LatencyNodeUserPair{Node: node, User: userIP, Latency: networktopo.ToUserLatency[userIP]})
				// fmt.Printf("Latency between %s and user %s: %f\n", node, userIP, networktopo.ToUserLatency[userIP])
			}
			sumNetworkTopologies = append(sumNetworkTopologies, sumNetworkTopology)
		}

	}

	return sumNetworkTopologies
}
