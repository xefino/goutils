package math

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Create a new test runner we'll use to test all the
// modules in the math package
func TestMath(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Math Suite")
}
