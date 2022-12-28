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
		Expect(async.Received()).Should(BeFalse())
		result, err, rec := async.Await()

		// Finally, verify that we recevied the default value of the result and an error
		Expect(async.Received()).Should(BeTrue())
		Expect(rec).Should(BeTrue())
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
		Expect(async.Received()).Should(BeFalse())
		result, err, rec := async.Await()

		// Finally, verify that we recevied the result of the function and that no error occurred
		Expect(async.Received()).Should(BeTrue())
		Expect(rec).Should(BeTrue())
		Expect(result).Should(Equal(42))
		Expect(err).ShouldNot(HaveOccurred())
	})

	// Test that, if the Await function is called twice, then only the first call will return true
	It("Do - Await requested twice - Receive occurs once", func() {

		// First, create the asyncer and attempt to start the asyncer with a function; this should not fail
		async := NewAsyncer[int]()
		async.Do(func() (int, error) {
			time.Sleep(10 * time.Millisecond)
			return 42, nil
		})

		// Next, wait for the operation to complete
		Expect(async.Received()).Should(BeFalse())
		result, err, rec := async.Await()

		// Now, verify that we recevied the result of the function and that no error occurred
		Expect(async.Received()).Should(BeTrue())
		Expect(rec).Should(BeTrue())
		Expect(result).Should(Equal(42))
		Expect(err).ShouldNot(HaveOccurred())

		// Finally, attempt to receive from the Await function again; rec should be false now
		_, _, rec = async.Await()
		Expect(rec).Should(BeFalse())
	})
})
