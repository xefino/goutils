package time

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/xefino/quantum-api-go/data"
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

	// Tests that, if the resolution was zero, then calling TimeframeAdder will panic
	It("TimeframeAdder - Resolution zero - Panics", func() {
		Expect(func() {
			_ = TimeframeAdder(0, data.Frequency_Day)
		}).Should(Panic())
	})

	// Tests that, if the frequency was invalid, then calling TimeframeAdder will panic
	It("TimeframeAdder - Frequency invalid - Panics", func() {
		Expect(func() {
			_ = TimeframeAdder(1, data.Frequency_InvalidFrequency)
		}).Should(Panic())
	})

	// Tests the data conditions under which the TimeframeAdder works
	DescribeTable("TimeframeAdder - Frequency valid - Conditions",
		func(res int, freq data.Frequency, expected time.Time) {
			t := time.Date(2022, time.August, 1, 11, 24, 30, 0, time.UTC)
			Expect(TimeframeAdder(res, freq)(t)).Should(Equal(expected))
		},
		Entry("Timeframe is seconds - Works", 10, data.Frequency_Second,
			time.Date(2022, time.August, 1, 11, 24, 40, 0, time.UTC)),
		Entry("Timeframe is minutes - Works", 5, data.Frequency_Minute,
			time.Date(2022, time.August, 1, 11, 29, 30, 0, time.UTC)),
		Entry("Timeframe is hours - Works", 2, data.Frequency_Hour,
			time.Date(2022, time.August, 1, 13, 24, 30, 0, time.UTC)),
		Entry("Timeframe is days - Works", 3, data.Frequency_Day,
			time.Date(2022, time.August, 4, 11, 24, 30, 0, time.UTC)),
		Entry("Timeframe is weeks - Works", 1, data.Frequency_Week,
			time.Date(2022, time.August, 8, 11, 24, 30, 0, time.UTC)),
		Entry("Timeframe is months - Works", 2, data.Frequency_Month,
			time.Date(2022, time.October, 1, 11, 24, 30, 0, time.UTC)),
		Entry("Timeframe is quarters - Works", 2, data.Frequency_Quarter,
			time.Date(2023, time.February, 1, 11, 24, 30, 0, time.UTC)),
		Entry("Timeframe is years - Works", 5, data.Frequency_Year,
			time.Date(2027, time.August, 1, 11, 24, 30, 0, time.UTC)))
})
