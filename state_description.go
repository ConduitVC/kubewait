package main

type ResourceState string
type ResourceType string

// Empty cluster state is always matched
type StateDescription struct {
	Type           ResourceType    `json:"type"`
	LabelSelector  string          `json:"labelSelector",omitempty`
	ResourceName   string          `json:"resourceName",omitempty`
	RequiredStates []ResourceState `json:"requiredStates"`
	Namespace      string          `json:"namespace",omitempty`
}

const (
	PodResource ResourceType = "Pod"
	JobResource ResourceType = "Job"
)

const (
	ResourceReady     ResourceState = "Ready"
	ResourceSucceeded ResourceState = "Succeeded"
	ResourceFailed    ResourceState = "Failed"
	resourceWaiting   ResourceState = "waiting"
	ResourceComplete  ResourceState = "Complete"
	// Running used only for jobs
	ResourceRunning ResourceState = "Running"
)
