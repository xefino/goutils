package strings

import (
	"strings"

	"golang.org/x/exp/constraints"
)

// IsEmpty returns true if the string is nil or empty
// but will return false otherwise
func IsEmpty[S ~string](value S) bool {
	return value == ""
}

// ModifyAndJoin applies a modifier function to each string in the list submitted to the function
// and then joins them all together using separator
func ModifyAndJoin[S ~string](applier func(string) string, separator string, items ...S) string {
	applied := make([]string, len(items))
	for i, item := range items {
		applied[i] = applier(string(item))
	}

	return strings.Join(applied, separator)
}

// Quote creates a string from an inner string by adding quote strings before and after
func Quote[S ~string](inner S, quote string) string {
	return quote + string(inner) + quote
}

// Delimit creates a list of count values of s separated by count - 1 values of sep. Note that this
// function will panic if count is 0 or if the result of ((len(s) + len(sep)) * count) overflows.
func Delimit[C constraints.Integer](s string, sep string, count C) string {
	return strings.Repeat(s+sep, int(count)-1) + s
}
