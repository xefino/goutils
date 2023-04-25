package orm

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/xefino/goutils/testutils"
	"github.com/xefino/goutils/utils"
)

// Create a new test runner we'll use to test all the
// modules in the orm package
func TestORM(t *testing.T) {
	gomega.RegisterFailHandler(Fail)
	RunSpecs(t, "ORM Suite")
}

var _ = Describe("Common Tests", func() {

	// Tests the conditions under which the Run function will return an error
	DescribeTable("RunQuery - Failures",
		func(queryFails bool, scanFails bool, verifier func(*utils.GError)) {

			// First, create the logger and discard any messages directed to the standard output
			logger := utils.NewLogger("query", "test")
			logger.Discard()

			// Next, create our mock database; this should not fail
			db, mock, err := sqlmock.New()
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

			// Tell the mock which queries we expect and what should be returned from them
			queryStmt := mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM test_table"))
			if queryFails {
				queryStmt.WillReturnError(fmt.Errorf("QueryContext failed"))
			} else {

				// If we want to return data then create it here
				rows := sqlmock.NewRows([]string{"key", "value"}).
					AddRow("key1", "value1").AddRow("key2", "value2")
				if scanFails {
					rows.RowError(0, fmt.Errorf("Scan failed"))
				}

				// Modify the query statement to return our test data
				queryStmt.WillReturnRows(rows)
			}

			// Now, create and run a new test query from our test table; this should return an error
			data, err := RunQuery[testType](context.Background(), NewQuery().From("test_table"), db, logger)

			// Finally, verify the error, that we received no data, and that our expectations were met
			verifier(err.(*utils.GError))
			gomega.Expect(data).Should(gomega.BeEmpty())
			gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
		},
		Entry("QueryContext fails - Error", true, false, testutils.ErrorVerifier("test", "orm",
			"/goutils/sql/orm/common.go", "", "RunQuery", 32, testutils.InnerErrorVerifier("QueryContext failed"),
			"Failed to query orm.testType data from \"test_table\"", "[test] orm.RunQuery "+
				"(/goutils/sql/orm/common.go 32): Failed to query orm.testType data from \"test_table\", "+
				"Inner:\n\tQueryContext failed.")),
		Entry("ReadRows fails - Error", false, true, testutils.ErrorVerifier("test", "orm",
			"/goutils/sql/orm/common.go", "", "RunQuery", 39, testutils.InnerErrorVerifier("Row could not be read, error: Scan failed"),
			"Failed to read orm.testType data returned from \"test_table\"", "[test] orm.RunQuery (/goutils/sql/orm/common.go 39): "+
				"Failed to read orm.testType data returned from \"test_table\", Inner:\n\tRow could not be read, error: Scan failed.")))

})

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
