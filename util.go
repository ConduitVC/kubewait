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
