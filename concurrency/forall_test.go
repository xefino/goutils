package concurrency

import (
	"context"
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ForAll Tests", func() {

	// Tests the conditions describing how the ForAllAsync function operates
	// under various short-circuiting and return conditions
	DescribeTable("ForAllAsync Tests",
		func(cancelRequested bool, throwsError bool, cancelOnErr bool) {

			// Create our test variables
			called := 0
			finished := 0
			cancelled := 0
			routine := createRuntime(throwsError, cancelRequested, &called, &cancelled, &finished)

			// Next, concurrently run the routine over a large number of iteration
			err := ForAllAsync(context.Background(), 1000, cancelOnErr, routine)

			// Now, verify the error if it was expected
			if throwsError {
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal("Error occurred"))
			} else {
				Expect(err).ShouldNot(HaveOccurred())
			}

			// Finally, verify the test variables
			if cancelRequested || cancelOnErr {
				Expect(finished).Should(BeNumerically(">=", 419))
				Expect(called).Should(Equal(finished + cancelled + 1))
			} else {
				Expect(called).Should(Equal(1000))
				Expect(cancelled).Should(BeZero())
				if throwsError {
					Expect(finished).Should(Equal(999))
				} else {
					Expect(finished).Should(Equal(1000))
				}
			}
		},
		Entry("Cancellation Not Requested - All run", false, false, false),
		Entry("Cancellation Requested - Not all run", true, false, false),
		Entry("Routine Returns Error, Cancel-on-error False - All run", false, true, false),
		Entry("Routine Returns Error, Cancel-on-error True - Not all run", false, true, true))
})

// Helper function that we'll use to create a test runtime we can use
// for testing the concurrency functions
func createRuntime(returnsError bool, hasCancel bool, called *int, cancelled *int,
	finished *int) func(context.Context, int, context.CancelFunc) error {

	// Create the mutex we'll use to control our read-write to avoid contamination
	mut := new(sync.Mutex)

	// Create the routine and return it
	return func(ctx context.Context, index int, cancel context.CancelFunc) error {

		// Update the total number of called functions
		mut.Lock()
		*called++
		mut.Unlock()

		// Check if we had cancellation requested; if we did then update
		// the cancellation total and return early
		select {
		case <-ctx.Done():
			mut.Lock()
			defer mut.Unlock()
			*cancelled++
			return nil
		default:
		}

		// If we want to test around error then throw an error partway through
		// Otherwise, if we want to test around cancellation then ensure that
		// the function is cancelled prematurely
		if returnsError && index == 420 {
			return fmt.Errorf("Error occurred")
		} else if hasCancel && index == 420 {
			cancel()
			return nil
		}

		// If we don't request cacnellation then update the counter
		mut.Lock()
		defer mut.Unlock()
		*finished++
		return nil
	}
}
