package math

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Math Conversion Tests", func() {

	// Test the conditions under which the VerifyInteger function will return an error
	DescribeTable("VerifyInteger - Failures",
		func(raw float64, message string) {

			// Attempt to get an integer value from a floating-point
			// value; this should fail
			result, err := VerifyInteger(raw)

			// Verify the failure
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(message))
			Expect(result).Should(Equal(int64(-1)))
		},
		Entry("Value greater than MaxInt - Error", 2e22,
			"decimal value of 20000000000000000000000.000000 was outside the bounds of an int64"),
		Entry("Value less than MinInt - Error", -2e22,
			"decimal value of -20000000000000000000000.000000 was outside the bounds of an int64"),
		Entry("Value is decimal - Error", 42.09,
			"decimal value of 42.090000 contains non-zero fraction, 0.090000"))

	// Test that, if the VerifyInteger does not return an error, then an integer value
	// will be produced from the raw value
	It("VerifyInteger - Works", func() {

		// Attempt to get an integer value from a floating-point
		// value; this should not fail
		result, err := VerifyInteger(3.5812e6)

		// Verify the result
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).Should(Equal(int64(3581200)))
	})
})
