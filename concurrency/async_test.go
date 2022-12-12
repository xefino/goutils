package concurrency

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Asyncer Tests", func() {

	// Test that the Do/Await functions return an error if the inner function fails
	It("Do - Returns error - Await returns error", func() {

		// First, create the asyncer
		async := NewAsyncer[int]()

		// Next, attempt to start the asyncer with a function; this should fail
		async.Do(func() (int, error) {
			time.Sleep(10 * time.Millisecond)
			return 0, fmt.Errorf("failed")
		})

		// Now, wait for the operation to complete
		result, err := async.Await()

		// Finally, verify that we recevied the default value of the result and an error
		Expect(result).Should(Equal(0))
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("failed"))
	})

	// Test that the Do/Await function returns the result of the function if it does not fail
	It("Do - Returns value - Await returns value", func() {

		// First, create the asyncer
		async := NewAsyncer[int]()

		// Next, attempt to start the asyncer with a function; this should not fail
		async.Do(func() (int, error) {
			time.Sleep(10 * time.Millisecond)
			return 42, nil
		})

		// Now, wait for the operation to complete
		result, err := async.Await()

		// Finally, verify that we recevied the result of the function and that no error occurred
		Expect(result).Should(Equal(42))
		Expect(err).ShouldNot(HaveOccurred())
	})
})
