package dynamodb

import "time"

// IDynamoDBOption defines the functionality that will allow the behavior of a
// DatabaseConnection to be modified at construction
type IDynamoDBOption interface {
	Apply(*DatabaseConnection)
}

// WithBackoffStart allows the user to set the starting time to use when backing off from
// a DynamoDB error that should be retried
type WithBackoffStart time.Duration

// Apply modifies the DatabaseConnection so that it has the start interval defined by this object
func (w WithBackoffStart) Apply(conn *DatabaseConnection) {
	conn.startInterval = time.Duration(w)
}

// WithBackoffEnd allows the user to set the ending time to use when backing off from a
// DynamoDB error that should be retried
type WithBackoffEnd time.Duration

// Apply modifies the DatabaseConnection so that it has the end interval defined by this object
func (w WithBackoffEnd) Apply(conn *DatabaseConnection) {
	conn.endInterval = time.Duration(w)
}

// WithBackoffMaxElapsed allows the user to set the maximum time that should be allowed when
// DynamoDB returns an error that should be retried
type WithBackoffMaxElapsed time.Duration

// Apply modifies the DatabaseConnection so that it has the maximum interval defined by this object
func (w WithBackoffMaxElapsed) Apply(conn *DatabaseConnection) {
	conn.maxElapsed = time.Duration(w)
}
