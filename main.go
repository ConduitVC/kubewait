package main

import (
	"context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	config, err := rest.InClusterConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	description := &StateDescription{
		Type:           "Pod",
		LabelSelector:  "",
		Namespace:      "kube-system",
		RequiredStates: []ResourceState{ResourceReady},
	}
	ctx := context.Background()
	matcher := NewPodMatcher(clientset, description)
	matcher.Start(ctx)
}
