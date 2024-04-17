package gurobi

import (
	"fmt"
	"os/exec"
)

func RunGurobi() {
	// Run gurobi model.ipynb
	cmd := exec.Command("python3", "plas_static/pkg/gurobi/model.py")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error running Gurobi:", err)
		return
	}
	fmt.Println(string(output))
}
