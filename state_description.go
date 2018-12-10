package main

import (
	"context"
	// "k8s.io/api/core/v1"
)

// Empty cluster state is always matched
type StateDescription struct {
	Type           string
	LabelSelector  string
	RequiredStates []ResourceState
	Namespace      string
}

type ResourceState string

const (
	ResourceReady     ResourceState = "Ready"
	ResourceSucceeded ResourceState = "Succeeded"
	ResourceFailed    ResourceState = "Failed"
	resourceWaiting   ResourceState = "waiting"
)

type Matcher interface {
	Start(context.Context) error
	Done() <-chan bool
	Stop(context.Context) error
}

func (StateDescription) ParseString(ctx context.Context, desc string) (*StateDescription, error) {
	return nil, nil
}
