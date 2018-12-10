package main

import (
	"context"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	testcore "k8s.io/client-go/testing"
)

func TestPodMatcherSinglePod(t *testing.T) {
	description := &StateDescription{
		Namespace:      "test-ns",
		Type:           "Pod",
		LabelSelector:  "",
		RequiredPhases: []string{v1.PodRunning},
	}
	podlist := &v1.PodList{
		Items: []v1.Pod{
			v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "test-ns",
				},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
			},
		},
	}
	fake := fakeclientset.NewSimpleClientset()
	watcher := watch.NewFakeWithChanSize(1, false)
	fake.PrependReactor("list", "pods", func(action testcore.Action) (bool, runtime.Object, error) {
		return true, podlist, nil
	})
	fake.PrependWatchReactor("pods", testcore.DefaultWatchReactor(watcher, nil))
	matcher := NewPodMatcher(fake, description)
	go matcher.Start(context.Background())

	//simulate watch update
	sleepDuration, _ := time.ParseDuration("100ms")
	time.Sleep(sleepDuration)
	watcher.Modify(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "test-ns",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
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

func TestPodMatcherPodAdded(t *testing.T) {
	description := &StateDescription{
		Type:           "Pod",
		Namespace:      "test-ns",
		LabelSelector:  "",
		RequiredPhases: []string{"Succeeded"},
	}
}
