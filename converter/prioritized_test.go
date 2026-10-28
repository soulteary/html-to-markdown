package converter

import (
	"reflect"
	"sync"
	"testing"
)

func TestPrioritizedSlice(t *testing.T) {

	var values = prioritizedSlice[string]{
		prioritized("b", PriorityStandard),
		prioritized("c", PriorityLate),
		prioritized("a", PriorityEarly),
	}

	values.Sort()

	var expected = prioritizedSlice[string]{
		{
			Value:    "a",
			Priority: PriorityEarly,
		},
		{
			Value:    "b",
			Priority: PriorityStandard,
		},
		{
			Value:    "c",
			Priority: PriorityLate,
		},
	}

	if !reflect.DeepEqual(values, expected) {
		t.Errorf("expected %+v but got %+v", expected, values)
	}
}

func TestHandlerSnapshotSurvivesRegistration(t *testing.T) {
	var m sync.RWMutex
	var handlers handlerList[string]
	handlers.add(prioritized("late", PriorityLate))
	handlers.add(prioritized("standard", PriorityStandard))
	previous := sortedHandlers(&m, &handlers)
	handlers.add(prioritized("early", PriorityEarly))
	current := sortedHandlers(&m, &handlers)
	if len(previous) != 2 || previous[0].Value != "standard" || previous[1].Value != "late" {
		t.Fatalf("registration changed an active snapshot: %+v", previous)
	}
	if len(current) != 3 || current[0].Value != "early" || current[1].Value != "standard" || current[2].Value != "late" {
		t.Fatalf("new snapshot has incorrect priorities: %+v", current)
	}
}
