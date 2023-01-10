package random

import "math/rand"

// Common values to be used when creating random strings
var (
	Uppercase    = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	Lowercase    = []rune("abcdefghijklmnopqrstuvwxyz")
	Numbers      = []rune("0123456789")
	Alphanumeric = []rune(string(Uppercase) + string(Lowercase) + string(Numbers) + "_-~")
)

// RandomNRunes generates a string by choosing N runes at random from the runes list provided
func RandomNRunes(n int, runes []rune) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}

	return string(b)
}
