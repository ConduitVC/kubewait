package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	funk "github.com/thoas/go-funk"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

var jobPermittedStates = []ResourceState{ResourceComplete, ResourceFailed, ResourceRunning}

type JobMatcher struct {
	clientset   kubernetes.Interface
	watcher     watch.Interface
	description StateDescription
	done        chan bool
	jobstate    map[string]ResourceState
}

type JobValidator struct {
	BaseValidator
}

func (v *JobValidator) Validate(ctx context.Context, description StateDescription) error {
	err := v.BaseValidator.Validate(ctx, description)
	if err != nil {
		return err
	}
	for _, requiredState := range description.RequiredStates {
		if !funk.Contains(jobPermittedStates, requiredState) {
			return ErrStateNotValidForResourceType(description, requiredState)
		}
	}
	return nil
}

func NewJobValidator() Validator {
	return &JobValidator{}
}

func NewJobMatcher(clientset kubernetes.Interface, description StateDescription) Matcher {
	return &JobMatcher{
		clientset:   clientset,
		watcher:     nil,
		description: description,
		done:        make(chan bool, 1),
		jobstate:    make(map[string]ResourceState),
	}
}

func (m *JobMatcher) Start(ctx context.Context) error {
	options := metav1.ListOptions{
		LabelSelector: m.description.LabelSelector,
	}

	log.WithFields(log.Fields{
		"namespace":     m.description.Namespace,
		"type":          m.description.Type,
		"labelselector": m.description.LabelSelector,
	}).Debug("fetching initial context")

	jobs, err := m.clientset.BatchV1().Jobs(m.description.Namespace).List(options)
	if err != nil {
		return err
	}

	for _, job := range jobs.Items {
		state := getJobResourceState(&job)
		m.jobstate[job.Name] = state

		log.WithFields(log.Fields{
			"jobName":  job.Name,
			"jobState": state,
		}).Debug("added to jobstate")
	}

	if MatchStateMap(m.jobstate, m.description.RequiredStates) {
		return nil
	}

	m.watcher, err = m.clientset.BatchV1().Jobs(m.description.Namespace).Watch(options)
	if err != nil {
		return err
	}

	log.Debug("watching for updates")
	for event := range m.watcher.ResultChan() {
		ctxLogger := log.WithFields(log.Fields{
			"eventType": event.Type,
		})
		switch event.Type {
		case watch.Added:
			job := event.Object.(*batchv1.Job)
			state := getJobResourceState(job)
			m.jobstate[job.Name] = state

			ctxLogger.WithFields(log.Fields{
				"jobName":  job.Name,
				"jobState": state,
			}).Debug("added to job state")
		case watch.Modified:
			job := event.Object.(*batchv1.Job)
			state := getJobResourceState(job)
			m.jobstate[job.Name] = state

			ctxLogger.WithFields(log.Fields{
				"jobName":  job.Name,
				"jobState": state,
			}).Debug("updated job state")
		case watch.Deleted:
			job := event.Object.(*batchv1.Job)
			_, ok := m.jobstate[job.Name]
			if ok {
				ctxLogger.WithFields(log.Fields{
					"jobName": job.Name,
				}).Debug("deleted from job state")
			}
		case watch.Error:
			// TODO: do something with this error
		}
		if MatchStateMap(m.jobstate, m.description.RequiredStates) {
			log.Info("state description matched by cluster")
			select {
			case <-m.done:
			default:
				close(m.done)
			}
			break
		}
	}
	return nil
}

func (m *JobMatcher) Done() <-chan bool {
	return m.done
}

func (m *JobMatcher) Stop(ctx context.Context) error {
	defer func() {
		select {
		case <-m.done:
		default:
			close(m.done)
		}
	}()
	if m.watcher != nil {
		m.watcher.Stop()
	}
	return nil
}

func getJobResourceState(job *batchv1.Job) ResourceState {
	for _, condition := range job.Status.Conditions {
		// An explicit check is added for JobFailed to allow for addition
		// of more job conditions in future Kubernetes releases
		if condition.Type == batchv1.JobComplete {
			if condition.Status == v1.ConditionTrue {
				return ResourceComplete
			}
		} else if condition.Type == batchv1.JobFailed {
			if condition.Status == v1.ConditionTrue {
				return ResourceFailed
			}
		}
	}
	// check if any containers are active in the pod to check if pod is running
	if job.Status.Active != 0 {
		return ResourceRunning
	}
	return resourceWaiting
}
