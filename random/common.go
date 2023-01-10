package random

import (
	"math/rand"
	"time"
)

// Initialize the random number generator with the current time
func init() {
	rand.Seed(time.Now().UnixNano())
}
