package main

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestGetStateDescriptionsFromEnv(t *testing.T) {
	const jsonDescription = `[{
        "type": "Pod",
        "labelSelector": "",
        "requiredStates": [ "Ready" ],
        "namespace": "default"
    }]`

	const kubewaitEnv = "KUBEWAIT_ENV"

	os.Setenv(kubewaitEnv, jsonDescription)
	descriptions, err := GetStateDescriptionsFromEnv(kubewaitEnv)
	if err != nil {
		t.Fatal(err)
	}
	log.Debug(descriptions)
}
