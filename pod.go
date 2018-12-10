package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type PodMatcher struct {
	clientset   kubernetes.Interface
	description *StateDescription
	podstate    map[string]v1.PodPhase
	watcher     watch.Interface
	done        chan bool
}

func NewPodMatcher(clientset kubernetes.Interface, description *StateDescription) *PodMatcher {
	return &PodMatcher{
		clientset:   clientset,
		description: description,
		done:        make(chan bool, 1),
		podstate:    make(map[string]v1.PodPhase),
	}
}

func (p *PodMatcher) Start(ctx context.Context) error {
	options := metav1.ListOptions{
		LabelSelector: p.description.LabelSelector,
	}
	log.WithFields(log.Fields{
		"namespace":     p.description.Namespace,
		"labelselector": p.description.LabelSelector,
	}).Info("fetching initial context")
	pods, err := p.clientset.CoreV1().Pods(p.description.Namespace).List(options)
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		p.podstate[pod.Name] = pod.Status.Phase
		log.WithFields(log.Fields{
			"podName":  pod.Name,
			"podPhase": pod.Status.Phase,
		}).Info("added to podstate")
	}
	p.watcher, err = p.clientset.CoreV1().Pods(p.description.Namespace).Watch(options)
	if err != nil {
		return err
	}
	log.Info("watching for updates")
	for event := range p.watcher.ResultChan() {
		ctxLogger := log.WithFields(log.Fields{
			"eventType": event.Type,
		})

		switch event.Type {
		case watch.Added:
			pod := event.Object.(*v1.Pod)
			p.podstate[pod.Name] = pod.Status.Phase
			ctxLogger.WithFields(log.Fields{
				"podName":  pod.Name,
				"podPhase": pod.Status.Phase,
			}).Info("added to pod state")
		case watch.Modified:
			pod := event.Object.(*v1.Pod)
			p.podstate[pod.Name] = pod.Status.Phase
			ctxLogger.WithFields(log.Fields{
				"podName":  pod.Name,
				"podPhase": pod.Status.Phase,
			}).Info("updated pod state")
		case watch.Deleted:
			pod := event.Object.(*v1.Pod)
			_, ok := p.podstate[pod.Name]
			if ok {
				delete(p.podstate, pod.Name)
				ctxLogger.WithFields(log.Fields{
					"podName":  pod.Name,
					"podPhase": pod.Status.Phase,
				}).Info("removed from pod state")
			}
		case watch.Error:
			// TODO: Do something with this error
			return nil
		}

		if p.match() {
			log.Info("state matched by cluster")
			select {
			case <-p.done:
			default:
				close(p.done)
			}
			return nil
		}
	}
	return nil
}

func (p *PodMatcher) Done() <-chan bool {
	return p.done
}

func (p *PodMatcher) Stop(ctx context.Context) error {
	defer func() {
		select {
		case <-p.done:
		default:
			close(p.done)
		}
	}()
	p.watcher.Stop()
	return nil
}

func (p *PodMatcher) match() bool {
	for _, currentPhase := range p.podstate {
		isRequiredPhase := false
		for _, phase := range p.description.RequiredPhases {
			if phase == currentPhase {
				isRequiredPhase = true
			}
		}
		if !isRequiredPhase {
			return false
		}
	}
	return true
}
