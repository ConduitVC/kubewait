package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	funk "github.com/thoas/go-funk"
)

type Validator interface {
	Validate(context.Context, StateDescription) error
}

type BaseValidator struct{}

func (BaseValidator) Validate(ctx context.Context, description StateDescription) error {
	log.Debugf("validating: %v", description)
	if len(description.RequiredStates) == 0 {
		return ErrNoRequiredStates(description)
	}
	if funk.Contains(description.RequiredStates, resourceWaiting) {
		log.Debug("description contains waiting as required state...failing")
		return ErrWaitingStateReserved(description)
	}
	return nil
}
