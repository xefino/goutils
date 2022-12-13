package time

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Time Formatting Tests", func() {

	// Test that the UnixNanoString works when the second is in decimals
	// up to the nanoseconds place
	It("UnixNanoString - Works", func() {

		// Create our test time
		time := time.Date(2022, time.June, 17, 23, 59, 59, 900838091, time.UTC)

		// Attempt to convert the time to a nanosecond timestamp
		timestamp := UnixNanoString(time)

		// Verify the timestamp
		Expect(timestamp).Should(Equal("1655510399900838091"))
	})

	// Tests that if there are no nanoseconds on the timestamp, then calling
	// the UnixNanoString will return a nanosecond timestamp
	It("UnixNanoString - Nanoseconds zero - Works", func() {

		// Create our test time
		time := time.Date(2022, time.June, 17, 23, 59, 59, 0, time.UTC)

		// Attempt to convert the time to a nanosecond timestamp
		timestamp := UnixNanoString(time)

		// Verify the timestamp
		Expect(timestamp).Should(Equal("1655510399000000000"))
	})

	// Test that the Epoch works when the second in in decimals up to the
	// nanoseconds place
	It("Epoch - Works", func() {

		// Create our new test duration
		duration := 23*time.Hour + 14*time.Second + 179*time.Millisecond

		// Attempt to convert the duration to a nanosecond timestamp
		timestamp := Epoch(duration)

		// Verify the timestamp
		Expect(timestamp).Should(Equal("82814179000000"))
	})

	// Tests that if there are no nanoseconds on the duration, then calling
	// the Epoch function will return a nanosecond timestamp
	It("Epoch - Nanoseconds zero - Works", func() {

		// Create our new test duration
		duration := 23*time.Hour + 14*time.Second

		// Attempt to convert the duration to a nanosecond timestamp
		timestamp := Epoch(duration)

		// Verify the timestamp
		Expect(timestamp).Should(Equal("82814000000000"))
	})

	// Tests that the Date function works to convert a time object to a string
	// representing the date in a YYYY-MM-DD format
	It("Date - Works", func() {

		// Create our test time
		time := time.Date(2022, time.June, 17, 23, 59, 59, 0, time.UTC)

		// Attempt to convert the time to a date string
		str := Date(time)

		// Verify the date string
		Expect(str).Should(Equal("2022-06-17"))
	})

})
