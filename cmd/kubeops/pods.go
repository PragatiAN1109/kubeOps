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

// listPods returns every Pod in every namespace.
func listPods(clientset *kubernetes.Clientset) ([]corev1.Pod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	// An empty namespace ("") means all namespaces.
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not list pods (is the cluster reachable?): %w", err)
	}
	return pods.Items, nil
}

// isPodReady reports whether the Pod's Ready condition is True.
func isPodReady(pod corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// isPodHealthy treats Ready Pods and finished Pods (e.g. completed Jobs) as healthy.
func isPodHealthy(pod corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodSucceeded || isPodReady(pod)
}

// countPods returns the total, healthy and unhealthy Pod counts.
func countPods(pods []corev1.Pod) (total, healthy, unhealthy int) {
	for _, pod := range pods {
		if isPodHealthy(pod) {
			healthy++
		} else {
			unhealthy++
		}
	}
	return len(pods), healthy, unhealthy
}

// waitingReason returns the first waiting reason found in the Pod's
// init containers or regular containers, or "" if none are waiting.
func waitingReason(pod corev1.Pod) string {
	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			return cs.State.Waiting.Reason
		}
	}
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			return cs.State.Waiting.Reason
		}
	}
	return ""
}

// diagnosePod returns a short status and a suggested action for one Pod.
func diagnosePod(pod corev1.Pod) (status, action string) {
	// Container waiting reasons are the most specific signal, so check them first.
	switch waitingReason(pod) {
	case "CrashLoopBackOff":
		return "CrashLoopBackOff", "Container keeps crashing; check kubectl logs --previous"
	case "ImagePullBackOff":
		return "ImagePullBackOff", "Check image name, tag and registry access"
	case "ErrImagePull":
		return "ErrImagePull", "Check image name, tag and registry access"
	}

	switch pod.Status.Phase {
	case corev1.PodSucceeded:
		return "Completed", "None"
	case corev1.PodFailed:
		return "Failed", "Check kubectl describe pod and logs"
	case corev1.PodPending:
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionFalse {
				return "Pending (Unschedulable)", "Check events: resources, node selectors, taints or PVCs"
			}
		}
		return "Pending", "Check kubectl describe pod events"
	case corev1.PodRunning:
		if !isPodReady(pod) {
			return "Running (NotReady)", "Check readiness probe and container logs"
		}
		return "Healthy", "None"
	}

	return string(pod.Status.Phase), "Check kubectl describe pod"
}

func runHealth(clientset *kubernetes.Clientset) error {
	pods, err := listPods(clientset)
	if err != nil {
		return err
	}

	total, healthy, unhealthy := countPods(pods)
	fmt.Println("Pod Health Summary")
	fmt.Println("Total Pods:    ", total)
	fmt.Println("Healthy Pods:  ", healthy)
	fmt.Println("Unhealthy Pods:", unhealthy)
	return nil
}

func runDiagnose(clientset *kubernetes.Clientset) error {
	pods, err := listPods(clientset)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tPOD\tSTATUS\tSUGGESTED ACTION")
	for _, pod := range pods {
		status, action := diagnosePod(pod)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", pod.Namespace, pod.Name, status, action)
	}
	return w.Flush()
}
