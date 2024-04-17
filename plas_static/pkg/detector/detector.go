package detector

import sn "plas_static/pkg/snapshotter"

type SLOViolationLogger struct {
	DeploymentName string
	UserIP         string
	NodeName       string
	RTT            float64
	SLO            int
}

// Takes the RTT results and compares them to the SLO, returning a boolean indicating whether the SLO is met
func SLOViolationDetector(NodeSnap []sn.NodeSnap, SLO int) ([]SLOViolationLogger, int, int) {
	var SLOViolationLog []SLOViolationLogger
	sumUserRTT := 0
	Violation := 0

	for _, nodeSnap := range NodeSnap {
		for _, deploymentSnap := range nodeSnap.Deployments {
			for _, rttResult := range deploymentSnap.RTT {
				sumUserRTT++
				if int(rttResult.RTT) > SLO {
					SLOViolationLog = append(SLOViolationLog, SLOViolationLogger{
						DeploymentName: deploymentSnap.DeploymentName,
						UserIP:         rttResult.UserIP,
						NodeName:       nodeSnap.NodeName,
						RTT:            rttResult.RTT,
						SLO:            SLO,
					})
					Violation++
				}

			}

		}
	}

	return SLOViolationLog, sumUserRTT, Violation
}
