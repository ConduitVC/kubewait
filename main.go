package main

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const DefaultEnv = "KUBEWAIT"

func init() {
	if env, _ := os.LookupEnv("ENV"); env == "DEBUG" {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	config, err := rest.InClusterConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	log.Debugf("loaded kubernetes clientset\n")
	descriptions, err := GetStateDescriptionsFromEnv(DefaultEnv)
	if err != nil {
		panic(err)
	}
	log.Debugf("loaded state descriptions: %v\n", descriptions)
	ctx := context.Background()
	wait(ctx, clientset, descriptions)
}
