package math

import "golang.org/x/exp/constraints"

// EqualsAny returns true if lhs is equal to any item in others
func EqualsAny[T comparable](lhs T, others ...T) bool {
	for _, other := range others {
		if lhs == other {
			return true
		}
	}

	return false
}

// Max calculates the maximum of a list of ordered items, returning it.
// Note that, if this function is called with no arguments, then it will
// panic as it depends on at least one item being in the slice
func Max[T constraints.Ordered](items ...T) T {
	max := items[0]
	for i := 1; i < len(items); i++ {
		if items[i] > max {
			max = items[i]
		}
	}

	return max
}

// Min calculates the minimum of a list of ordered items, returning it.
// Note that, if this function is called with no arguments, then it will
// panic as it depends on at least one item being in the slice
func Min[T constraints.Ordered](items ...T) T {
	min := items[0]
	for i := 1; i < len(items); i++ {
		if items[i] < min {
			min = items[i]
		}
	}

	return min
}
