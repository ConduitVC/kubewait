package main

import (
	"context"
	"sync"

	"k8s.io/client-go/kubernetes"
)

func wait(ctx context.Context, clientset kubernetes.Interface, descriptions []StateDescription) {
	for _, description := range descriptions {
		validator, ok := getValidator(clientset, description)
		if !ok {
			panic("could not find validator for resource type " + description.Type)
		}
		if err := validator.Validate(description); err != nil {
			panic("description not valid: " + err.Error())
		}
	}

	var wg sync.WaitGroup
	for _, description := range descriptions {
		matcher, ok := getMatcher(clientset, description)
		if !ok {
			panic("could not find matcher for resource type " + description.Type)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			matcher.Start(ctx)
		}()
	}
	wg.Wait()
}

func getValidator(clientset kubernetes.Interface, description StateDescription) (Validator, bool) {
	switch description.Type {
	case Pod:
		return NewPodValidator(description), true
	}
	return nil, false
}

func getMatcher(clientset kubernetes.Interface, description StateDescription) (Matcher, bool) {
	switch description.Type {
	case Pod:
		return NewPodMatcher(clientset, description), true
	}
	return nil, false
}
