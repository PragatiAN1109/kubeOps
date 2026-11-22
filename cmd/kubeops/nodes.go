package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// listNodes returns every node in the cluster.
func listNodes(clientset *kubernetes.Clientset) ([]corev1.Node, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not list nodes (is the cluster reachable?): %w", err)
	}
	return nodes.Items, nil
}

// nodeStatus returns "Ready" if the node's Ready condition is True, otherwise "NotReady".
func nodeStatus(node corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return "Ready"
		}
	}
	return "NotReady"
}

func runNodes(clientset *kubernetes.Clientset) error {
	nodes, err := listNodes(clientset)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tCPU\tMEMORY")
	for _, node := range nodes {
		cpu := node.Status.Capacity[corev1.ResourceCPU]
		memory := node.Status.Capacity[corev1.ResourceMemory]
		memoryMi := memory.Value() / (1024 * 1024)
		fmt.Fprintf(w, "%s\t%s\t%s\t%dMi\n", node.Name, nodeStatus(node), cpu.String(), memoryMi)
	}
	return w.Flush()
}
