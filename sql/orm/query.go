package orm

import (
	"context"
	"database/sql"
	"strings"

	xsql "github.com/xefino/goutils/sql"
	xstr "github.com/xefino/goutils/strings"
	"github.com/xefino/goutils/utils"
)

// Constant Boolean connection operations
const (
	And = "AND"
	Or  = "OR"
)

// Query allows for the creation of execution of SQL queries in a programmatic manner
type Query[T any] struct {
	fields    []string
	table     string
	filter    *strings.Builder
	groupBy   string
	orderBy   string
	arguments []any
	logger    *utils.Logger
}

// NewQuery creates a new Query from a logger with default values
func NewQuery[T any](logger *utils.Logger) *Query[T] {
	return &Query[T]{
		fields:    make([]string, 0),
		filter:    new(strings.Builder),
		arguments: make([]any, 0),
		logger:    logger,
	}
}

// String converts a Query to its string equivalent
func (query *Query[T]) String() string {

	// First, if we have fields we want to select then connect them all with commas. Othewise, we'll
	// just assume we're querying all fields so use a star
	fields := "*"
	if len(query.fields) > 0 {
		fields = strings.Join(query.fields, ", ")
	}

	// Next, if we have any where clause then convert that to a string now
	var where string
	if query.filter.Len() > 0 {
		where = " WHERE " + query.filter.String()
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

	// Finally, add all the various query pieces together and return them
	return "SELECT " + fields + " FROM " + query.table + where + orderBy + groupBy
}

// Select determines which fields should be selected in the query, returning the modified query so that
// this function can be chained with others
func (query *Query[T]) Select(fields ...string) *Query[T] {
	query.fields = fields
	return query
}

// From sets the table that data should be pulled from, returning the modified query so that this function
// can be chained with others
func (query *Query[T]) From(table string) *Query[T] {
	query.table = table
	return query
}

// Where sets the filter clauses that should be used to determine which rows are returned from the query.
// The op variable should be either AND or OR and is used to chain the clauses together. This function
// returns the modified query so that it can be cahined with other functions.
func (query *Query[T]) Where(op string, clauses ...WhereClause[T]) *Query[T] {
	if len(clauses) == 0 {
		return query
	}

	clauses[0].ModifyQuery(query)
	if len(clauses) == 1 {
		return query
	}

	for _, clause := range clauses[1:] {
		query.filter.WriteByte(' ')
		query.filter.WriteString(op)
		query.filter.WriteByte(' ')
		clause.ModifyQuery(query)
	}

	return query
}

// GroupBy sets the fields that rows should be grouped by, returning the modified query so that this
// function can be chained with others
func (query *Query[T]) GroupBy(fields ...string) *Query[T] {
	query.groupBy = strings.Join(fields, ", ")
	return query
}

// OrderBy sets the fields that rows should be sorted by, returning the modified query so that this
// function can be chained with others
func (query *Query[T]) OrderBy(fields ...string) *Query[T] {
	query.orderBy = strings.Join(fields, ", ")
	return query
}

// Run runs the query against a provided database connection, returning the query results
func (query *Query[T]) Run(ctx context.Context, db *sql.DB) ([]*T, error) {

	// Attempt to run the query against a database connection; if this fails then log and return an error
	rows, err := db.QueryContext(ctx, query.String(), query.arguments...)
	if err != nil {
		typ := new(T)
		return nil, query.logger.Error(err, "Failed to query %T data from the %q table", *typ, query.table)
	}

	// Read the rows into a list of assets if we couldn't read the rows data then return an error
	data, err := xsql.ReadRows[T](rows)
	if err != nil {
		typ := new(T)
		return nil, query.logger.Error(err, "Failed to read %T data", *typ)
	}

	// Return the data we read
	return data, nil
}
