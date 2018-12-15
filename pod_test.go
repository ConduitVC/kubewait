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

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestPodMatcherSinglePod(t *testing.T) {
	description := StateDescription{
		Namespace:      "test-ns",
		Type:           "Pod",
		LabelSelector:  "",
		RequiredStates: []ResourceState{ResourceReady},
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
					Conditions: []v1.PodCondition{
						v1.PodCondition{
							Type:   v1.PodReady,
							Status: v1.ConditionFalse,
						},
					},
				},
			},
		},
	}
	fake := fakeclientset.NewSimpleClientset()
	watcher := watch.NewFakeWithChanSize(1, false)
	fake.PrependReactor("list", "pods", func(action testcore.Action) (bool, runtime.Object, error) {
		log.Info("returning fake pods")
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
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
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

func TestPodMatcherPodAdded(t *testing.T) {
	description := StateDescription{
		Type:           "Pod",
		Namespace:      "test-ns",
		LabelSelector:  "",
		RequiredStates: []ResourceState{ResourceReady},
	}
	podlist := &v1.PodList{
		Items: []v1.Pod{
			v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "test-ns",
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					Conditions: []v1.PodCondition{
						v1.PodCondition{
							Type:   v1.PodReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
		},
	}
	fake := fakeclientset.NewSimpleClientset()
	watcher := watch.NewFakeWithChanSize(1, false)
	fake.PrependReactor("list", "pods", func(action testcore.Action) (bool, runtime.Object, error) {
		log.Info("returning fake pods")
		return true, podlist, nil
	})
	fake.PrependWatchReactor("pods", testcore.DefaultWatchReactor(watcher, nil))
	matcher := NewPodMatcher(fake, description)
	go matcher.Start(context.Background())
	// Add a new pod
	watcher.Add(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: "test-ns",
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionFalse,
				},
			},
		},
	})

	select {
	case <-matcher.Done():
		t.Fatalf("should not succeed when added pod is in Pending phase")
	default:
	}

	watcher.Modify(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: "test-ns",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	})

	timeoutDuration, _ := time.ParseDuration("500ms")
	select {
	case <-time.After(timeoutDuration):
		t.Fatalf("matcher did not return after 500ms")
	case <-matcher.Done():
	}
}

func TestPodMatcherPendingPodDeleted(t *testing.T) {
	description := StateDescription{
		Type:           "Pod",
		Namespace:      "test-ns",
		LabelSelector:  "",
		RequiredStates: []ResourceState{ResourceReady},
	}
	podlist := &v1.PodList{
		Items: []v1.Pod{
			v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "test-ns",
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					Conditions: []v1.PodCondition{
						v1.PodCondition{
							Type:   v1.PodReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-2",
					Namespace: "test-ns",
				},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
					Conditions: []v1.PodCondition{
						v1.PodCondition{
							Type:   v1.PodReady,
							Status: v1.ConditionFalse,
						},
					},
				},
			},
		},
	}
	fake := fakeclientset.NewSimpleClientset()
	watcher := watch.NewFakeWithChanSize(1, false)
	fake.PrependReactor("list", "pods", func(action testcore.Action) (bool, runtime.Object, error) {
		log.Info("returning fake pods")
		return true, podlist, nil
	})
	fake.PrependWatchReactor("pods", testcore.DefaultWatchReactor(watcher, nil))
	matcher := NewPodMatcher(fake, description)
	go matcher.Start(context.Background())
	// Delete pending pod
	watcher.Delete(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: "test-ns",
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionFalse,
				},
			},
		},
	})

	timeoutDuration, _ := time.ParseDuration("500ms")
	select {
	case <-time.After(timeoutDuration):
		t.Fatalf("matcher did not return after 500ms")
	case <-matcher.Done():
	}
}

func TestPodMatcherDeleteRunningPod(t *testing.T) {
	description := StateDescription{
		Type:           "Pod",
		Namespace:      "test-ns",
		LabelSelector:  "",
		RequiredStates: []ResourceState{ResourceReady},
	}
	podlist := &v1.PodList{
		Items: []v1.Pod{
			v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "test-ns",
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					Conditions: []v1.PodCondition{
						v1.PodCondition{
							Type:   v1.PodReady,
							Status: v1.ConditionTrue,
						},
					},
				},
			},
			v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-2",
					Namespace: "test-ns",
				},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
					Conditions: []v1.PodCondition{
						v1.PodCondition{
							Type:   v1.PodReady,
							Status: v1.ConditionFalse,
						},
					},
				},
			},
		},
	}
	fake := fakeclientset.NewSimpleClientset()
	watcher := watch.NewFakeWithChanSize(1, false)
	fake.PrependReactor("list", "pods", func(action testcore.Action) (bool, runtime.Object, error) {
		log.Info("returning fake pods")
		return true, podlist, nil
	})
	fake.PrependWatchReactor("pods", testcore.DefaultWatchReactor(watcher, nil))
	matcher := NewPodMatcher(fake, description)
	go matcher.Start(context.Background())
	// Delete pending pod
	watcher.Delete(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "test-ns",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	})

	select {
	case <-matcher.Done():
		t.Fatalf("test should not pass when pod is in Pending state")
	default:
	}

	watcher.Delete(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: "test-ns",
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionFalse,
				},
			},
		},
	})

	select {
	case <-matcher.Done():
		t.Fatalf("test should not pass when no pods are present")
	default:
	}

	watcher.Add(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-3",
			Namespace: "test-ns",
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionFalse,
				},
			},
		},
	})

	select {
	case <-matcher.Done():
		t.Fatalf("test should not pass when pod is in Pending state")
	default:
	}

	watcher.Modify(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-3",
			Namespace: "test-ns",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionFalse,
				},
			},
		},
	})

	select {
	case <-matcher.Done():
		t.Fatalf("test should not pass when pod is running but not ready")
	default:
	}

	watcher.Modify(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-3",
			Namespace: "test-ns",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	})

	timeoutDuration, _ := time.ParseDuration("500ms")
	select {
	case <-time.After(timeoutDuration):
		t.Fatalf("matcher did not return after 500ms")
	case <-matcher.Done():
	}
}

func TestPodValidator(t *testing.T) {
	validDescription := StateDescription{
		Type:           "Pod",
		Namespace:      "test-ns",
		LabelSelector:  "",
		RequiredStates: []ResourceState{ResourceReady, ResourceSucceeded, ResourceFailed},
	}

	validator := &PodValidator{}
	if err := validator.Validate(context.Background(), validDescription); err != nil {
		t.Fatal(err)
	}

	badDescription := StateDescription{
		Type:           "Pod",
		Namespace:      "test-ns",
		LabelSelector:  "",
		RequiredStates: []ResourceState{resourceWaiting, ResourceReady, ResourceSucceeded, ResourceFailed},
	}

	if err := validator.Validate(context.Background(), badDescription); err == nil || err.Error() != ErrWaitingStateReserved(badDescription).Error() {
		t.Fatalf("validation should fail with: %v , instead it failed with %v", ErrWaitingStateReserved(badDescription), err)
	}

	emptyRequiredStatesDescription := StateDescription{
		Type:           "Pod",
		Namespace:      "test-ns",
		LabelSelector:  "",
		RequiredStates: []ResourceState{},
	}

	if err := validator.Validate(context.Background(), emptyRequiredStatesDescription); err == nil || err.Error() != ErrNoRequiredStates(emptyRequiredStatesDescription).Error() {
		t.Fatalf("validation should fail with: %v , instead it failed with %v", ErrNoRequiredStates(badDescription), err)
	}
}
