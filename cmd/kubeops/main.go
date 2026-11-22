// Command kubeops is a small, read-only CLI for checking a Kubernetes cluster.
package main

import (
	"fmt"
	"os"
	"time"
)

// apiTimeout limits how long any single Kubernetes API call may take.
const apiTimeout = 10 * time.Second

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: kubeops <command>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  health    Summarize Pod health across all namespaces")
	fmt.Fprintln(os.Stderr, "  diagnose  List Pods with detected problems and suggested actions")
	fmt.Fprintln(os.Stderr, "  nodes     List nodes with status, CPU and memory")
	fmt.Fprintln(os.Stderr, "  report    Print a JSON summary of the cluster")
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(2)
	}

	command := os.Args[1]
	switch command {
	case "health", "diagnose", "nodes", "report":
	case "help", "-h", "--help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q\n\n", command)
		usage()
		os.Exit(2)
	}

	clientset, err := newClientset()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	switch command {
	case "health":
		err = runHealth(clientset)
	case "diagnose":
		err = runDiagnose(clientset)
	case "nodes":
		err = runNodes(clientset)
	case "report":
		err = runReport(clientset)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
