package concurrency

import (
	"context"
	"sync"
)

// ForAllAsync runs the provided funcction concurrently for every entry
// in the length provided in such a way that cancellation will short-circuit
// execution of the remaining functions. The routine is provided with a
// context that will be notified when short-circuiting occurs and a cancallation
// function that can be used to short-circuit the operation. If cancelOnErr is
// set to true, then routines will be not run after one returns an error
func ForAllAsync(ctx context.Context, length int, cancelOnErr bool,
	routine func(context.Context, int, context.CancelFunc) error) error {

	// First, create an inner context and cancellation function
	// from the context we received with the function
	ctxInner, cancel := context.WithCancel(ctx)

	// Create the variables we'll need to manage the concurrency
	errs := make(chan error, length)
	wg := new(sync.WaitGroup)

	// Next, iterate over the length and run the routine for each
	// This loop may be broken out of prematurely if cancellation is
	// requested by the loop itself or any of the routines within
loop:
	for i := 0; i < length; i++ {

		// First, check whether cancellation has been requested
		// If this is the case then break out of the loop
		select {
		case <-ctxInner.Done():
			break loop
		default:
		}

		// Next, create a wrapper for our routine and start it
		// so we can run them all concurrently
		wg.Add(1)
		go func(ctx context.Context, index int, cancel context.CancelFunc) {
			defer wg.Done()

			// First, check whether cancellation has been requested
			// If this is the case then break out of the function
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Next, run the routine with the context and cancellation
			// function; if this returns an error then pass that error
			// to the channel so we can record them. If short-circuiting
			// is desired in this case then request it here
			if err := routine(ctx, index, cancel); err != nil {
				errs <- err
				if cancelOnErr {
					cancel()
				}
			}
		}(ctxInner, i, cancel)
	}

	// Now, wait for all the routines to finish and ensure that
	// the context is relinquished from memory properly
	wg.Wait()
	defer cancel()

	// Finally, check if we received an error. If we did then return
	// it; otherwise, return nil
	select {
	case err, _ := <-errs:
		return err
	default:
		return nil
	}
}
