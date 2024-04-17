package snapshotter

import "fmt"

// ResponseTimeEstimator estimates the response time of a system using the M/D/1 queueing model
func ResponseTimeEstimator(serviceTime float64, numServers int, arrivalRate float64) float64 {
	utilization := arrivalRate * serviceTime / float64(numServers)
	fmt.Printf("arrivalRate: %f, serviceTime: %f, numServers: %d, utilization: %f\n", arrivalRate, serviceTime, numServers, utilization)
	responseTime := serviceTime + serviceTime*utilization/(2*(1-utilization)) // Here the unit is in seconds, since it's req/s
	fmt.Printf("serviceTime: %f, responseTime: %f\n", serviceTime, responseTime)
	// Convert ResponseTime to milliseconds and then to a time.Duration
	return responseTime * 1000 // s to ms
}

// RoundTripTimeEstimator estimates the round trip time of a system using the M/D/1 queueing model
func RoundTripTimeEstimator(responseTime float64, latency float64) float64 {
	RTT := responseTime + latency
	return RTT
}
