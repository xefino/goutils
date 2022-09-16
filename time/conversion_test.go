package time

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Time Conversion Tests", func() {

	// Tests that the StartOfHour function works as expected
	It("StartOfHour - Works", func() {
		t := time.Date(2022, time.August, 1, 11, 24, 30, 0, time.UTC)
		sh := StartOfHour(t)
		Expect(sh.Format(time.RFC3339)).Should(Equal("2022-08-01T11:00:00Z"))
	})

	// Tests that the StartOfDay function works as expected
	It("StartOfDay - Works", func() {
		t := time.Date(2022, time.August, 1, 11, 24, 30, 0, time.UTC)
		sd := StartOfDay(t)
		Expect(sd.Format(time.RFC3339)).Should(Equal("2022-08-01T00:00:00Z"))
	})
})
