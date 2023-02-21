package lambda

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Create a new test runner we'll use to test all the
// modules in the lambda package
func TestLambda(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lambda Suite")
}
