package main

// ResourceState describes the states a resource can be in.
type ResourceState string

// ResourceType describes the types of resources that can be watched.
type ResourceType string

// StateDescription is a JSON description of a resource and the state that the resource must be in
// so that the cluster state match succeeds. If no such resources are found, the match does not succeed.
type StateDescription struct {
	Type           ResourceType    `json:"type"`
	LabelSelector  string          `json:"labelSelector",omitempty`
	RequiredStates []ResourceState `json:"requiredStates"`
	Namespace      string          `json:"namespace",omitempty`
}

const (
	// PodResource is used to match k8s pods.
	PodResource ResourceType = "Pod"
	// JobResource is used to match k8s jobs.
	JobResource ResourceType = "Job"
)

const (
	ResourceReady     ResourceState = "Ready"
	ResourceSucceeded ResourceState = "Succeeded"
	ResourceFailed    ResourceState = "Failed"
	resourceWaiting   ResourceState = "waiting"
	ResourceComplete  ResourceState = "Complete"
	ResourceRunning   ResourceState = "Running"
)
