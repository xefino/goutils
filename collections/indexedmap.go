package collections

import "sync"

// IndexedMap is a list structure with O(1) lookup for a field that is
// indexed based on a key value
type IndexedMap[U comparable, T any] struct {
	indexes map[U]int
	data    []T
	ctrl    *sync.RWMutex
}

// NewIndexedMap creates a new, empty IndexedMap
func NewIndexedMap[U comparable, T any]() *IndexedMap[U, T] {
	return &IndexedMap[U, T]{
		indexes: make(map[U]int),
		data:    make([]T, 0),
		ctrl:    new(sync.RWMutex),
	}
}

// Add adds a new key and value to the indexed map. If the overwrite
// function is true then, if a collision occurs, the item being added
// will take precedence over the existing item. Otherwise, the existing
// item will take precedence. This operation may be O(1) unless the
// underlying list needs to be resized, in which case it will be O(N)
func (m *IndexedMap[U, T]) Add(key U, value T, overwrite bool) {
	m.AddIf(key, value, func(T, T) bool { return overwrite })
}

// AddIf adds a new key and value to the indexed map. If a collision
// occurs, then the onCollision function will be called with the existing
// item and item to be added. If it returns true then the existing item
// will be overwritten and otherwise it will be ignored.
func (m *IndexedMap[U, T]) AddIf(key U, value T, onCollision func(T, T) bool) {
	m.ctrl.Lock()
	defer m.ctrl.Unlock()
	if index, ok := m.indexes[key]; ok && onCollision(m.data[index], value) {
		m.data[index] = value
	} else if !ok {
		m.indexes[key] = len(m.data)
		m.data = append(m.data, value)
	}
}

// Exists determines whether or not the key exists in the indexed map.
// This operation is guaranteed to be O(1).
func (m *IndexedMap[U, T]) Exists(key U) bool {
	m.ctrl.RLock()
	defer m.ctrl.RUnlock()
	_, ok := m.indexes[key]
	return ok
}

// At retrieves the item located at the index provided. This function will
// panic if the index is located outside the bounds of the list. If a negative
// value is provided then the item will be indexed from the end of the list
func (m *IndexedMap[U, T]) At(index int) T {
	m.ctrl.RLock()
	defer m.ctrl.RUnlock()
	if index < 0 {
		return m.data[len(m.data)+index]
	} else {
		return m.data[index]
	}
}

// Get retrieves the item from the indexed list associated with the
// key. True will be returned if the value was found, otherwise false
// will be returned. This operation is guaranteed O(1).
func (m *IndexedMap[U, T]) Get(key U) (T, bool) {
	m.ctrl.RLock()
	defer m.ctrl.RUnlock()

	// First, attempt to get index associated with the key
	index, ok := m.indexes[key]

	// Next, if we don't have any index associated with the key then
	// create an empty item and return it, and false
	if !ok {
		var empty T
		return empty, false
	}

	// Finally, since we found the index then return the item
	// associated with the index and true
	return m.data[index], true
}

// Remove deletes the object associated with the key from the indexed
// map, if the key exists. A value of true will be returned if the key
// was actually associated with an item; otherwise false will be returned.
// This operation requires a resize of the array so it will run in O(M)
// where M is the number of elements from the index of the deleted entry
// to the end of the list.
func (m *IndexedMap[U, T]) Remove(key U) bool {
	m.ctrl.Lock()
	defer m.ctrl.Unlock()

	// First, attempt to get the index associated with the key; return
	// false if we don't have the key in our index
	index, ok := m.indexes[key]
	if !ok {
		return false
	}

	// Next, delete the key from the index and delete the item from
	// our list of data
	delete(m.indexes, key)
	m.data = append(m.data[:index], m.data[index+1:]...)

	// Now, iterate over all the values in the index and update the
	// index mapping to propagate the delete operation
	for key, i := range m.indexes {
		if i > index {
			m.indexes[key] = i - 1
		}
	}

	return true
}

// Keys returns the collection of keys associated with the indexed map
// as a slice, allowing for users to access the entire search space of
// the collection. This operation is guaranteed O(1).
func (m *IndexedMap[U, T]) Keys() []U {
	m.ctrl.RLock()
	defer m.ctrl.RUnlock()
	keys := make([]U, 0)
	for key := range m.indexes {
		keys = append(keys, key)
	}

	return keys
}

// Data returns the data associated with the indexed map as a slice,
// allowing for users to access the data that was being stored. This
// operation is guaranteed O(1).
func (m *IndexedMap[U, T]) Data() []T {
	m.ctrl.RLock()
	defer m.ctrl.RUnlock()
	return m.data
}

// Length returns the number of elements in the indexed map. This
// operation is guaranteed O(1).
func (m *IndexedMap[U, T]) Length() int {
	m.ctrl.RLock()
	defer m.ctrl.RUnlock()
	return len(m.data)
}

// ForEach iterates over the entire indexed map and calls a function
// for each key and associated value. If the function returns true,
// the iteration will continue; otherwise, it will not. This operation
// will be O(N) if the inner function does not return false.
func (m *IndexedMap[U, T]) ForEach(loopFunc func(U, T) bool) {
	m.ctrl.RLock()
	defer m.ctrl.RUnlock()
	for key, i := range m.indexes {
		if !loopFunc(key, m.data[i]) {
			return
		}
	}
}
