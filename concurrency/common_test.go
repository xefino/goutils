package concurrency

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Create a new test runner we'll use to test all the
// modules in the concurrency package
func TestConcurrency(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Concurrency Suite")
}
