package orm

import (
	"fmt"
	"strings"

	"github.com/xefino/goutils/collections"
	xstr "github.com/xefino/goutils/strings"
)

// WhereClause describes the functionality that should exist in any where clause terms
type WhereClause interface {
	ModifyQuery(*Query) string
}

// QueryTerm is the base where clause that allows a single field-parameter comparison
type QueryTerm struct {
	Name     string
	Operator string
	Value    Parameter
}

// NewQueryTerm creates a new QueryTerm from a name, operation and parameter
func NewQueryTerm(name string, op string, param Parameter) *QueryTerm {
	return &QueryTerm{
		Name:     name,
		Operator: op,
		Value:    param,
	}
}

// ModifyQuery modifies the query to include this query term
func (term *QueryTerm) ModifyQuery(query *Query) string {
	return fmt.Sprintf("%s %s %s", term.Name, term.Operator, term.Value.ModifyQuery(query))
}

// UnaryQueryTerm creates a new query term that allows a single query term to be combined with a unary operator
type UnaryQueryTerm struct {
	Operator string
	Clause   WhereClause
}

// NewUnaryQueryTerm creates a new unary query term from an operator and an inner where clause
func NewUnaryQueryTerm(op string, clause WhereClause) *UnaryQueryTerm {
	return &UnaryQueryTerm{
		Operator: op,
		Clause:   clause,
	}
}

// ModifyQuery modifies the query to include this query term
func (term *UnaryQueryTerm) ModifyQuery(query *Query) string {

	// Get the result of the clause; if this is empty then we have no work to do here
	result := term.Clause.ModifyQuery(query)
	if xstr.IsEmpty(result) {
		return ""
	}

	// Return the result enclosed in parentheses, preceeded by the NOT keyword
	return "NOT (" + result + ")"
}

// MultiQueryTerm creates a new query term that allows multiple query terms to be joined together inside
// a set of parentheses. This is intended to allow for alternating sets of AND/OR logic (i.e. A AND (B OR C))
type MultiQueryTerm struct {
	Operator string
	Inner    []WhereClause
}

// NewMultiQueryTerm creates a new multi-query term from a connecting operator and a list of inner terms
func NewMultiQueryTerm(op string, terms ...WhereClause) *MultiQueryTerm {
	return &MultiQueryTerm{
		Operator: op,
		Inner:    terms,
	}
}

// ModifyQuery modifies the query to include this query term
func (term *MultiQueryTerm) ModifyQuery(query *Query) string {

	// If we have no inner terms then we have no work to do so return here
	if len(term.Inner) == 0 {
		return ""
	}

	// Iterate over all the clauses, modify the query with each and then combine them into a comma-delimited
	// list, and then return the result
	connector := " " + term.Operator + " "
	return fmt.Sprintf("(%s)", strings.Join(collections.Convert(
		func(clause WhereClause) string { return clause.ModifyQuery(query) }, term.Inner...), connector))
}

// FunctionCallQueryTerm creates a new function call query term that allows the user to inject an SQL function
// call into the WHERE clause of an SQL query
type FunctionCallQueryTerm struct {
	Call      string
	Arguments []any
}

// NewFunctionCallQueryTerm creates a new function call query term from a function name and arguments
func NewFunctionCallQueryTerm(call string, args ...any) *FunctionCallQueryTerm {
	return &FunctionCallQueryTerm{
		Call:      call,
		Arguments: args,
	}
}

// ModifyQuery modifies the query to include this query term
func (term *FunctionCallQueryTerm) ModifyQuery(query *Query) string {
	query.arguments = append(query.arguments, term.Arguments...)
	return term.Call
}

// Not creates a new where clause negating the clause sent to it as a parameter
func Not(clause WhereClause) WhereClause {
	return NewUnaryQueryTerm("NOT", clause)
}

// Equals creates a new where clause stating that a field is equal to a parameter
func Equals(field string, value any, constant bool) WhereClause {
	return NewQueryTerm(field, "=", param(value, constant))
}

// NotEquals creates a new where clause stating that a field is not equal to a parameter
func NotEquals(field string, value any, constant bool) WhereClause {
	return NewQueryTerm(field, "<>", param(value, constant))
}

// GreaterThan creates a new where clause stating that a field is greater than a parameter
func GreaterThan(field string, value any, constant bool) WhereClause {
	return NewQueryTerm(field, ">", param(value, constant))
}

// GreaterThanOrEqualTo creates a new where clause stating that a field is greater than or equal to a parameter
func GreaterThanOrEqualTo(field string, value any, constant bool) WhereClause {
	return NewQueryTerm(field, ">=", param(value, constant))
}

// LessThan creates a new where clause stating that a field is less than a parameter
func LessThan(field string, value any, constant bool) WhereClause {
	return NewQueryTerm(field, "<", param(value, constant))
}

// LessThanOrEqualTo creates a new where clause stating that a field is less than or equal to a parameter
func LessThanOrEqualTo(field string, value any, constant bool) WhereClause {
	return NewQueryTerm(field, "<=", param(value, constant))
}

// Like creates a new where clause stating that a field should be compared to a value using the LIKE keyword
func Like(field string, value any, constant bool) WhereClause {
	return NewQueryTerm(field, "LIKE", param(value, constant))
}

// IsNull creates a new where clause asserting that a field is null
func IsNull(field string) WhereClause {
	return NewQueryTerm(field, "IS", NewKeyword(Null))
}

// NotNull creates a new where clause asserting that a field is not null
func NotNull(field string) WhereClause {
	return NewQueryTerm(field, "IS NOT", NewKeyword(Null))
}

// Between creates a new where clause stating that a field is between two values
func Between[T any](field string, lower T, lowerConst bool, upper T, upperConst bool) WhereClause {
	return NewMultiQueryTerm(And, NewQueryTerm(field, ">=", param(lower, lowerConst)),
		NewQueryTerm(field, "<", param(upper, upperConst)))
}
