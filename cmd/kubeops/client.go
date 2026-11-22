package main

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// newClientset builds a Kubernetes client the same way kubectl does:
// it reads $KUBECONFIG (which may list several files) or falls back to
// ~/.kube/config, and uses the current context.
func newClientset() (*kubernetes.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not load kubeconfig (set $KUBECONFIG or create ~/.kube/config): %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create Kubernetes client: %w", err)
	}
	return clientset, nil
}
