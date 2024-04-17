package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/go-ping/ping"
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

func Pinger(endpoints Endpoints) NetworkTopology {

	localIP, err := GetLocalIP()
	if err != nil {
		fmt.Printf("Failed to get local IP: %v\n", err)
		return NetworkTopology{}
	}

	hostname, err := GetHostname()
	if err != nil {
		fmt.Printf("Failed to get hostname: %v\n", err)
		return NetworkTopology{}
	}

	topology := NetworkTopology{
		NodeInfo:       make(map[string]string),
		ToNodesLatency: make(map[string]float64),
		ToUserLatency:  make(map[string]float64),
	}

	topology.NodeInfo[hostname] = localIP

	// Pinging node endpoints
	for nodeName, ip := range endpoints.Node {
		if ip == localIP {
			continue
		}

		pinger, err := ping.NewPinger(ip)
		if err != nil {
			fmt.Printf("Failed to create pinger for node %s: %v\n", nodeName, err)
			continue
		}

		pinger.Count = 3
		pinger.Timeout = time.Second * 10
		pinger.Run()

		stats := pinger.Statistics()
		topology.ToNodesLatency[nodeName] = stats.AvgRtt.Seconds() * 1000 // in ms
	}

	// Pinging user endpoints
	for _, ip := range endpoints.User {
		pinger, err := ping.NewPinger(ip)
		if err != nil {
			fmt.Printf("Failed to create pinger for user: %v\n", err)
		} else {
			pinger.Count = 3
			pinger.Timeout = time.Second * 10
			pinger.Run()

			stats := pinger.Statistics()
			topology.ToUserLatency[ip] = stats.AvgRtt.Seconds() * 1000 // in ms
		}
	}
	return topology
}

// GetLocalIP returns the local IP address of the machine
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %v", err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()
				if strings.HasPrefix(ip, "130.104") {
					return ip, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no IP address starting with 130.104 found")
}

func GetHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %v", err)
	}

	return hostname, nil
}
