package sqs

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

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
