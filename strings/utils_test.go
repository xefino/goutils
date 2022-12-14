package strings

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Create a new test runner we'll use to test all the
// modules in the strings package
func TestStrings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Strings Suite")
}

var _ = Describe("Strings Tests", func() {

	// Tests that the IsEmpty function works when the string is empty
	DescribeTable("IsEmpty - Conditions",
		func(value string, expected bool) {

			// Check whether or not the value is empty
			result := IsEmpty(value)

			// Verify the result
			Expect(result).Should(Equal(expected))
		},
		EntryDescription("IsEmpty(%s) == %b?"),
		Entry("Empty String - True", "", true),
		Entry("White-space String - False", "\t\n    ", false),
		Entry("Non-empty String - False", "derp", false))

	// Tests that IsEmpty works if the value inherits from a string
	It("IsEmpty - Value inherits from string - Works", func() {

		// Create a fake type that we'll use
		type fakeString string

		// Create a new instance of our fake string type
		f := fakeString("derp")

		// Check whether or not the value is empty; this should return false
		result := IsEmpty(f)

		// Verify the result
		Expect(result).Should(BeFalse())
	})

	// Test that, if an empty list is provided to ModifyAndJoin, then an empty string will be returned
	It("ModifyAndJoin - No entries - Empty string returned", func() {
		result := ModifyAndJoin[string](func(raw string) string { return "A" + raw + "A" }, ",")
		Expect(result).Should(BeEmpty())
	})

	// Test that, if a non-empty list is provided to ModifyAndJoin, then each element will be modified and
	// the resulting list will be joined together using the separator
	It("ModifyAndJoin - Contains elements - Modified string returned", func() {
		result := ModifyAndJoin(func(raw string) string { return "A" + raw + "A" }, ",", "herp", "derp")
		Expect(result).Should(Equal("AherpA,AderpA"))
	})

	// Test that Quote works as expected
	It("Quote - Works", func() {
		result := Quote("A", "derp")
		Expect(result).Should(Equal("derpAderp"))
	})

	// Test that, if Delimit is called with a count argument of 0, then the function will panic
	It("Delimit - Count is 0 - Panic", func() {
		Expect(func() {
			Delimit("?", ", ", 0)
		}).Should(Panic())
	})

	// Tests the conditions determining how the Delimit function should operate
	DescribeTable("Delimit - Conditions",
		func(s string, sep string, count int, result string) {
			Expect(Delimit(s, sep, uint(count))).Should(Equal(result))
		},
		Entry("s is empty - Works", "", ", ", 2, ", "),
		Entry("sep is empty - Works", "d", "", 2, "dd"),
		Entry("count is 1 - Works", "?", ", ", 1, "?"),
		Entry("s and sep not empty, count > 1 - Works", "?", ", ", 3, "?, ?, ?"))
})
