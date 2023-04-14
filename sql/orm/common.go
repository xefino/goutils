package orm

import (
	"context"
	"database/sql"

	xsql "github.com/xefino/goutils/sql"
	"github.com/xefino/goutils/utils"
)

// Statement describes the functionality that should be associated with an SQL-executable statement
type Statement interface {

	// Source describes what resource the statement is being executed against
	Source() string

	// String returns the string representation of the statement
	String() string

	// Arguments returns the list of arguments that should be injected with the statement
	Arguments() []any
}

// Run runs the query against a provided database connection, returning the query results
func RunQuery[TReturn any, TStatement Statement](ctx context.Context, query TStatement, db *sql.DB,
	logger *utils.Logger) ([]*TReturn, error) {

	// Attempt to run the query against a database connection; if this fails then log and return an error
	rows, err := db.QueryContext(ctx, query.String(), query.Arguments()...)
	if err != nil {
		typ := new(TReturn)
		return nil, logger.Error(err, "Failed to query %T data from %q", *typ, query.Source())
	}

	// Read the rows into a list of assets if we couldn't read the rows data then return an error
	data, err := xsql.ReadRows[TReturn](rows)
	if err != nil {
		typ := new(TReturn)
		return nil, logger.Error(err, "Failed to read %T data returned from %q", *typ, query.Source())
	}

	// Return the data we read
	return data, nil
}
