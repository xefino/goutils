package concurrency

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Buffer Tests", func() {

	// Test that the Size function works as expected
	It("Size - Works", func() {
		buf := NewBuffer[int](10)
		Expect(buf.Size()).Should(Equal(10))
	})

	// Test that calling Get when no data exists in the buffer will return false and nil
	It("Get - No data - Nil returned", func() {

		// First, create our test buffer
		buf := NewBuffer[int](10)

		// Next, attempt to get data from our buffer
		data, ok := buf.Get()

		// Finally, verify that we got not data
		Expect(data).Should(BeNil())
		Expect(ok).Should(BeFalse())
	})

	// Test that, if Get is called before the limit is reached then data will be returned, but if
	// it is called after the limit is reached the function will block until Release is called
	It("Get - Over size - Blocked", func() {

		// First, create our test buffer
		buf := NewBuffer[int](10)
		buf.Load(
			generateBufferItem(1), generateBufferItem(2),
			generateBufferItem(42), generateBufferItem(100),
			generateBufferItem(99), generateBufferItem(69),
			generateBufferItem(420), generateBufferItem(1000),
			generateBufferItem(9001), generateBufferItem(-1),
			generateBufferItem(-2), generateBufferItem(4096))

		// Next, iterate over the first 10 items and attempt to get each
		// without releasing the buffer
		expected := []int{1, 2, 42, 100, 99, 69, 420, 1000, 9001, -1}
		for i := 0; i < 10; i++ {
			data, ok := buf.Get()
			Expect(*data).Should(Equal(expected[i]))
			Expect(ok).Should(BeTrue())
		}

		// Now, attempt to get one more item; this should block
		mut := new(sync.RWMutex)
		retrieved := false
		go func() {
			data, ok := buf.Get()
			mut.Lock()
			retrieved = true
			mut.Unlock()
			Expect(*data).Should(Equal(-2))
			Expect(ok).Should(BeTrue())
		}()

		// Verify that the retrieval was blocked and then release the buffer
		mut.RLock()
		Expect(retrieved).Should(BeFalse())
		buf.Release()
		mut.RUnlock()

		// Finally, wait a bit and then verify that the block was released
		time.Sleep(1 * time.Millisecond)
		mut.RLock()
		Expect(retrieved).Should(BeTrue())
		mut.RUnlock()
	})
})

// Helper function that generates a pointer to our test value
func generateBufferItem(value int) *int {
	return &value
}
