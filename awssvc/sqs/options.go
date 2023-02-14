package sqs

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// SQSOption contains the functionality necessary to modify the SQS connection
type SQSOption interface {
	Apply(*SQSConnection)
}

// WithSendMessagesBatchSize allows the user to modify the expected batch size to use when sending data
// to the SQS.SendMessageBatch function
type WithSendMessagesBatchSize int

// Apply sets the send-message batch size associated with the SQS connection
func (w WithSendMessagesBatchSize) Apply(sqs *SQSConnection) {
	sqs.sendBatchSize = int(w)
}

// SendMessageOption describes the functionality necessary to configure an SQS.SendMessageInput beyond
// the base fields required
type SendMessageOption interface {
	Apply(*sqs.SendMessageInput)
}

// WithDeduplicationID allows the user to set the deduplication ID on the SQS.SendMessageInput
type WithDeduplicationID string

// Apply sets the deduplication ID on the SQS.SendMessageInput
func (w WithDeduplicationID) Apply(input *sqs.SendMessageInput) {
	input.MessageDeduplicationId = aws.String(string(w))
}

// WithMessageGroupID allows the user to set the message group ID on the SQS.SendMessageInput
type WithMessageGroupID string

// Apply sets the message group ID on the SQS.SendMessageInput
func (w WithMessageGroupID) Apply(input *sqs.SendMessageInput) {
	input.MessageGroupId = aws.String(string(w))
}

// Helper type that allows the user to add a message attribute to the SQS.SendMessageInput
type withMessageAttribute struct {
	key  string
	attr *types.MessageAttributeValue
}

// WithMessageAttribute allows the user to add a new message attribute or orverwrite an existing
// message attribute on the SQS.SendMessageInput
func WithMessageAttribute(key string, attr *types.MessageAttributeValue) withMessageAttribute {
	return withMessageAttribute{
		key:  key,
		attr: attr,
	}
}

// Apply sets the attribute value associated with the key on the message attribute values collection
// associated with the SQS.SendMessageInput provided
func (w withMessageAttribute) Apply(input *sqs.SendMessageInput) {
	if len(input.MessageAttributes) == 0 {
		input.MessageAttributes = make(map[string]types.MessageAttributeValue)
	}

	input.MessageAttributes[w.key] = *w.attr
}

// SendMessageBatchOption describes the functionality necessary to configure an SQS.SendMessageBatchInput
// beyond the base fields required
type SendMessageBatchOption interface {
	ApplyBatch(int, *types.SendMessageBatchRequestEntry)
}

// Helper type that allows the user to set the deduplication ID on the SQS.SendMessageBatchInput
type withBatchDeduplicationID struct {
	applier func(int) string
}

// WithBatchDeduplicationID creates a new object from an applier function that will set the deduplication
// ID on each request entry in the SQS.SendMessageBatchInput
func WithBatchDeduplicationID(applier func(int) string) withBatchDeduplicationID {
	return withBatchDeduplicationID{applier: applier}
}

// ApplyBatch sets the deduplication ID on each request entry in the SQS.SendMessageBatchInput
func (w withBatchDeduplicationID) ApplyBatch(index int, entry *types.SendMessageBatchRequestEntry) {
	entry.MessageDeduplicationId = aws.String(w.applier(index))
}

// Helper type that allows the user to set the message group ID on the SQS.SendMessageBatchInput
type withBatchMessageGroupID struct {
	applier func(int) string
}

// WithBatchMessageGroupID creates a new object from an applier function that will set the message group
// ID on each request entry in the SQS.SendMessageBatchInput
func WithBatchMessageGroupID(applier func(int) string) withBatchMessageGroupID {
	return withBatchMessageGroupID{applier: applier}
}

// ApplyBatch sets the message group ID on each request entry in the SQS.SendMessageBatchInput
func (w withBatchMessageGroupID) ApplyBatch(index int, entry *types.SendMessageBatchRequestEntry) {
	entry.MessageGroupId = aws.String(w.applier(index))
}

// Helper type that allows the user to add a new message attribute value or overwrite an existing one
// associated with the key provided on each entry in the SQS.SendMessageBatchInput
type withBatchMessageAttribute struct {
	key     string
	applier func(int) *types.MessageAttributeValue
}

// WithBatchMessageAttribute creates a new object from an attribute key and an applier function that
// will add a new message attribute or orverwrite an existing message attribute on each request entry
// in the SQS.SendMessageBatchInput
func WithBatchMessageAttribute(key string, applier func(int) *types.MessageAttributeValue) withBatchMessageAttribute {
	return withBatchMessageAttribute{
		key:     key,
		applier: applier,
	}
}

// ApplyBatch sets the attribute value associated with the key on the message attribute values collection
// associated with each request entry in the SQS.SendMessageBatchInput
func (w withBatchMessageAttribute) ApplyBatch(index int, entry *types.SendMessageBatchRequestEntry) {
	if len(entry.MessageAttributes) == 0 {
		entry.MessageAttributes = make(map[string]types.MessageAttributeValue)
	}

	entry.MessageAttributes[w.key] = *w.applier(index)
}

// GetQueueUrlOption describes the functionality necessary to configure an SQS.GetQueueUrlInput beyond
// the base fields required
type GetQueueUrlOption interface {
	Apply(*sqs.GetQueueUrlInput)
}

// WithAWSAccountID allows the user to set the queue owner AWS account ID on the SQS.GetQueueUrlInput
type WithAWSAccountID string

// Apply sets the queue owner AWS account ID on the SQS.GetQueueUrlInput
func (w WithAWSAccountID) Apply(input *sqs.GetQueueUrlInput) {
	input.QueueOwnerAWSAccountId = aws.String(string(w))
}
