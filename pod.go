package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	funk "github.com/thoas/go-funk"
)

var podPermittedStates = []ResourceState{ResourceReady, ResourceSucceeded, ResourceFailed}

// PodMatcher
type PodMatcher struct {
	clientset   kubernetes.Interface
	description StateDescription
	podstate    map[string]ResourceState
	watcher     watch.Interface
	done        chan bool
}

// PodValidator
type PodValidator struct {
	BaseValidator
}

func (p *PodValidator) Validate(ctx context.Context, description StateDescription) error {
	err := p.BaseValidator.Validate(ctx, description)
	if err != nil {
		return err
	}

	for _, requiredState := range description.RequiredStates {
		if !funk.Contains(podPermittedStates, requiredState) {
			return ErrStateNotValidForResourceType(description, requiredState)
		}
	}
	return nil
}

func NewPodValidator() Validator {
	return &PodValidator{}
}

func NewPodMatcher(clientset kubernetes.Interface, description StateDescription) Matcher {
	return &PodMatcher{
		clientset:   clientset,
		description: description,
		done:        make(chan bool, 1),
		podstate:    make(map[string]ResourceState),
	}
}

func (p *PodMatcher) Start(ctx context.Context) error {
	options := metav1.ListOptions{
		LabelSelector: p.description.LabelSelector,
	}
	logger := log.WithFields(log.Fields{
		"namespace":     p.description.Namespace,
		"type":          p.description.Type,
		"labelselector": p.description.LabelSelector,
	})

	logger.Debug("fetching initial context")

	pods, err := p.clientset.CoreV1().Pods(p.description.Namespace).List(options)
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		state := getPodResourceState(&pod)
		p.podstate[pod.Name] = state

		log.WithFields(log.Fields{
			"podName":  pod.Name,
			"podState": state,
		}).Debug("added to podstate")
	}

	logger.Debug("fetched context")
	if match := MatchStateMap(p.podstate, p.description.RequiredStates); match {
		logger.Debug("match: ", match)
		return nil
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
			state := getPodResourceState(pod)
			p.podstate[pod.Name] = state
			ctxLogger.WithFields(log.Fields{
				"podName":  pod.Name,
				"podState": state,
			}).Debug("added to pod state")
		case watch.Modified:
			pod := event.Object.(*v1.Pod)
			state := getPodResourceState(pod)
			p.podstate[pod.Name] = state
			ctxLogger.WithFields(log.Fields{
				"podName":  pod.Name,
				"podState": state,
			}).Debug("updated pod state")
		case watch.Deleted:
			pod := event.Object.(*v1.Pod)
			_, ok := p.podstate[pod.Name]
			if ok {
				delete(p.podstate, pod.Name)
				ctxLogger.WithFields(log.Fields{
					"podName": pod.Name,
				}).Debug("removed from pod state")
			}
		case watch.Error:
			// TODO: Do something with this error
			return nil
		}

		if MatchStateMap(p.podstate, p.description.RequiredStates) {
			log.Info("state description matched by cluster")
			select {
			case <-p.done:
			default:
				close(p.done)
			}
			break
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
	if p.watcher != nil {
		p.watcher.Stop()
	}
	return nil
}

func getPodResourceState(pod *v1.Pod) ResourceState {
	// check for ready
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return ResourceReady
		}
	}
	switch pod.Status.Phase {
	case v1.PodSucceeded:
		return ResourceSucceeded
	case v1.PodFailed:
		return ResourceFailed
	default:
	}
	return resourceWaiting
}
