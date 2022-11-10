package servicehelpers

import (
	"context"
	"os"
	"os/signal"
)

// Awaiter allows for the encapsulation of the data necessary to
// handle user-generated cancellations
type Awaiter struct {
	signalChan chan os.Signal
}

// Stop informs the OS that we no longer want to monitor the signal channel
func (a *Awaiter) Stop() {
	signal.Stop(a.signalChan)
}

// CancelOnInterrupt ensures that the context cancels when the user-intercept
// signal is received from the standard input
func CancelOnInterrupt(ctx context.Context) (context.Context, *Awaiter) {

	// First, create a context and cancel function we'll use to interrupt when
	// the process is cancelled by the user
	ctx, cancel := context.WithCancel(ctx)

	// Next, create a channel we'll use to cacel the function when we receive
	// a Control-C from the standard input
	awaiter := Awaiter{signalChan: make(chan os.Signal, 1)}
	signal.Notify(awaiter.signalChan, os.Interrupt, os.Kill)

	// Finally, spin up a separate process that will wait for messages on the
	// channel and cancel if the signal notifies; also quit if we receive a message
	// that the context was cancelled somewhere else
	go func() {
		select {
		case <-awaiter.signalChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx, &awaiter
}
