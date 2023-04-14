package orm

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("Where Tests", func() {

	// Test that the ModifyQuery function works as expected for the QueryTerm
	It("QueryTerm - ModifyQuery - Works", func() {

		// First, create our query term
		term := NewQueryTerm("value", "=", NewArgument("derp"))

		// Next, modify a new query with the query term
		query := NewQuery()
		result := term.ModifyQuery(query)

		// Finally, verify the result of the modification
		gomega.Expect(result).Should(gomega.Equal("value = ?"))
		gomega.Expect(query.arguments).Should(gomega.HaveLen(1))
		gomega.Expect(query.arguments[0]).Should(gomega.Equal("derp"))
	})

	// Test that the ModifyQuery function returns no data if the inner query returns no result
	It("UnaryQueryTerm - ModifyQuery - No inner result - Works", func() {

		// First, create our unary query term with no inner query terms
		term := NewUnaryQueryTerm("NOT", NewMultiQueryTerm(And))

		// Next, attempt to modify the query with our unary query term
		result := term.ModifyQuery(NewQuery())

		// Finally, verify that there was no resulting query clause
		gomega.Expect(result).Should(gomega.BeEmpty())
	})

	// Test that the ModifyQuery function works as expected for the UnaryQueryTerm
	It("UnaryQueryTerm - ModifyQuery - Works", func() {

		// First, create our unary query term with an inner query term
		term := NewUnaryQueryTerm("NOT", NewQueryTerm("value", "=", NewConstant("derp")))

		// Next, attempt to modify the query with our unary query term
		result := term.ModifyQuery(NewQuery())

		// Finally, verify the resulting query clause
		gomega.Expect(result).Should(gomega.Equal("NOT (value = 'derp')"))
	})

	// Test that the ModifyQuery function will produce no output if there are no inner terms provided
	// to the MultiQueryTerm when it was created
	It("MultiQueryTerm - ModifyQuery - No inner terms - Works", func() {

		// First, create our multi-query term with no inner query terms
		term := NewMultiQueryTerm(And)

		// Next, attempt to modify the query with our multi-query term
		result := term.ModifyQuery(NewQuery())

		// Finally, verify that no modification took place
		gomega.Expect(result).Should(gomega.BeEmpty())
	})

	// Test that the ModifyQuery function works as expected for the MultiQueryTerm when at least one
	// inner query term is present
	It("MultiQueryTerm - ModifyQuery - At least one inner term - Works", func() {

		// First, create our multi-query term with no inner query terms
		term := NewMultiQueryTerm(And, NewQueryTerm("value", ">=", NewArgument("derp")),
			NewQueryTerm("value", "<", NewConstant("herp")))

		// Next, attempt to modify the query with our multi-query term
		query := NewQuery()
		result := term.ModifyQuery(query)

		// Finally, verify that the resulting query clause and arguments
		gomega.Expect(result).Should(gomega.Equal("(value >= ? AND value < 'herp')"))
		gomega.Expect(query.arguments).Should(gomega.HaveLen(1))
		gomega.Expect(query.arguments[0]).Should(gomega.Equal("derp"))
	})

	// Test that the ModifyQuery function works as expected for the FunctionCallQueryTerm
	It("FunctionCallQueryTerm - ModifyQuery - Works", func() {

		// First, create our multi-query term with no inner query terms
		term := NewFunctionCallQueryTerm("RLIKE(value, '.*' + ? + '.*', 'i')", "derp")

		// Next, attempt to modify the query with our function call query term
		query := NewQuery()
		result := term.ModifyQuery(query)

		// Finally, verify that the resulting query clause and arguments
		gomega.Expect(result).Should(gomega.Equal("RLIKE(value, '.*' + ? + '.*', 'i')"))
		gomega.Expect(query.arguments).Should(gomega.HaveLen(1))
		gomega.Expect(query.arguments[0]).Should(gomega.Equal("derp"))
	})

	// Test that the Not function works as expected
	It("Not - Works", func() {

		// First, create our term using the Not function
		term := Not(NewQueryTerm("value", "=", NewConstant("derp")))

		// Next, attempt to modify the query with our where clause
		result := term.ModifyQuery(NewQuery())

		// Finally, verify the resulting query clause
		gomega.Expect(result).Should(gomega.Equal("NOT (value = 'derp')"))
	})

	// Test that the various binary operator functions work as expected
	DescribeTable("Binary Operator Function Tests",
		func(op func(string, any, bool) WhereClause, isConstant bool, resultStr string) {

			// First, create our binary term with a constant field name, value and a flag indicating
			// whether the value is constant or a variable
			term := op("value", "derp", isConstant)

			// Next, attempt to modify the query with our where clause
			query := NewQuery()
			result := term.ModifyQuery(query)

			// Finally, verify that the resulting query clause and arguments
			gomega.Expect(result).Should(gomega.Equal(resultStr))
			if isConstant {
				gomega.Expect(query.arguments).Should(gomega.BeEmpty())
			} else {
				gomega.Expect(query.arguments).Should(gomega.HaveLen(1))
				gomega.Expect(query.arguments[0]).Should(gomega.Equal("derp"))
			}
		},
		Entry("Equals - Argument - Works", Equals, false, "value = ?"),
		Entry("Equals - Constant - Works", Equals, true, "value = 'derp'"),
		Entry("NotEquals - Argument - Works", NotEquals, false, "value <> ?"),
		Entry("NotEquals - Constant - Works", NotEquals, true, "value <> 'derp'"),
		Entry("GreaterThan - Argument - Works", GreaterThan, false, "value > ?"),
		Entry("GreaterThan - Constant - Works", GreaterThan, true, "value > 'derp'"),
		Entry("GreaterThanOrEqualTo - Argument - Works", GreaterThanOrEqualTo, false, "value >= ?"),
		Entry("GreaterThanOrEqualTo - Constant - Works", GreaterThanOrEqualTo, true, "value >= 'derp'"),
		Entry("LessThan - Argument - Works", LessThan, false, "value < ?"),
		Entry("LessThan - Constant - Works", LessThan, true, "value < 'derp'"),
		Entry("LessThanOrEqualTo - Argument - Works", LessThanOrEqualTo, false, "value <= ?"),
		Entry("LessThanOrEqualTo - Constant - Works", LessThanOrEqualTo, true, "value <= 'derp'"),
		Entry("Like - Argument - Works", Like, false, "value LIKE ?"),
		Entry("Like - Constant - Works", Like, true, "value LIKE 'derp'"))

	// Test that the Between function works for all possible data conditions
	DescribeTable("Between - Conditions",
		func(lowerConstant bool, upperConstant bool, resultStr string) {

			// First, create our between term with a constant field name and field values, and two flags
			// indicating whether the bounds are constant or variable
			term := Between("value", 25, lowerConstant, 50, upperConstant)

			// Next, attempt to modify the query with our where clause
			query := NewQuery()
			result := term.ModifyQuery(query)

			// Finally, verify that the resulting query clause and arguments
			gomega.Expect(result).Should(gomega.Equal(resultStr))
			if lowerConstant && upperConstant {
				gomega.Expect(query.arguments).Should(gomega.BeEmpty())
			} else if lowerConstant {
				gomega.Expect(query.arguments).Should(gomega.HaveLen(1))
				gomega.Expect(query.arguments[0]).Should(gomega.Equal(50))
			} else if upperConstant {
				gomega.Expect(query.arguments).Should(gomega.HaveLen(1))
				gomega.Expect(query.arguments[0]).Should(gomega.Equal(25))
			} else {
				gomega.Expect(query.arguments).Should(gomega.HaveLen(2))
				gomega.Expect(query.arguments[0]).Should(gomega.Equal(25))
				gomega.Expect(query.arguments[1]).Should(gomega.Equal(50))
			}
		},
		Entry("Lower is constant, Upper is constant - Works", true, true, "(value >= 25 AND value < 50)"),
		Entry("Lower is constant, Upper is argument - Works", true, false, "(value >= 25 AND value < ?)"),
		Entry("Lower is argument, Upper is constant - Works", false, true, "(value >= ? AND value < 50)"),
		Entry("Lower is argument, Upper is argument - Works", false, false, "(value >= ? AND value < ?)"))
})
