package orm

import (
	"strings"

	"github.com/xefino/goutils/collections"
	xstr "github.com/xefino/goutils/strings"
)

// SQL constants
const (
	All = "*"
)

// Constant Boolean connection operations
const (
	And = "AND"
	Or  = "OR"
)

// Query allows for the creation of execution of SQL queries in a programmatic manner
type Query struct {
	fields    []string
	table     string
	filter    string
	groupBy   string
	orderBy   string
	limit     string
	offset    string
	arguments []any
}

// NewQuery creates a new Query from a logger with default values
func NewQuery() *Query {
	return &Query{
		fields:    make([]string, 0),
		arguments: make([]any, 0),
	}
}

// Source returns the source of the Query, i.e. its table name
func (query *Query) Source() string {
	return query.table
}

// String converts a Query to its string equivalent
func (query *Query) String() string {

	// First, if we have fields we want to select then connect them all with commas. Othewise, we'll
	// just assume we're querying all fields so use a star
	fields := "*"
	if len(query.fields) > 0 {
		fields = strings.Join(query.fields, ", ")
	}

	// Next, if we have any where clause then convert that to a string now
	var where string
	if !xstr.IsEmpty(query.filter) {
		where = " WHERE " + query.filter
	}

	// Now, if we have any order-by fields then inject them into a clause
	var orderBy string
	if !xstr.IsEmpty(query.orderBy) {
		orderBy = " ORDER BY " + query.orderBy
	}

	// If we have any group-by fields then inject them into a clause as well
	var groupBy string
	if !xstr.IsEmpty(query.groupBy) {
		groupBy = " GROUP BY " + query.groupBy
	}

	// If we have a limit set then inject it into a clause as well
	var limit string
	if !xstr.IsEmpty(query.limit) {
		limit = " LIMIT " + query.limit
	}

	// If we have an offset then inject it into a clause
	var offset string
	if !xstr.IsEmpty(query.offset) {
		offset = " OFFSET " + query.offset
	}

	// Finally, add all the various query pieces together and return them
	return "SELECT " + fields + " FROM " + query.table + where + orderBy + groupBy + limit + offset
}

// Arguments returns the arguments that should be injected into the Query
func (query *Query) Arguments() []any {
	return query.arguments
}

// Select determines which fields should be selected in the query, returning the modified query so that
// this function can be chained with others
func (query *Query) Select(fields ...string) *Query {
	query.fields = fields
	return query
}

// From sets the table that data should be pulled from, returning the modified query so that this function
// can be chained with others
func (query *Query) From(table string) *Query {
	query.table = table
	return query
}

// FromQuery sets the table that the data should be pulled from so that it is the results of an inner query.
// This function returns the modified query so that it can be chained with others
func (query *Query) FromQuery(inner *Query) *Query {
	query.table = "(" + inner.String() + ")"
	query.arguments = append(query.arguments, inner.arguments...)
	return query
}

// Where sets the filter clauses that should be used to determine which rows are returned from the query.
// The op variable should be either AND or OR and is used to chain the clauses together. This function
// returns the modified query so that it can be cahined with other functions.
func (query *Query) Where(op string, clauses ...WhereClause) *Query {

	// If we have no clauses then return the query here
	if len(clauses) == 0 {
		return query
	}

	// Otherwise, we have at leat one clause so create our connector, and join all our clauses toegether with it,
	// writing the resulting string to the query filter
	connector := " " + op + " "
	query.filter = strings.Join(collections.Convert(
		func(clause WhereClause) string { return clause.ModifyQuery(query) }, clauses...), connector)
	return query
}

// GroupBy sets the fields that rows should be grouped by, returning the modified query so that this
// function can be chained with others
func (query *Query) GroupBy(fields ...string) *Query {
	query.groupBy = strings.Join(fields, ", ")
	return query
}

// OrderBy sets the fields that rows should be sorted by, returning the modified query so that this
// function can be chained with others
func (query *Query) OrderBy(fields ...string) *Query {
	query.orderBy = strings.Join(fields, ", ")
	return query
}

// Limit sets the limit on the number of rows that may be returned by the query
func (query *Query) Limit(limit any, constant bool) *Query {
	query.limit = param(limit, constant).ModifyQuery(query)
	return query
}

// Offset sets the offset on the number of rows that should be skipped before being returned by the query
func (query *Query) Offset(offset any, constant bool) *Query {
	query.offset = param(offset, constant).ModifyQuery(query)
	return query
}
