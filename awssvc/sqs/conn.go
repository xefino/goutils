package sqs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/xefino/goutils/utils"
)

// SQSConnection contains functinoality allowing for systemical access to SQS
type SQSConnection struct {
	sqs    SQSAPI
	logger *utils.Logger
}

// NewSQSConnection creates a new SQS connection from an AWS session and logger
func NewSQSConnection(cfg aws.Config, logger *utils.Logger) *SQSConnection {
	return FromClient(sqs.NewFromConfig(cfg), logger)
}

// FromClient creates a new SQS connection from an SQS client, a logger and options
func FromClient(inner SQSAPI, logger *utils.Logger) *SQSConnection {

	// First, create our SQS connection from the config and logger with default values
	conn := SQSConnection{
		sqs:    inner,
		logger: logger,
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
			return nil, conn.logger.Error(err, "Failed to convert payload to JSON")
		}

		// Next, encode the JSON data as a base-64 string and then embed it into a send-message input
		body := base64.StdEncoding.EncodeToString(data)
		input.Entries[i] = types.SendMessageBatchRequestEntry{
			Id:          aws.String(strconv.FormatInt(int64(i), 10)),
			MessageBody: aws.String(body),
		}

		// Finally, iterate over all the options and apply each to this entry
		for _, option := range options {
			option.ApplyBatch(i, &input.Entries[i])
		}
	}

	// Finally, attempt to send the batched message to SQS; if this fails then return an error
	output, err := conn.sqs.SendMessageBatch(ctx, &input)
	if err != nil {
		return nil, conn.logger.Error(err, "Failed to send SQS batched message to %q", url)
	}

	return output, err
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
