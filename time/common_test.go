package time

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Create a new test runner we'll use to test all the
// modules in the time package
func TestTime(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Time Suite")
}

var _ = Describe("Common Tests", func() {

	// Tests that the Today function returns the a time corresponding
	// to the current day
	It("Today - Works", func() {

		// First, get the time for today
		today := Today()

		// Next, get the current time
		now := time.Now()

		// Finally, verify that the time is correct
		Expect(today.Year()).To(Equal(now.Year()))
		Expect(today.Month()).To(Equal(now.Month()))
		Expect(today.Day()).To(Equal(now.Day()))
		Expect(today.Hour()).To(BeZero())
		Expect(today.Minute()).To(BeZero())
		Expect(today.Second()).To(BeZero())
		Expect(today.Nanosecond()).To(BeZero())
		Expect(today.Location()).To(Equal(time.UTC))
	})
})
