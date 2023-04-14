package math

import (
	"fmt"
	"math"
	"strconv"

	"golang.org/x/exp/constraints"
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

// FormatInt formats the integer value as a string. The function accepts an optional argument of an
// integer value that will be used to inform the base of the integer. This argument, if not provided
// or invalid, will default to 10, meaning that the value will be formatted as a base-10 integer
func FormatInt[T constraints.Signed](value T, args ...any) string {

	// Since we have an integer, we'll first check and see if we have at least one argument. If
	// we do then we'll attempt to parse it is an integer and then use it as the base for the number
	// format. Otherwise, we'll just assume that the number is base 10
	base := 10
	if len(args) >= 1 {
		cBase, ok := args[0].(int)
		if ok && cBase > 0 {
			base = cBase
		}
	}

	// Format the integer as a string and return it
	return strconv.FormatInt(int64(value), base)
}

// FormatUint formats the unsigned integer value as a string. The function accepts an optional argument
// of an integer value that will be used to inform the base of the usigned integer. This argument, if
// not provided or invalid, will default to 10, meaning that the value will be formatted as a base-10 integer
func FormatUint[T constraints.Unsigned](value T, args ...any) string {

	// Since we have an integer, we'll first check and see if we have at least one argument. If
	// we do then we'll attempt to parse it is an integer and then use it as the base for the number
	// format. Otherwise, we'll just assume that the number is base 10
	base := 10
	if len(args) >= 1 {
		cBase, ok := args[0].(int)
		if ok && cBase > 0 {
			base = cBase
		}
	}

	// Format the unsigned integer as a string and return it
	return strconv.FormatUint(uint64(value), base)
}

// FormatFloat formats the floating-point value as a string. the function accepts up to two optional
// arguments: a byte that will be used to describe the format of the value which defaults to 'f' if
// no value is provided and an integer value describing the precision of the floating-point value which
// defaults to -1. The arguments must be provided in order, if provided.
func FormatFloat[T constraints.Float](value T, args ...any) string {

	// First, we'll attempt to parse the format parameter. If we have at least one argument, then
	// we'll check that the argument is a byte. Otherwise' we'll assume a format of 'f'
	fmt := byte('f')
	if len(args) >= 1 {
		cFmt, ok := args[0].(byte)
		if ok {
			fmt = cFmt
		}
	}

	// Next, we'll attempt to parse the precision. If we have at least two arguments, then we'll
	// check that the argument is an integer. Otherwise, we'll assume a precision of -1
	prec := -1
	if len(args) >= 2 {
		cPrec, ok := args[0].(int)
		if ok && cPrec >= -1 {
			prec = cPrec
		}
	}

	// Now, we'll attempt to parse the size of the number, based on the type of floating-point value
	// we receive. We'll check if the value can be asserted to a 32-bit floating-point value or assume
	// a 64-bit floating-point value otherwise
	size := 64
	if _, ok := any(value).(float32); ok {
		size = 32
	}

	// Finally, format the floating-point value to a string using the format, precision and size and return it
	return strconv.FormatFloat(float64(value), fmt, prec, size)
}
