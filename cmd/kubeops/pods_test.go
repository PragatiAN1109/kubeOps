package main

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

// newPod builds a small Pod with the given phase, Ready condition and waiting reason.
func newPod(phase corev1.PodPhase, ready bool, waiting string) corev1.Pod {
	readyStatus := corev1.ConditionFalse
	if ready {
		readyStatus = corev1.ConditionTrue
	}

	pod := corev1.Pod{}
	pod.Status.Phase = phase
	pod.Status.Conditions = []corev1.PodCondition{
		{Type: corev1.PodReady, Status: readyStatus},
	}
	if waiting != "" {
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{
			{State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: waiting}}},
		}
	}
	return pod
}

func TestDiagnosePod(t *testing.T) {
	unschedulable := newPod(corev1.PodPending, false, "")
	unschedulable.Status.Conditions = append(unschedulable.Status.Conditions,
		corev1.PodCondition{Type: corev1.PodScheduled, Status: corev1.ConditionFalse, Reason: "Unschedulable"})

	tests := []struct {
		name string
		pod  corev1.Pod
		want string
	}{
		{"healthy", newPod(corev1.PodRunning, true, ""), "Healthy"},
		{"crash loop", newPod(corev1.PodRunning, false, "CrashLoopBackOff"), "CrashLoopBackOff"},
		{"image pull backoff", newPod(corev1.PodPending, false, "ImagePullBackOff"), "ImagePullBackOff"},
		{"err image pull", newPod(corev1.PodPending, false, "ErrImagePull"), "ErrImagePull"},
		{"unschedulable", unschedulable, "Pending (Unschedulable)"},
		{"pending creating", newPod(corev1.PodPending, false, "ContainerCreating"), "Pending"},
		{"running not ready", newPod(corev1.PodRunning, false, ""), "Running (NotReady)"},
		{"completed", newPod(corev1.PodSucceeded, false, ""), "Completed"},
		{"failed", newPod(corev1.PodFailed, false, ""), "Failed"},
	}

	for _, tt := range tests {
		got, _ := diagnosePod(tt.pod)
		if got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestCountPods(t *testing.T) {
	pods := []corev1.Pod{
		newPod(corev1.PodRunning, true, ""),
		newPod(corev1.PodRunning, false, ""),
		newPod(corev1.PodPending, false, "ImagePullBackOff"),
		newPod(corev1.PodSucceeded, false, ""),
	}

	total, healthy, unhealthy := countPods(pods)
	if total != 4 || healthy != 2 || unhealthy != 2 {
		t.Errorf("got total=%d healthy=%d unhealthy=%d, want 4, 2, 2", total, healthy, unhealthy)
	}
}
