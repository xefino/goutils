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
})
