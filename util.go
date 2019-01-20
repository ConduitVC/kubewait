package main

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

func GetStateDescriptionsFromEnv(env string) ([]StateDescription, error) {
	strval, ok := os.LookupEnv(env)
	if !ok {
		return []StateDescription{}, errors.New("value not found")
	}
	decoder := json.NewDecoder(strings.NewReader(strval))
	descriptions := make([]StateDescription, 0)
	err := decoder.Decode(&descriptions)
	if err != nil {
		return []StateDescription{}, err
	}
	return descriptions, nil
}

func MatchStateMap(current map[string]ResourceState, required []ResourceState) bool {
	// do not match if no resources are available
	if len(current) == 0 {
		return false
	}

	for _, c := range current {
		isRequiredState := false
		for _, rs := range required {
			if rs == c {
				isRequiredState = true
				break
			}
		}
		if !isRequiredState {
			return false
		}
	}
	return true
}
