package strings

// IsEmpty returns true if the string is nil or empty
// but will return false otherwise
func IsEmpty[S ~string](value S) bool {
	return value == ""
}
