package converter

import (
	"sort"
	"sync"
)

const (
	// PriorityEarly means that the handler will be run **early** in the process.
	// To run it even earlier you need to subtract from this number.
	PriorityEarly = 100

	// PriorityStandard is for handlers that don't need to be run in a particular order.
	PriorityStandard = 500

	// PriorityLate means that the handler will be run **late** in the process.
	// To run it even later you need to add to this number.
	PriorityLate = 1000
)

type prioritizedValue[V any] struct {
	Value    V
	Priority int
}

type prioritizedSlice[V any] []prioritizedValue[V]

func (s prioritizedSlice[V]) Sort() {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Priority < s[j].Priority
	})
}

func prioritized[V any](v V, priority int) prioritizedValue[V] {
	return prioritizedValue[V]{v, priority}
}

// handlerList retains registration order separately from its immutable sorted
// snapshot. All access is protected by the owning Converter's mutex.
// Keeping registration order preserves the existing equal-priority sorting.
type handlerList[V any] struct {
	registered prioritizedSlice[V]
	sorted     prioritizedSlice[V]
	ready      bool
}

func (h *handlerList[V]) add(value prioritizedValue[V]) {
	h.registered = append(h.registered, value)
	h.sorted = nil
	h.ready = false
}

func sortedHandlers[V any](m *sync.RWMutex, h *handlerList[V]) prioritizedSlice[V] {
	m.RLock()
	if h.ready {
		result := h.sorted
		m.RUnlock()
		return result
	}
	m.RUnlock()

	m.Lock()
	defer m.Unlock()
	if !h.ready {
		h.sorted = append(prioritizedSlice[V](nil), h.registered...)
		h.sorted.Sort()
		h.ready = true
	}
	return h.sorted
}
