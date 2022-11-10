package servicehelpers

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Create a new test runner we'll use to test all the
// modules in the service-helpers package
func TestServiceHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Helpers Suite")
}

var _ = Describe("Service Helper Tests", func() {

	// Tests that, if the user sends a SIGINT to the function after the
	// CancelOnInterrupt function is called, then the function will be cancelled
	It("CancelOnInterrupt - SIGINT received - Cancelled", func() {

		// First, create the context we'll use to interrupt
		ctx, awaiter := CancelOnInterrupt(context.Background())
		defer awaiter.Stop()

		// Next, setup a function that will update a variable when
		// the context is cancelled
		ctrl := new(sync.RWMutex)
		done := false
		go func() {
			<-ctx.Done()
			ctrl.Lock()
			defer ctrl.Unlock()
			done = true
		}()

		// Now, simulate cancellation by sending a SIGINT to the channel
		awaiter.signalChan <- os.Interrupt
		time.Sleep(20 * time.Millisecond)

		// Finally, verify that the variable was updated correctly
		ctrl.RLock()
		defer ctrl.RUnlock()
		Expect(done).Should(BeTrue())
	})

	// Tests that, if the cancellation function returned by the CancelOnInterrupt
	// function is called, then the function will be cancelled
	It("CancelOnInterrupt - Cancellation requested - Cancelled", func() {

		// First, create the context we'll use to interrupt
		ctx, cancel := context.WithCancel(context.Background())
		ctx, awaiter := CancelOnInterrupt(ctx)
		defer awaiter.Stop()

		// Next, setup a function that will update a variable when
		// the context is cancelled
		ctrl := new(sync.RWMutex)
		done := false
		go func() {
			<-ctx.Done()
			ctrl.Lock()
			defer ctrl.Unlock()
			done = true
		}()

		// Now, cancel the process
		cancel()
		time.Sleep(20 * time.Millisecond)

		// Finally, verify that the variable was updated correctly
		ctrl.RLock()
		defer ctrl.RUnlock()
		Expect(done).Should(BeTrue())
	})
})
