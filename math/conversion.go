package math

import (
	"fmt"
	"math"
)

// VerifyInteger converts a floating-point number to an integer and verifies that
// no information will be lost in doing so. If this is not the case then an error
// will be returned
func VerifyInteger(decimal float64) (int64, error) {

	// First, check if the decimal value is out of bounds for an integer
	// If this is the case then the value cannot be contained in a normal
	// integer so return an error
	if decimal > math.MaxInt64 || decimal < math.MinInt64 {
		return -1, fmt.Errorf("decimal value of %f was outside the bounds of an int64", decimal)
	}

	// Next, extract the whole number and fraction from the decimal value
	whole, frac := math.Modf(decimal)

	// Now, check if the fractional value is greater than the smallest value
	// we can reasonably track so we can ensure that we don't have any false
	// positives resulting from floating point precision issues
	epsilon := 1e-9
	if !(frac < epsilon || frac > 1.0-epsilon) {
		return -1, fmt.Errorf("decimal value of %f contains non-zero fraction, %f", decimal, frac)
	}

	// Finally, convert the integer part to an int and return it
	return int64(whole), nil
}
