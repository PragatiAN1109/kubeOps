package main

import (
	"encoding/json"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
)

// Report is the JSON summary printed by the report command.
type Report struct {
	TotalPods     int    `json:"totalPods"`
	HealthyPods   int    `json:"healthyPods"`
	UnhealthyPods int    `json:"unhealthyPods"`
	NodeCount     int    `json:"nodeCount"`
	Timestamp     string `json:"timestamp"`
}

func runReport(clientset *kubernetes.Clientset) error {
	pods, err := listPods(clientset)
	if err != nil {
		return err
	}
	nodes, err := listNodes(clientset)
	if err != nil {
		return err
	}

	total, healthy, unhealthy := countPods(pods)
	report := Report{
		TotalPods:     total,
		HealthyPods:   healthy,
		UnhealthyPods: unhealthy,
		NodeCount:     len(nodes),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
