package sql

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Create a new test runner we'll use to test all the
// modules in the reflection package
func TestReflection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reflection Suite")
}

var _ = Describe("ReadRows - Tests", func() {

	// Tests the conditions under which the readRows function will return an error
	DescribeTable("ReadRows - Failures",
		func(extraColumn bool, dataInvalid bool, rowsFailed bool, message string) {

			// First, create a mock SQL database connection; this should not fail
			db, mock, err := sqlmock.New()
			Expect(err).ShouldNot(HaveOccurred())

			// Next, generate our column names. If we want to test around a column
			// name-field name mismatch then add an extra value here
			columnNames := []string{"key1", "key2", "value1", "Value2"}
			if extraColumn {
				columnNames = append(columnNames, "extra")
			}

			// Generate the rows object that will be returned from the query
			// Set the data appropriate to the failure type we want to test
			results := mock.NewRows(columnNames)
			if dataInvalid {
				results.AddRow("forty-two", 42, "derp", "herp")
			} else if extraColumn {
				results.AddRow(420, 69, "derp", "herp", "sherbert")
			} else {
				results.AddRow(420, 69, "derp", "herp")
				if rowsFailed {
					results.RowError(0, fmt.Errorf("Rows failed to row"))
				}
			}

			// Setup a string to add to our expected query to test around missing columns
			var added string
			if extraColumn {
				added = ", extra"
			}

			// Setup our mock to expect the query value we're about to send
			mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf(
				"SELECT key1, key2, value1, Value2%s FROM testdb.test WHERE id = ?", added))).
				WithArgs(22).WillReturnRows(results)

			// Now, submit a query to our test database; we expect this query not to fail
			rows, err := db.QueryContext(context.Background(),
				fmt.Sprintf("SELECT key1, key2, value1, Value2%s FROM testdb.test WHERE id = ?", added), 22)
			Expect(err).ShouldNot(HaveOccurred())

			// Finally, attempt to read the data returned by the query; this should fail
			_, err = ReadRows[testValue](rows)

			// Verify the failure
			Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(message))
		},
		Entry("Returned column missing from object - Error", true, false, false,
			"Columns [extra] were not mapped to any field on sql.testValue"),
		Entry("Scan fails - Error", false, true, false,
			"Failed to read row into object of type sql.testValue, error: sql: Scan error on column index 0, name \"key1\": "+
				"converting driver.Value type string (\"forty-two\") to a int64: invalid syntax"),
		Entry("rows.Err returns Error - Error", false, false, true,
			"Row could not be read, error: Rows failed to row"))

	// Tests that, if the readRows function does not return an error, then all the rows
	// returned from the query will be read into a list of data objects and returned
	It("readRows - No failures - Data returned", func() {

		// First, create a mock SQL database connection; this should not fail
		db, mock, err := sqlmock.New()
		Expect(err).ShouldNot(HaveOccurred())

		// Next, generate the rows object that will be returned from the query
		results := mock.NewRows([]string{"VALUE_3", "key1", "key2", "value1", "Value2", "defined", "Value4"}).
			AddRow(11.9, 420, 69, "derp", "herp", "32", true)

		// Setup our mock to expect the query value we're about to send
		mock.ExpectQuery(regexp.QuoteMeta(
			"SELECT VALUE_3, key1, key2, value1, Value2, defined, Value4 FROM testdb.test WHERE id = ?")).
			WithArgs(22).WillReturnRows(results)

		// Now, submit a query to our test database; we expect this query not to fail
		rows, err := db.QueryContext(context.Background(),
			"SELECT VALUE_3, key1, key2, value1, Value2, defined, Value4 FROM testdb.test WHERE id = ?", 22)
		Expect(err).ShouldNot(HaveOccurred())

		// Finally, attempt to read the data returned by the query; this should not fail
		data, err := ReadRows[testValue2](rows)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())

		// Verify the data
		Expect(data).Should(HaveLen(1))
		Expect(data[0].private).Should(BeEmpty())
		Expect(data[0].Key1).Should(Equal(420))
		Expect(data[0].Key2).Should(Equal(69))
		Expect(data[0].Value1).Should(Equal("derp"))
		Expect(data[0].Value2).Should(Equal("herp"))
		Expect(data[0].Value3).Should(Equal(float32(11.899999618530273)))
		Expect(data[0].Value4).Should(BeTrue())
		Expect(data[0].Defined).Should(Equal(defined([]byte("32"))))
	})
})

// Helper type that we use for testing the ReadRows function
type testValue struct {
	private string
	Key1    int    `json:"key1"`
	Key2    int    `json:"Key2" sql:"key2"`
	Value1  string `json:"value1"`
	Value2  string
}

// Define type that we'll use to check that type definitions work
type defined []byte

// Define a scan function that we'll use to test around the ability of the algorithm to
// utilize the sql.Scanner interface when implemented on types
func (d *defined) Scan(raw interface{}) error {
	*d = defined([]byte(raw.(string)))
	return nil
}

// Helper type that we'll use for testing SQL reading
type testValue2 struct {
	private string
	Key1    int    `json:"key1"`
	Key2    int    `json:"Key2" sql:"key2"`
	Value1  string `json:"value1"`
	Value2  string
	Value3  float32 `sql:"value_3"`
	Value4  bool
	Defined defined
}
