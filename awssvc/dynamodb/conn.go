package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
	"github.com/cenkalti/backoff/v4"
	"github.com/xefino/goutils/utils"
)

// DatabaseConnection contains functinoality allowing for systemical access to DynamoDB
type DatabaseConnection struct {
	db            DynamoDBAPI
	startInterval time.Duration
	endInterval   time.Duration
	maxElapsed    time.Duration
	tagKey        string
	logger        *utils.Logger
}

// NewDatabaseConnection creates a new DynamoDB database connection from an AWS session and logger
func NewDatabaseConnection(cfg aws.Config, logger *utils.Logger, opts ...IDynamoDBOption) *DatabaseConnection {
	return FromClient(dynamodb.NewFromConfig(cfg), logger, opts...)
}

// FromClient creates a new DynamoDB database connection from a DynamoDB client, a logger and options
func FromClient(inner DynamoDBAPI, logger *utils.Logger, opts ...IDynamoDBOption) *DatabaseConnection {

	// First, create our database connection from the config and logger with default values
	conn := DatabaseConnection{
		db:            inner,
		startInterval: 500,
		endInterval:   60000,
		maxElapsed:    900000,
		tagKey:        "json",
		logger:        logger.ChangeFrame(4),
	}

	// Next, iterate over the options provided and update the associated values in the connection
	for _, opt := range opts {
		opt.Apply(&conn)
	}

	// Finally, return a reference to the connection
	return &conn
}

