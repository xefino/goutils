package orm

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// Create a new test runner we'll use to test all the
// modules in the orm package
func TestORM(t *testing.T) {
	gomega.RegisterFailHandler(Fail)
	RunSpecs(t, "ORM Suite")
}

// Helper type that we'll use for returning data from SQL queries
type testType struct {
	Key   string `sql:"key"`
	Value string `sql:"value"`
}

// Helper function that verifies the fields on the test type
func verifyTestType(data *testType, key string, value string) {
	gomega.Expect(data.Key).Should(gomega.Equal(key))
	gomega.Expect(data.Value).Should(gomega.Equal(value))
}
