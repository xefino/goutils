package dynamodb

import (
	"github.com/xefino/goutils/utils"
)

// Error describes an error returned by the DynamoDB database connection
type Error struct {
	*utils.GError
	TableName string
}

// NewError creates a new Error from an inner error, table name, message and arguments
func (conn *DatabaseConnection) NewError(inner error, tableName string,
	message string, args ...interface{}) *Error {
	return &Error{
		GError:    conn.logger.Error(inner, message, args...),
		TableName: tableName,
	}
}
