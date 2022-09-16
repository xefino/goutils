package collections

import (
	"sync"
)

// ConcurrentList is a list structure that is thread-safe
type ConcurrentList[T any] struct {
	data   []T
	resize bool
	ctrl   *sync.RWMutex
}

// NewConcurrentList creates a new concurrent list from an existing
// varadabatic list
func NewConcurrentList[T any](items ...T) *ConcurrentList[T] {
	return &ConcurrentList[T]{
		data: items,
		ctrl: new(sync.RWMutex),
	}
}

// WithResize modifies the list, allowing it to dynamically resize in cases
// where the capacity is larger than the length (this will likely only be
// useful in cases of extremely large datasets). This function returns the
// modified list object so that it can be chained with other modifiers
func (list *ConcurrentList[T]) WithResize() *ConcurrentList[T] {
	list.ctrl.Lock()
	defer list.ctrl.Unlock()
	list.resize = true
	return list
}

// Length returns the number of items in the list
func (list *ConcurrentList[T]) Length() uint {
	list.ctrl.RLock()
	defer list.ctrl.RUnlock()
	return uint(len(list.data))
}

// At returns the item at the given index
func (list *ConcurrentList[T]) At(index uint) T {
	list.ctrl.RLock()
	defer list.ctrl.RUnlock()
	return list.data[index]
}

// From returns the tail of the list from the given index
func (list *ConcurrentList[T]) From(index uint) []T {
	list.ctrl.RLock()
	defer list.ctrl.RUnlock()
	return list.data[index:]
}

// To returns the head of the list to the given index
func (list *ConcurrentList[T]) To(index uint) []T {
	list.ctrl.RLock()
	defer list.ctrl.RUnlock()
	return list.data[:index]
}

// Slice returns the part of the list from the start index
// to the end index
func (list *ConcurrentList[T]) Slice(start uint, end uint) []T {
	list.ctrl.RLock()
	defer list.ctrl.RUnlock()
	return list.data[start:end]
}

// Append adds one or more items to the end of the list
func (list *ConcurrentList[T]) Append(items ...T) {
	list.ctrl.Lock()
	defer list.ctrl.Unlock()
	list.data = append(list.data, items...)
}

// Clear removes all data from the list and returns it
func (list *ConcurrentList[T]) Clear() []T {
	list.ctrl.Lock()
	defer list.ctrl.Unlock()
	data := list.data
	list.data = make([]T, 0)
	return data
}

// RemoveAt removes the item at the given index from the list
// and returns it
func (list *ConcurrentList[T]) RemoveAt(index uint) T {
	list.ctrl.Lock()
	defer list.ctrl.Unlock()
	item := list.data[index]
	list.data = append(list.data[:index], list.data[index+1:]...)
	list.doResize()
	return item
}

// PopFront removes the first N items from the front of the list
// and returns them
func (list *ConcurrentList[T]) PopFront(n uint) []T {
	list.ctrl.Lock()
	defer list.ctrl.Unlock()
	items := list.data[:n]
	list.data = list.data[n:]
	list.doResize()
	return items
}

// Helper function that resizes the data list if its capacity has
// grown too large to preserve memory in the case of large lists
func (list *ConcurrentList[T]) doResize() {
	if list.resize && cap(list.data)>>1 > len(list.data) {
		temp := make([]T, len(list.data))
		copy(temp, list.data)
		list.data = temp
	}
}
