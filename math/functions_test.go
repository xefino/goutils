package math

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Equality Tests", func() {

	// Test that the EqualsAny function will return false if no comparison
	// arguments are provided to the function
	It("EqualsAny - List empty - False", func() {

		// Try to call EqualsAny with one argument; this should return false
		result := EqualsAny(42)

		// Verify the result
		Expect(result).Should(BeFalse())
	})

	// Test that the EqualsAny function will return false if none of the
	// comparison arguments are equal to the first argument
	It("EqualsAny - None are equal - False", func() {

		// Try to call EqualsAny with more than one argument, none of which
		// are equal to the first argument; this should return false
		result := EqualsAny(42, 24, 99, 87, 100)

		// Verify the result
		Expect(result).Should(BeFalse())
	})

	// Tests that the EqualsAny function will return true if one or more of
	// the comparison arguments are equal to the first argument
	It("EqualsAny - Multiple equal - True", func() {

		// Try to call EqualsAny with more than one argument, more than one
		// of which are equal to the first argument; this should return true
		result := EqualsAny(42, 24, 99, 87, 42, 42, 100, 43, 41, 24, 42)

		// Verify the result
		Expect(result).Should(BeTrue())
	})

	// Tests that, if the Max function is called with no arguments, then it will panic
	It("Max - No arguments - Panic", func() {
		Expect(func() {
			_ = Max[int]()
		}).Should(Panic())
	})

	// Tests that, if Max is provided only one argument, then that argument will be returned
	It("Max - One argument - Returned", func() {
		max := Max(42)
		Expect(max).Should(Equal(42))
	})

	// Tests that, if only two arguments are provided, Max will return the maximum
	DescribeTable("Max - Two arguments - Maximum returned",
		func(arg1 float64, arg2 float64, result float64) {
			max := Max(arg1, arg2)
			Expect(max).Should(Equal(result))
		},
		Entry("First < Second - Second returned", 42.0, 100.98, 100.98),
		Entry("First > Second - First returned", -23.99, -1000.55, -23.99),
		Entry("First = Second - First returned", 42.0, 42.0, 42.0))

	// Test that, if more than two arguments are provided, Max will return the largest of them
	DescribeTable("Max - More than two arguments - Maximum returned",
		func(result string, args ...string) {
			max := Max(args...)
			Expect(max).Should(Equal(result))
		},
		Entry("First largest - First returned", "first", "first", "Second", "Third"),
		Entry("Second largest - Second returned", "derp", "HERP", "derp", "3herbert"),
		Entry("Last largest - Last returned", "zebra", "one", "two", "three", "zebra"))

	// Tests that, if the Min function is called with no arguments, then it will panic
	It("Min - No arguments - Panic", func() {
		Expect(func() {
			_ = Min[int]()
		}).Should(Panic())
	})

	// Tests that, if Min is provided only one argument, then that argument will be returned
	It("Min - One argument - Returned", func() {
		min := Min(42)
		Expect(min).Should(Equal(42))
	})

	// Tests that, if only two arguments are provided, Min will return the minimum
	DescribeTable("Min - Two arguments - Minimum returned",
		func(arg1 float64, arg2 float64, result float64) {
			min := Min(arg1, arg2)
			Expect(min).Should(Equal(result))
		},
		Entry("First < Second - First returned", 42.0, 100.98, 42.0),
		Entry("First > Second - Second returned", -23.99, -1000.55, -1000.55),
		Entry("First = Second - First returned", 42.0, 42.0, 42.0))

	// Test that, if more than two arguments are provided, Min will return the smallest of them
	DescribeTable("Min - More than two arguments - Minimum returned",
		func(result string, args ...string) {
			min := Min(args...)
			Expect(min).Should(Equal(result))
		},
		Entry("First smallest - First returned", "First", "First", "second", "third"),
		Entry("Second smallest - Second returned", "Derp", "herp", "Derp", "sherbert"),
		Entry("Last smallest - Last returned", "Zebra", "one", "two", "three", "Zebra"))
})
