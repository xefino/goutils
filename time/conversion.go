package time

import (
	"fmt"
	"time"

	"github.com/xefino/quantum-api-go/data"
)

// StartOfHour trims everything after the hour from the provided time
func StartOfHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

// StartOfDay trims everything after the day from the provided time
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// TimeframeAdder generates an adder that can be used to get the next time from the current time
// based on the timeframe provided. This function will panic if the frequency is not one we recognize
// or the multiplier is zero. Note that this function operates differently on times and dates. For times,
// this function relies on time.Add and for dates it relies on time.AddDate. Note that this is done
// to ensure that addition works with respect to unevenly-spaced durations (months, quarters, years)
// as well as evenly-spaced durations (seconds, minutes, hours, days).
func TimeframeAdder(multiplier int, freq data.Frequency) func(time.Time) time.Time {

	// First, check if the resolution was zero. If it was then panic
	if multiplier == 0 {
		panic("resolution was zero")
	}

	// Next, for small frequencies, we'll generate a duration that we can add to the time. Otherwise,
	// for large frequencies, we'll return a duration function that adds dates directly.
	var dur time.Duration
	switch freq {
	case data.Frequency_Second:
		dur = time.Duration(multiplier) * time.Second
	case data.Frequency_Minute:
		dur = time.Duration(multiplier) * time.Minute
	case data.Frequency_Hour:
		dur = time.Duration(multiplier) * time.Hour
	case data.Frequency_Day:
		return func(start time.Time) time.Time {
			return start.AddDate(0, 0, int(multiplier))
		}
	case data.Frequency_Week:
		return func(start time.Time) time.Time {
			return start.AddDate(0, 0, int(7*multiplier))
		}
	case data.Frequency_Month:
		return func(start time.Time) time.Time {
			return start.AddDate(0, int(multiplier), 0)
		}
	case data.Frequency_Quarter:
		return func(start time.Time) time.Time {
			return start.AddDate(0, int(3*multiplier), 0)
		}
	case data.Frequency_Year:
		return func(start time.Time) time.Time {
			return start.AddDate(int(multiplier), 0, 0)
		}
	default:
		panic(fmt.Sprintf("frequency %s was not expected", freq))
	}

	// Finally, since the frequency was small, we'll return a function that adds the duration to the
	// time provided, producing another time
	return func(start time.Time) time.Time {
		return start.Add(dur)
	}
}
