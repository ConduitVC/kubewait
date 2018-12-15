package main

import "fmt"

type ValidationError struct {
	StateDescription
	Message string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("%s: %v", v.Message, v.StateDescription)
}

func ErrNoRequiredStates(description StateDescription) error {
	return &ValidationError{
		Message:          "no \"requiredStates\" provded for resource",
		StateDescription: description,
	}
}

func ErrWaitingStateReserved(description StateDescription) error {
	return &ValidationError{
		Message:          "\"waiting\" state is reserved for internal use",
		StateDescription: description,
	}
}

func ErrStateNotValidForResourceType(description StateDescription, state ResourceState) error {
	return &ValidationError{
		Message:          fmt.Sprintf("\"%s\" state is not valid for resource type \"%s\"", state, description.Type),
		StateDescription: description,
	}
}
