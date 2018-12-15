package main

import (
	"context"
	// "k8s.io/api/core/v1"
)

type ResourceState string
type ResourceType string

// Empty cluster state is always matched
type StateDescription struct {
	Type           ResourceType    `json:"type"`
	LabelSelector  string          `json:"labelSelector"`
	RequiredStates []ResourceState `json:"requiredStates"`
	Namespace      string          `json:"namespace",omitempty`
}

const (
	Pod ResourceType = "Pod"
	Job ResourceType = "Job"
)

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

type Validator interface {
	Validate(StateDescription) error
}