// PutItem writes an item to DynamoDB, overwriting an existing item if there is one
func (conn *DatabaseConnection) PutItem(ctx context.Context,
	input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {

	// Attempt to retry the operation to put the item in the table; if this
	// fails then we'll return the associated error. Otherwise, return the output
	var output *dynamodb.PutItemOutput
	err := conn.doRetry(ctx, *input.TableName, "PUT", func() error {
		var inner error
		output, inner = conn.db.PutItem(ctx, input)
		return inner
	})

	return output, err
}

// GetItem retrieves an item from DynamoDB
func (conn *DatabaseConnection) GetItem(ctx context.Context,
	input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {

	// Attempt to retry the operation to get the item from the table; if this
	// fails then we'll return the associated error. Otherwise, return the output
	var output *dynamodb.GetItemOutput
	err := conn.doRetry(ctx, *input.TableName, "GET", func() error {
		var inner error
		output, inner = conn.db.GetItem(ctx, input)
		return inner
	})

	return output, err
}

// UpdateItem updates desired fields on an item in DynamoDB
func (conn *DatabaseConnection) UpdateItem(ctx context.Context,
	input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {

	// Attempt to retry the operation to update the item in the table; if this
	// fails then we'll return the associated error. Otherwise, return the output
	var output *dynamodb.UpdateItemOutput
	err := conn.doRetry(ctx, *input.TableName, "UPDATE", func() error {
		var inner error
		output, inner = conn.db.UpdateItem(ctx, input)
		return inner
	})

	return output, err
}

// DeleteItem removes the item associated with the input from DynamoDB
func (conn *DatabaseConnection) DeleteItem(ctx context.Context,
	input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {

	// Attempt to retry the operation to delete the item from the table; if this
	// fails then we'll return the associated error. Otherwise, return the output
	var output *dynamodb.DeleteItemOutput
	err := conn.doRetry(ctx, *input.TableName, "DELETE", func() error {
		var inner error
		output, inner = conn.db.DeleteItem(ctx, input)
		return inner
	})

	return output, err
}

// BatchWrite makes a number of write requests against a table in DynamoDB. This function does not
// return collection or capacity statistics.
func (conn *DatabaseConnection) BatchWrite(ctx context.Context, tableName string,
	requests ...types.WriteRequest) error {

	// First, calculate the length from the number of requests. If there were no requests
	// then exit (this also serves as the base case for our recursion)
	length := len(requests)
	if length == 0 {
		return nil
	}

	conn.logger.Log("Attempting batch-write of %d entries to %s...", length, tableName)

	// Next, iterate over all the requests and chunk them so we don't have issues with the AWS
	// batch size and request limits; accumulate any unprocessed items
	retries := make([]types.WriteRequest, 0)
	for current := 0; current < length; current += 25 {
		next := current + 25

		// First, get the chunk of data we'll attempt to write. If we have 25 or more items left in
		// the chunk then we'll take the whole page; otherwise, we'll take the remainder of the list
		var chunk []types.WriteRequest
		if next > length {
			chunk = requests[current:]
		} else {
			chunk = requests[current:next]
		}

		// Next, attempt to write the items to DynamoDB; if this fails then return an error
		unprocessed, err := conn.batchWriteInner(ctx, tableName, chunk)
		if err != nil {
			return err
		}

		// Finally, save any unprocessed items so they can be written later
		retries = append(retries, unprocessed...)
	}

	// Finally, again attempt to do a batch write to the table with our unprocessed data
	conn.logger.Log("Batch-write to %s completed. Retries? %t", tableName, len(retries) == 0)
	return conn.BatchWrite(ctx, tableName, retries...)
}

// Query makes a search on a particular partition in a DynamoDB table and returns the results. This
// function does not return capacity statistics, just the queried results.
func (conn *DatabaseConnection) Query(ctx context.Context,
	input *dynamodb.QueryInput) ([]map[string]types.AttributeValue, error) {
	results := make([]map[string]types.AttributeValue, 0)

	// We'll start a loop that will query each page of results until all the pages have been retrieved
	for index := 0; ; index++ {

		// First, attempt the query with a backoff-retry loop
		var output *dynamodb.QueryOutput
		err := conn.doRetry(ctx, *input.TableName, fmt.Sprintf("QUERY(%d)", index), func() error {
			var inner error
			output, inner = conn.db.Query(ctx, input)
			return inner
		})

		// If the query failed then pass the error back up
		if err != nil {
			return nil, err
		}

		// Next, append the results from the query to our accumulated list of results
		results = append(results, output.Items...)

		// Finally, check if the lsat-evaluated key is nil. If it is then we've finished our query so
		// we can break out of the loop. Otherwise, we'll use it to set the exclusive start key on the
		// input so we can get the next page
		if output.LastEvaluatedKey != nil {
			input.ExclusiveStartKey = output.LastEvaluatedKey
		} else {
			break
		}
	}

	// Return the accumulated results
	return results, nil
}

// Scan makes a search on an entire DynamoDB table and returns the result. This function does not return
// capacity statistics, just the scanned results
func (conn *DatabaseConnection) Scan(ctx context.Context,
	input *dynamodb.ScanInput) ([]map[string]types.AttributeValue, error) {
	results := make([]map[string]types.AttributeValue, 0)

	// We'll start a loop that will scan each page of results until all the pages have been retrieved
	for index := 0; ; index++ {

		// First, attempt the scan with a backoff-retry loop
		var output *dynamodb.ScanOutput
		err := conn.doRetry(ctx, *input.TableName, fmt.Sprintf("SCAN(%d)", index), func() error {
			var inner error
			output, inner = conn.db.Scan(ctx, input)
			return inner
		})

		// If the scan failed then pass the error back up
		if err != nil {
			return nil, err
		}

		// Next, append the results from the query to our accumulated list of results
		results = append(results, output.Items...)

		// Finally, check if the lsat-evaluated key is nil. If it is then we've finished our scan so
		// we can break out of the loop. Otherwise, we'll use it to set the exclusive start key on the
		// input so we can get the next page
		if output.LastEvaluatedKey != nil {
			input.ExclusiveStartKey = output.LastEvaluatedKey
		} else {
			break
		}
	}

	// Return the accumulated results
	return results, nil
}

// Helper function that writes a single batch (no more than a single page) of write requests to
// a single table in DynamoDB
func (conn *DatabaseConnection) batchWriteInner(ctx context.Context, tableName string,
	inputs []types.WriteRequest) ([]types.WriteRequest, error) {

	// Create our batch write input from the inputs
	request := dynamodb.BatchWriteItemInput{
		ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
		ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone,
		RequestItems: map[string][]types.WriteRequest{
			tableName: inputs,
		},
	}

	// Attempt to retry the operation to batch-write the items to the table; if this
	// fails then we'll return the associated error. Otherwise, return the output
	var output *dynamodb.BatchWriteItemOutput
	err := conn.doRetry(ctx, tableName, "BATCH WRITE", func() error {
		var inner error
		output, inner = conn.db.BatchWriteItem(ctx, &request)
		return inner
	})

	// If the backoff returned an error then pass that error up
	if err != nil {
		return nil, err
	}

	// The backoff did not return an error so return any unprocessed items
	return output.UnprocessedItems[tableName], nil
}

// Helper function that does a retry operation to handle a number of common AWS DynamoDB retry cases
func (conn *DatabaseConnection) doRetry(ctx context.Context, tableName string, verb string,
	operation func() error) error {
	conn.logger.Log("Attempting %s operation to %s in DynamoDB...", verb, tableName)

	// Attempt the operation with a backoff in the case where an intermittent failure occurs
	err := backoff.Retry(func() error {
		if err := operation(); err != nil {
			var message string

			// Check that the error type is one that we'd want to retry on. For throughput or request
			// limit exceptions, waiting a bit may allow the request to succeed. For internal server
			// errors, since we're not sure of the cause, we'll wait to see if the service manages
			// to fix itself. Otherwise, we'll return the error wrapped in a permanent failure
			inner := err.(*smithy.OperationError)
			switch casted := inner.Err.(type) {
			case *types.ProvisionedThroughputExceededException:
				message = *casted.Message
			case *types.RequestLimitExceeded:
				message = *casted.Message
			case *types.InternalServerError:
				message = *casted.Message
			default:
				return backoff.Permanent(err)
			}

			// If we reached this point then we want to retry so log a message stating that there was
			// a failure and we're going to retry
			conn.logger.Log("DynamoDB request to %s failed: %s. Retrying...",
				tableName, message)
			return err
		}

		// Finally, since the operation did not fail we'll return nil to tell the backoff that
		// there's nothing else to do here
		conn.logger.Log("Completed %s operation to %s in DynamoDB", verb, tableName)
		return nil
	}, backoff.WithContext(conn.createExponentialBackoff(), ctx))

	// For whatever reason, the operation failed so create an error and return it
	if err != nil {
		return conn.NewError(err, tableName, "%s request to %s in DynamoDB failed", verb, tableName)
	}

	return nil
}

// Helper function that can be used to create an exponential backoff
// timer from values stored on the connection
func (conn *DatabaseConnection) createExponentialBackoff() *backoff.ExponentialBackOff {

	// Create the timer with values from the requester and some values that
	// are standard to all exponential backoff timers from the backoff library
	timer := backoff.NewExponentialBackOff()
	timer.InitialInterval = conn.startInterval * time.Millisecond
	timer.MaxInterval = conn.endInterval * time.Millisecond
	timer.MaxElapsedTime = conn.maxElapsed * time.Millisecond
	timer.RandomizationFactor = backoff.DefaultRandomizationFactor
	timer.Multiplier = backoff.DefaultMultiplier
	timer.Clock = backoff.SystemClock

	// Reset the timer and return it
	timer.Reset()
	return timer
}
