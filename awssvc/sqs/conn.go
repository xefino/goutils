package sqs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/xefino/goutils/collections"
	"github.com/xefino/goutils/concurrency"
	"github.com/xefino/goutils/utils"
)

// SQSConnection contains functinoality allowing for systemical access to SQS
type SQSConnection struct {
	sqs           SQSAPI
	logger        *utils.Logger
	sendBatchSize int
}

// NewSQSConnection creates a new SQS connection from an AWS session and logger
func NewSQSConnection(cfg aws.Config, logger *utils.Logger, options ...SQSOption) *SQSConnection {
	return FromClient(sqs.NewFromConfig(cfg), logger, options...)
}

// FromClient creates a new SQS connection from an SQS client, a logger and options
func FromClient(inner SQSAPI, logger *utils.Logger, options ...SQSOption) *SQSConnection {

	// First, create our SQS connection from the config and logger with default values
	conn := SQSConnection{
		sqs:           inner,
		logger:        logger,
		sendBatchSize: 10,
	}

	// Next, iterate over all the options and apply each to the connection
	for _, option := range options {
		option.Apply(&conn)
	}

	// Finally, return a reference to the connection
	return &conn
}

// SendMessage attempts to convert the item to a message and send it to the SQS queue indicated by the
// URL. The options provided may be used to modify the request.
func (conn *SQSConnection) SendMessage(ctx context.Context, url string, item any,
	options ...SendMessageOption) (*sqs.SendMessageOutput, error) {

	// First, attempt to marshal the item to JSON; if this fails then return an error
	data, err := json.Marshal(item)
	if err != nil {
		return nil, conn.logger.Error(err, "Failed to convert payload to JSON")
	}

	// Next, encode the JSON data as a base-64 string and then embed it into a send-message input
	body := base64.StdEncoding.EncodeToString(data)
	input := sqs.SendMessageInput{
		MessageBody: aws.String(body),
		QueueUrl:    aws.String(url),
	}

	// Now, iterate over all the options and apply them to the input
	for _, option := range options {
		option.Apply(&input)
	}

	// Finally, attempt to send the message to SQS; if this fails then return an error
	output, err := conn.sqs.SendMessage(ctx, &input)
	if err != nil {
		return nil, conn.logger.Error(err, "Failed to send SQS message to %q", url)
	}

	return output, nil
}

// SendMessages attempts to convert the list of items to a batched message and send it to the SQS queue
// indicated by the URL. The options provided may be used to modify the request
func (conn *SQSConnection) SendMessages(ctx context.Context, url string, items []any,
	options ...SendMessageBatchOption) (*sqs.SendMessageBatchOutput, error) {

	// First, page the items into a number of slices
	pages := collections.Page(items, conn.sendBatchSize)

	// Next, create our combined output with failed and successful result entries
	output := sqs.SendMessageBatchOutput{
		Failed:     make([]types.BatchResultErrorEntry, 0),
		Successful: make([]types.SendMessageBatchResultEntry, 0),
	}

	// Now, iterate over all the pages and send each concurrently; this should fail if any page fails
	err := concurrency.ForAllAsync(ctx, len(pages), false,
		func(ctx context.Context, index int, cancel context.CancelFunc) error {

			// First, attempt to send the batched message; collect the output and error
			pageOut, err := conn.sendMessagesInner(ctx, url, pages[index], index*conn.sendBatchSize, options...)

			// Next, add the failed and successful results to the output if it exists
			if pageOut != nil {
				output.Failed = append(output.Failed, pageOut.Failed...)
				output.Successful = append(output.Successful, pageOut.Successful...)
			}

			// Finally, if we received an error then wrap, log and return it
			if err != nil {
				return conn.logger.Error(err, "Failed to send page %d of batched message to %q", index, url)
			}

			return nil
		})

	// Finally, return the combined output and error (if any)
	return &output, err
}

// GetURL retrieves the URL associated with the name of the SQS queue. The options provided may be used
// to modify the request.
func (conn *SQSConnection) GetURL(ctx context.Context, queueName string,
	options ...GetQueueUrlOption) (string, error) {

	// First, create our get-queue URL input from the queue name
	input := sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	}

	// Next, iterate over our options and apply each to the input
	for _, option := range options {
		option.Apply(&input)
	}

	// Now, attempt to get the URL associated with the input; if this fails then return an error
	output, err := conn.sqs.GetQueueUrl(ctx, &input)
	if err != nil {
		return "", conn.logger.Error(err, "Failed to retrieve SQS queue URL for queue %q", queueName)
	}

	// Finally, extract the URL from the output and return it
	return *output.QueueUrl, nil
}

// Helper function that sends a page of messages to the SQS queue
func (conn *SQSConnection) sendMessagesInner(ctx context.Context, url string, items []any, offset int,
	options ...SendMessageBatchOption) (*sqs.SendMessageBatchOutput, error) {

	// First, create the base send-message-batch input from the URL and a list of entries we'll fill in
	input := sqs.SendMessageBatchInput{
		QueueUrl: aws.String(url),
		Entries:  make([]types.SendMessageBatchRequestEntry, len(items)),
	}

	// Next, attempt to fill in the message batch entries with the items provided
	for i, item := range items {

		// First, attempt to marshal the item to JSON; if this fails then return an error
		data, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		// Next, encode the JSON data as a base-64 string and then embed it into a send-message input
		body := base64.StdEncoding.EncodeToString(data)
		input.Entries[i] = types.SendMessageBatchRequestEntry{
			Id:          aws.String(strconv.FormatInt(int64(i), 10)),
			MessageBody: aws.String(body),
		}

		// Finally, iterate over all the options and apply each to this entry
		for _, option := range options {
			option.ApplyBatch(i+offset, &input.Entries[i])
		}
	}

	// Finally, attempt to send the batched message to SQS; if this fails then return an error
	return conn.sqs.SendMessageBatch(ctx, &input)
}
