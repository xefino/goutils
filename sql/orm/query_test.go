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

var _ = Describe("Query Tests", func() {

	// Tests the conditions under which the Run function will return an error
	DescribeTable("Run - Failures",
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
			data, err := NewQuery[testType](logger).From("test_table").Run(context.Background(), db)

			// Finally, verify the error, that we received no data, and that our expectations were met
			verifier(err.(*utils.GError))
			gomega.Expect(data).Should(gomega.BeEmpty())
			gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
		},
		Entry("QueryContext fails - Error", true, false, testutils.ErrorVerifier("test", "orm",
			"/goutils/sql/orm/query.go", "Query", "Run", 130, testutils.InnerErrorVerifier("QueryContext failed"),
			"Failed to query orm.testType data from the \"test_table\" table", "[test] orm.Query.Run "+
				"(/goutils/sql/orm/query.go 130): Failed to query orm.testType data from the \"test_table\" "+
				"table, Inner: QueryContext failed.")),
		Entry("ReadRows fails - Error", false, true, testutils.ErrorVerifier("test", "orm",
			"/goutils/sql/orm/query.go", "Query", "Run", 137, testutils.InnerErrorVerifier("Row could not be read, error: Scan failed"),
			"Failed to read orm.testType data", "[test] orm.Query.Run (/goutils/sql/orm/query.go 137): "+
				"Failed to read orm.testType data, Inner: Row could not be read, error: Scan failed.")))

	// Tests that, if no error occurs, and only the table is specified, then all the data from that
	// table will be returned from the query
	It("Only FROM - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM test_table")).
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		data, err := NewQuery[testType](logger).From("test_table").Run(context.Background(), db)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value2")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})

	// Tests that, if no error occurs, and only a single SELECT field and the table is specified, then
	// all the data from that table will be returned from the query
	It("FROM, SELECT single field - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT TOP 1000 * FROM test_table")).
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		data, err := NewQuery[testType](logger).Select("TOP 1000 *").From("test_table").
			Run(context.Background(), db)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value2")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})

	// Tests that, if no error occurs, and multiple SELECT fields and the table is specified, then all
	// the data from that table will be returned from the query
	It("FROM, SELECT multiple fields - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table")).
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		data, err := NewQuery[testType](logger).Select("key", "value").From("test_table").Where(And).
			Run(context.Background(), db)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value2")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})

	// Tests that, if no error occurs, and multiple SELECT fields, the table and a single WHERE clause
	// are specified, then all the data from the table that conforms to the filter will be returned
	// from the query
	It("FROM, SELECT, WHERE constant comparison - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table WHERE value = 'value1'")).
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("key1", "value1").AddRow("key2", "value1"))

		// Now, attempt to create and run the query; this should not fail and should return data
		data, err := NewQuery[testType](logger).Select("key", "value").From("test_table").Where(And,
			NewConstantQueryTerm[testType]("value", Equals, "'value1'")).Run(context.Background(), db)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value1")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})

	// Tests that, if no error occurs, and multiple SELECT fields, the table and multiple WHERE clauses
	// are specified, then all the data from the table that conforms to the filter conditions will be
	// returned from the query
	It("FROM, SELECT, WHERE injected comparisons - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table WHERE value >= ? AND value < ?")).
			WithArgs("value1", "value2").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
			AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		data, err := NewQuery[testType](logger).Select("key", "value").From("test_table").Where(And,
			NewInjectedQueryTerm[testType]("value", GreaterThanEqualTo, "value1"),
			NewInjectedQueryTerm[testType]("value", LessThan, "value2"),
			NewMultiQueryTerm[testType](Or)).Run(context.Background(), db)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value2")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})

	// Tests that, if no error occurs, and multiple SELECT fields, the table and multiple WHERE clauses
	// are specified, then all the data from the table that conforms to the filter conditions will be
	// returned from the query
	It("FROM, SELECT, WHERE, multi-where clauses - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table WHERE (key LIKE key%) OR "+
			"(value >= ? AND value < ?)")).WithArgs("value1", "value2").
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		data, err := NewQuery[testType](logger).Select("key", "value").From("test_table").Where(Or,
			NewMultiQueryTerm[testType](And, NewConstantQueryTerm[testType]("key", Like, "key%")),
			NewMultiQueryTerm[testType](And,
				NewInjectedQueryTerm[testType]("value", GreaterThanEqualTo, "value1"),
				NewInjectedQueryTerm[testType]("value", LessThan, "value2"))).Run(context.Background(), db)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value2")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})

	// Tests that, if no error occurs, and multiple SELECT fields, the table, multiple WHERE clauses and
	// an ORDER BY condition are specified, then all the data from the table that conforms to the filter
	// conditions will be returned from the query
	It("SELECT, FROM, WHERE, ORDER BY - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table WHERE key LIKE key% OR "+
			"(value >= ? AND value < ?) ORDER BY key")).WithArgs("value1", "value2").
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		data, err := NewQuery[testType](logger).Select("key", "value").From("test_table").Where(Or,
			NewConstantQueryTerm[testType]("key", Like, "key%"), NewMultiQueryTerm[testType](And,
				NewInjectedQueryTerm[testType]("value", GreaterThanEqualTo, "value1"),
				NewInjectedQueryTerm[testType]("value", LessThan, "value2"))).OrderBy("key").
			Run(context.Background(), db)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value2")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})

	// Tests that, if no error occurs, and multiple SELECT fields, the table, multiple WHERE clauses, an
	// ORDER BY condition and a GROUP BY condition are specified, then all the data from the table that
	// conforms to the filter conditions will be returned from the query
	It("SELECT, FROM, WHERE, ORDER BY, GROUP BY - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table WHERE key LIKE key% OR "+
			"(value >= ? AND value < ?) ORDER BY key")).WithArgs("value1", "value2").
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		data, err := NewQuery[testType](logger).Select("key", "value").From("test_table").Where(Or,
			NewConstantQueryTerm[testType]("key", Like, "key%"), NewMultiQueryTerm[testType](And,
				NewInjectedQueryTerm[testType]("value", GreaterThanEqualTo, "value1"),
				NewInjectedQueryTerm[testType]("value", LessThan, "value2"))).OrderBy("key").GroupBy("key", "value").
			Run(context.Background(), db)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value2")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})
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
