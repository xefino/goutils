package concurrency

import "sync"

// Asyncer allows for an operation to be done asynchronously with the control flow and then
// awaited at the caller's convenience
type Asyncer[T any] struct {
	results chan T
	errors  chan error
	ret     *sync.Once
	recMut  *sync.Mutex
	rec     bool
}

// NewAsyncer creates a new Asyncer
func NewAsyncer[T any]() *Asyncer[T] {
	return &Asyncer[T]{
		results: make(chan T, 1),
		errors:  make(chan error, 1),
		ret:     new(sync.Once),
		recMut:  new(sync.Mutex),
		rec:     false,
	}
}

// Do runs the given function asynchronously, sending output and errors to the channels provided
func (a *Asyncer[T]) Do(action func() (T, error)) {
	go func() {

		// Run the function; if this fails then push the error to the errors channel. Otherwise,
		// push the result to the results channel
		result, err := action()
		if err != nil {
			a.errors <- err
		} else {
			a.results <- result
		}
	}()
}

// Received returns a Boolean value showing whether or not a value has been received from the asynchronous
// function at the time it is called.
func (a *Asyncer[T]) Received() bool {
	return a.rec
}

// Await blocks execution until something is received on either the errors channel or the results channel
// and then returns that value to the caller
func (a *Asyncer[T]) Await() (T, error, bool) {
	var result T
	var err error
	var rec bool
	a.ret.Do(func() {
		a.recMut.Lock()
		a.rec = true
		a.recMut.Unlock()
		rec = true
		select {
		case result = <-a.results:
		case err = <-a.errors:
		}
	})

	return result, err, rec
}
