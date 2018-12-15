package main

import (
	"context"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	testcore "k8s.io/client-go/testing"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestJobComplete(t *testing.T) {
	description := StateDescription{
		Namespace:      "test-ns",
		Type:           JobResource,
		LabelSelector:  "app=test",
		RequiredStates: []ResourceState{ResourceComplete},
	}
	fake := fakeclientset.NewSimpleClientset()
	jobList := &batchv1.JobList{
		Items: []batchv1.Job{
			batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
					},
					Name:      "job-1",
					Namespace: "test-ns",
				},
				Status: batchv1.JobStatus{
					Active:    1,
					Succeeded: 0,
					Failed:    0,
					Conditions: []batchv1.JobCondition{
						batchv1.JobCondition{
							Type:   batchv1.JobComplete,
							Status: v1.ConditionFalse,
						},
						batchv1.JobCondition{
							Type:   batchv1.JobFailed,
							Status: v1.ConditionFalse,
						},
					},
				},
			},
		},
	}
	watcher := watch.NewFakeWithChanSize(1, false)
	fake.PrependReactor("list", "jobs", func(action testcore.Action) (bool, runtime.Object, error) {
		return true, jobList, nil
	})
	fake.PrependWatchReactor("jobs", testcore.DefaultWatchReactor(watcher, nil))

	matcher := NewJobMatcher(fake, description)
	go matcher.Start(context.Background())

	// simulate real watch
	for _, job := range jobList.Items {
		watcher.Add(&job)
	}

	select {
	case <-matcher.Done():
		t.Fatal("matcher should not return")
	default:
	}

	sleepDuration, _ := time.ParseDuration("1s")
	time.Sleep(sleepDuration)
	watcher.Modify(&batchv1.Job{
		ObjectMeta: jobList.Items[0].ObjectMeta,
		Status: batchv1.JobStatus{
			Active:    0,
			Succeeded: 1,
			Failed:    0,
			Conditions: []batchv1.JobCondition{
				batchv1.JobCondition{
					Type:   batchv1.JobComplete,
					Status: v1.ConditionTrue,
				},
				batchv1.JobCondition{
					Type:   batchv1.JobFailed,
					Status: v1.ConditionFalse,
				},
			},
		},
	})

	// wait for matcher
	timeoutDuration, _ := time.ParseDuration("500ms")
	select {
	case <-time.After(timeoutDuration):
		t.Fatalf("matcher did not return after 500ms")
	case <-matcher.Done():
	}
}
