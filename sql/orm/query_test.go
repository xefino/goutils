package orm

import (
	"context"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/xefino/goutils/utils"
)

var _ = Describe("Query Tests", func() {

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
		data, err := RunQuery[testType](context.Background(), NewQuery().From("test_table"), db, logger)
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
		query := NewQuery().Select("TOP 1000 *").From("test_table")
		data, err := RunQuery[testType](context.Background(), query, db, logger)
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
		query := NewQuery().Select("key", "value").From("test_table").Where(And)
		data, err := RunQuery[testType](context.Background(), query, db, logger)
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
		query := NewQuery().Select("key", "value").From("test_table").Where(And, Equals("value", "value1", true))
		data, err := RunQuery[testType](context.Background(), query, db, logger)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value1")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})

	// Tests that, if an no error occurs, and multiple SELECT fields, a FROM clause extracted from an
	// inner query, and a single WHERE clause are specified, then all the data resulting from the inner
	// query that conforms to the filter will be returned from the query.
	It("FROM is inner query - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM (SELECT *, ROW_NUMBER() OVER (" +
			"PARTITION BY key ORDER BY value DESC) AS rn FROM test_table) WHERE rn = 1 AND value = 'value1'")).
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("key1", "value1").AddRow("key2", "value1"))

		// Now, attempt to create and run the query; this should not fail and should return data
		query := NewQuery().Select("key", "value").FromQuery(
			NewQuery().Select(All, "ROW_NUMBER() OVER (PARTITION BY key ORDER BY value DESC) AS rn").
				From("test_table")).Where(And, Equals("rn", 1, true), Equals("value", "value1", true))
		data, err := RunQuery[testType](context.Background(), query, db, logger)
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
		query := NewQuery().Select("key", "value").From("test_table").Where(And,
			GreaterThanOrEqualTo("value", "value1", false), LessThan("value", "value2", false), NewMultiQueryTerm(Or))
		data, err := RunQuery[testType](context.Background(), query, db, logger)
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
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table WHERE (key LIKE 'key_') OR "+
			"(value >= ? AND value < ?)")).WithArgs("value1", "value2").
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		query := NewQuery().Select("key", "value").From("test_table").Where(Or,
			NewMultiQueryTerm(And, Like("key", "key_", true)), Between("value", "value1", false, "value2", false))
		data, err := RunQuery[testType](context.Background(), query, db, logger)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value2")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})

	// Tests that, if no error occurs, and multiple SELECT fields, the table and a function-call WHERE
	// clause are specified, then all the data from the table that conforms to the filter conditions
	// will be returned from the query
	It("FROM, SELECT, WHERE, function call clause - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table WHERE RLIKE(key, '.*key.*', 'i') OR "+
			"(value >= ? AND value < ?)")).WithArgs("value1", "value2").
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		query := NewQuery().Select("key", "value").From("test_table").Where(Or,
			NewFunctionCallQueryTerm("RLIKE(key, '.*key.*', 'i')"),
			Between("value", "value1", false, "value2", false))
		data, err := RunQuery[testType](context.Background(), query, db, logger)
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
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table WHERE key LIKE 'key_' OR "+
			"(value >= ? AND value < ?) ORDER BY key")).WithArgs("value1", "value2").
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		query := NewQuery().Select("key", "value").From("test_table").Where(Or, Like("key", "key_", true),
			Between("value", "value1", false, "value2", false)).OrderBy("key")
		data, err := RunQuery[testType](context.Background(), query, db, logger)
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
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table WHERE key LIKE 'key_' OR "+
			"(value >= ? AND value < ?) GROUP BY key, value ORDER BY key")).WithArgs("value1", "value2").
			WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
				AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		query := NewQuery().Select("key", "value").From("test_table").Where(Or, Like("key", "key_", true),
			NewMultiQueryTerm(And, GreaterThanOrEqualTo("value", "value1", false),
				LessThan("value", "value2", false))).OrderBy("key").GroupBy("key", "value")
		data, err := RunQuery[testType](context.Background(), query, db, logger)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value2")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})

	// Tests that, if no error occurs, and multiple SELECT fields, the table, multiple WHERE clauses,
	// an ORDER BY condition, a GROUP BY condition and a LIMIT are specified, then all the data from
	// the table that conforms to the filter conditions will be returned from the query
	It("SELECT, FROM, WHERE, ORDER BY, GROUP BY, LIMIT - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table WHERE key LIKE 'key_' OR "+
			"(value >= ? AND value < ?) GROUP BY key, value ORDER BY key LIMIT 1000")).
			WithArgs("value1", "value2").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
			AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		query := NewQuery().Select("key", "value").From("test_table").Where(Or, Like("key", "key_", true),
			Between("value", "value1", false, "value2", false)).OrderBy("key").GroupBy("key", "value").Limit(1000, true)
		data, err := RunQuery[testType](context.Background(), query, db, logger)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value2")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})

	// Tests that, if no error occurs, and multiple SELECT fields, the table, multiple WHERE clauses,
	// an ORDER BY condition, a GROUP BY condition, a LIMIT and an OFFSET are specified, then all the
	// data from the table that conforms to the filter conditions will be returned from the query
	It("SELECT, FROM, WHERE, ORDER BY, GROUP BY, LIMIT, OFFSET - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table WHERE key LIKE 'key_' OR "+
			"(value >= ? AND value < ?) GROUP BY key, value ORDER BY key LIMIT 1000 OFFSET 0")).
			WithArgs("value1", "value2").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
			AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		query := NewQuery().Select("key", "value").From("test_table").Where(Or, Like("key", "key_", true),
			Between("value", "value1", false, "value2", false)).OrderBy("key").GroupBy("key", "value").
			Having(And).Limit(1000, true).Offset(0, true)
		data, err := RunQuery[testType](context.Background(), query, db, logger)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value2")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})

	// Tests that, if no error occurs, and multiple SELECT fields, the table, multiple WHERE clauses,
	// an ORDER BY condition, a GROUP BY condition, a HAVING condition, a LIMIT and an OFFSET are specified,
	// then all the data from the table that conforms to the filter conditions will be returned from the query
	It("SELECT, FROM, WHERE, ORDER BY, GROUP BY, HAVING, LIMIT, OFFSET - Works", func() {

		// First, create our logger and discard its output
		logger := utils.NewLogger("query", "test")
		logger.Discard()

		// Next, create our mock database connection; this should not fail
		db, mock, err := sqlmock.New()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Inform the mock of the queries we expect to be made and what should be returned
		mock.ExpectQuery(regexp.QuoteMeta("SELECT key, value FROM test_table WHERE key LIKE 'key_' OR "+
			"(value >= ? AND value < ?) GROUP BY key, value HAVING COUNT(*) > 1 ORDER BY key LIMIT 1000 OFFSET 0")).
			WithArgs("value1", "value2").WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).
			AddRow("key1", "value1").AddRow("key2", "value2"))

		// Now, attempt to create and run the query; this should not fail and should return data
		query := NewQuery().Select("key", "value").From("test_table").Where(Or, Like("key", "key_", true),
			Between("value", "value1", false, "value2", false)).OrderBy("key").GroupBy("key", "value").
			Having(And, GreaterThan("COUNT(*)", 1, true)).Limit(1000, true).Offset(0, true)
		data, err := RunQuery[testType](context.Background(), query, db, logger)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Finally, verify the data and the mock expectations
		gomega.Expect(data).Should(gomega.HaveLen(2))
		verifyTestType(data[0], "key1", "value1")
		verifyTestType(data[1], "key2", "value2")
		gomega.Expect(mock.ExpectationsWereMet()).ShouldNot(gomega.HaveOccurred())
	})
})
