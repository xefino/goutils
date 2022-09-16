package collections

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Concurrent List Tests", func() {

	// Tests that the Length function works as expected and
	// returns the number of elements in the list
	It("Length - Works", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Verify its length
		Expect(list.Length()).Should(Equal(uint(6)))
	})

	// Tests that attempting to access an item outside of the list
	// using the At function will cause a panic
	It("At - Index out of range - Panic", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Attempt to access an item outside of the list
		// This should panic
		Expect(func() {
			_ = list.At(100)
		}).Should(Panic())
	})

	// Tests that the At function works as expected and returns
	// the item at the index provided
	It("At - Works", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Iterate over the entire list and check that At
		// produces the correct object for each index provided
		for i, actual := range list.data {
			expected := list.At(uint(i))
			Expect(actual).Should(Equal(expected))
		}
	})

	// Tests that attempting to get a slice from an index outside
	// the bounds of the list by calling From will result in a panic
	It("From - Index out of range - Panic", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Attempt to access items outside the list
		// This should panic
		Expect(func() {
			_ = list.From(100)
		}).Should(Panic())
	})

	// Tests that attempting to get a slice from an index inside
	// the bounds of the list by calling From will return all the
	// items from the index until the end of the list
	It("From - Index inside list - Returned", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Attempt to access items from the 3rd index
		items := list.From(3)

		// Verify the data that was returned
		Expect(items).Should(HaveLen(3))
		Expect(items[0]).Should(Equal(5))
		Expect(items[1]).Should(Equal(8))
		Expect(items[2]).Should(Equal(10))
	})

	// Tests that attempting to get a slice from an index outside
	// the bounds of the list by calling To will result in a panic
	It("To - Index outside list - Panic", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Attempt to access items outside the list
		// This should panic
		Expect(func() {
			_ = list.To(100)
		}).Should(Panic())
	})

	// Tests that attempting to get a slice from an index inside
	// the bounds of the list by calling To will return all the
	// items from the beginning of the list until the index
	It("To - Index inside list - Returned", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Attempt to access items to the 3rd index
		items := list.To(3)

		// Verify the data that was returned
		Expect(items).Should(HaveLen(3))
		Expect(items[0]).Should(Equal(1))
		Expect(items[1]).Should(Equal(2))
		Expect(items[2]).Should(Equal(4))
	})

	// Tests the conditions under which the Slice function should panic
	DescribeTable("Slice - Panic Conditions",
		func(from int, to int) {

			// Create a new list
			list := NewConcurrentList(1, 2, 4, 5, 8, 10)

			// Attempt to access items improperly; this should fail
			Expect(func() {
				_ = list.Slice(uint(from), uint(to))
			}).Should(Panic())
		},
		Entry("Start index outside list", 100, 200),
		Entry("End index outside list", 1, 200),
		Entry("Start index greater than end index", 3, 1))

	// Tests that calling Slice with the same start and end
	// indices will return an empty slice
	It("Slice - Start index equals end index - Empty slice returned", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Attempt to get a slice from the list with the same
		// start and end indices; this should return an empty list
		results := list.Slice(3, 3)

		// Verify that the list is empty
		Expect(results).Should(BeEmpty())
	})

	// Tests that attempting to get a slice from start and end
	// indices that are within the bounds of the list by calling
	// Slice will return all the items from the start index to
	// the end index
	It("Slice - Works", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Attempt to get a slice from the list
		results := list.Slice(2, 4)

		// Verify the data that we got
		Expect(results).Should(HaveLen(2))
		Expect(results[0]).Should(Equal(4))
		Expect(results[1]).Should(Equal(5))
	})

	// Test that, if the Append function is called without any arguments
	// then nothing will be added to the list and no change will occur
	It("Append - Empty list provided - No change", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Append an empty list to the list
		list.Append()

		// Verify that the list is unchanged
		Expect(list.Length()).Should(Equal(uint(6)))
		Expect(list.data).Should(Equal([]int{1, 2, 4, 5, 8, 10}))
	})

	// Test that, if the Append function is called with a non-empty list
	// of items, then all the items will be added to the end of the list
	It("Append - Works", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Append some items to the list
		list.Append(42, 84, 100, 999, -1)

		// Verify that all the results were added to the list
		Expect(list.Length()).Should(Equal(uint(11)))
		Expect(list.data).Should(Equal([]int{1, 2, 4, 5, 8, 10, 42, 84, 100, 999, -1}))
	})

	// Verify that the Clear function works as expected, removing all the
	// data from the list and returning it from the function
	It("Clear - Works", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Clear the list
		data := list.Clear()

		// Verify that the data returned from the Clear operation
		// is equal to the data that was in the list and that the
		// list is now empty
		Expect(data).Should(HaveLen(6))
		Expect(data).Should(Equal([]int{1, 2, 4, 5, 8, 10}))
		Expect(list.data).Should(BeEmpty())
	})

	// Test that, if the RemoveAt function is called with an index that is
	// outside the bounds of the list, then the function will panic
	It("RemoveAt - Index outside list - Panic", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Attempt to access items outside the list
		// This should panic
		Expect(func() {
			_ = list.RemoveAt(100)
		}).Should(Panic())
	})

	// Test that, if the RemoveAt function is called with an index that is
	// within the bounds of the list, then the element at that index will
	// be removed from the list and returned
	It("RemoveAt - Index inside list - Returned", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Remove the item at the given index
		item := list.RemoveAt(3)

		// Verify that the item was removed from the list
		Expect(item).Should(Equal(5))
		Expect(list.Length()).Should(Equal(uint(5)))
		Expect(list.data).Should(Equal([]int{1, 2, 4, 8, 10}))
	})

	// Test that, if the PopFront function is called with an index that is
	// outside the bounds of the list, then the function will panic
	It("PopFront - Index outside list - Panic", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Attempt to access items outside the list
		// This should panic
		Expect(func() {
			_ = list.PopFront(100)
		}).Should(Panic())
	})

	// Test that, if the PopFront function is called with an index that is
	// inside the bounds of the list, then the items forward of the index will
	// be removed from the list and returned as a slice
	It("PopFront - Index inside list - Returned", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10)

		// Remove the first N items from the list
		items := list.PopFront(3)

		// Verify that the data is removed from the list and returned
		Expect(list.Length()).Should(Equal(uint(3)))
		Expect(list.data).Should(Equal([]int{5, 8, 10}))
		Expect(items).Should(HaveLen(3))
		Expect(items).Should(Equal([]int{1, 2, 4}))
	})

	// Test that, if the list is resize-enabled and requesting that data be popped
	// returns more than half the list, then the capacity will be reduced
	It("PopFront - Resize-enabled, Resize requested - Resized", func() {

		// Create a new list
		list := NewConcurrentList(1, 2, 4, 5, 8, 10).WithResize()
		Expect(list.data).Should(HaveLen(6))
		Expect(list.data).Should(HaveCap(6))

		// Remove the first N items from the list
		items := list.PopFront(3)

		// Verify that the data is removed from the list and returned
		Expect(list.Length()).Should(Equal(uint(3)))
		Expect(list.data).Should(HaveLen(3))
		Expect(list.data).Should(HaveCap(3))
		Expect(list.data).Should(Equal([]int{5, 8, 10}))
		Expect(items).Should(HaveLen(3))
		Expect(items).Should(Equal([]int{1, 2, 4}))
	})
})
