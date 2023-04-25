package sqs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/xefino/goutils/awssvc/policy"
	"github.com/xefino/goutils/awssvc/testing"
	"github.com/xefino/goutils/testutils"
	"github.com/xefino/goutils/utils"
)

var _ = Describe("SQS Connection Tests", Ordered, func() {

	// Ensure that the AWS config is created before each test; this could be set as a global variable
	var cfg aws.Config
	BeforeAll(func() {
		cfg = testing.TestAWSConfig(context.Background(), "us-east-1", 4100)
	})

	// Ensure that our test queue exists before the start of each test
	var queueUrl string
	BeforeEach(func() {

		// First, create a new client from our test config
		client := sqs.NewFromConfig(cfg)

		// Next, create our SQS policy
		policy := policy.Policy{
			Version: "2008-10-17",
			ID:      "__default_policy_ID",
			Statements: []*policy.Statement{
				{
					ID:            "__owner_statement",
					Effect:        policy.Allow,
					PrincipalArns: []string{"arn:aws:kms:eu-west-2:111122223333:root"},
					ActionArns:    policy.Actions{policy.SqsAll},
					ResourceArns:  []string{"*"},
				},
			},
		}

		// Serialize our test SQS policy to JSON; this should not fail
		policyStr, err := json.Marshal(policy)
		if err != nil {
			panic(err)
		}

		// Now, create a new create-queue input with our test queue and policy document
		testSqs := sqs.CreateQueueInput{
			QueueName: aws.String("test-queue"),
			Attributes: map[string]string{
				"Policy": string(policyStr),
			},
		}

		// Finally, attempt to create the queue; this should not fail
		output, err := client.CreateQueue(context.Background(), &testSqs)
		if err != nil {
			panic(err)
		}

		// Save the queue URL for later
		queueUrl = *output.QueueUrl
	})

	// Ensure that our test queue is cleaned up after each test
	AfterEach(func() {

		// First, create a new SQS client from our test config
		client := sqs.NewFromConfig(cfg)

		// Next, create a new delete-queue input
		testSqs := sqs.DeleteQueueInput{
			QueueUrl: aws.String(queueUrl),
		}

		// Finally, attempt to delete the queue; this should not fail
		if _, err := client.DeleteQueue(context.Background(), &testSqs); err != nil {
			panic(err)
		}
	})

	// Tests that, if the GetQueueUrl function fails, then calling the GetURL function will return an error
	It("GetURL - Queue name does not match - Error", func() {

		// First, create a new logger and discard its output
		logger := utils.NewLogger("testd", "test")
		logger.Discard()

		// Next, inject our test config and logger into a new SQS connection
		client := NewSQSConnection(cfg, logger)

		// Now, attempt to retrieve the URL of the queue; this should fail
		url, err := client.GetURL(context.Background(), "test-fail", WithAWSAccountID("568549577244"))

		// Finally, verify the error that was returned
		Expect(url).Should(BeEmpty())
		testutils.ErrorVerifier("test", "sqs", "/goutils/awssvc/sqs/conn.go", "SQSConnection", "GetURL", 137,
			testutils.InnerErrorVerifier("operation error SQS: GetQueueUrl, https response error StatusCode: "+
				"400, RequestID: 00000000-0000-0000-0000-000000000000, AWS.SimpleQueueService.NonExistentQueue: "),
			"Failed to retrieve SQS queue URL for queue \"test-fail\"", "[test] sqs.SQSConnection.GetURL "+
				"(/goutils/awssvc/sqs/conn.go 137): Failed to retrieve SQS queue URL for queue \"test-fail\", "+
				"Inner:\n\toperation error SQS: GetQueueUrl, https response error StatusCode: 400, RequestID: "+
				"00000000-0000-0000-0000-000000000000, AWS.SimpleQueueService.NonExistentQueue: .")(err.(*utils.GError))
	})

	// Tests that, if no errors occur, then calling the GetURL function will return the queue's URL
	It("GetURL - No failures - URL returned", func() {

		// First, create a new logger and discard its output
		logger := utils.NewLogger("testd", "test")
		logger.Discard()

		// Next, inject our test config and logger into a new SQS connection
		client := NewSQSConnection(cfg, logger)

		// Now, attempt to retrieve the URL of the queue; this should not fail
		url, err := client.GetURL(context.Background(), "test-queue")

		// Finally, verify the URL we retrieved
		Expect(err).ShouldNot(HaveOccurred())
		Expect(url).Should(Equal(queueUrl))
	})

	// Tests that, if the payload cannot be converted to JSON, then calling SendMessage will result in an error
	It("SendMessage - JSON marshal fails - Error", func() {

		// First, create a new logger and discard its output
		logger := utils.NewLogger("testd", "test")
		logger.Discard()

		// Next, inject our test config and logger into a new SQS connection
		client := NewSQSConnection(cfg, logger)

		// Create our data type and an instance of it that we'll attempt to send to SQS
		item := struct {
			Chan chan error
			Key  int
		}{
			Chan: make(chan error),
			Key:  42,
		}

		// Now, attempt to send the message to SQS; this should fail
		output, err := client.SendMessage(context.Background(), queueUrl, item, WithDeduplicationID("derp"),
			WithMessageGroupID("derp1"), WithMessageAttribute("derp", &types.MessageAttributeValue{DataType: aws.String("string"), StringValue: aws.String("derp")}))

		// Finally, verify the error we received
		Expect(output).Should(BeNil())
		testutils.ErrorVerifier("test", "sqs", "/goutils/awssvc/sqs/conn.go", "SQSConnection", "SendMessage", 56,
			testutils.InnerErrorVerifier("json: unsupported type: chan error"), "Failed to convert payload to JSON",
			"[test] sqs.SQSConnection.SendMessage (/goutils/awssvc/sqs/conn.go 56): Failed to convert payload to JSON, "+
				"Inner:\n\tjson: unsupported type: chan error.")(err.(*utils.GError))
	})

	// Tests that, if the the inner call to SendMessage fails, then calling SendMessage will result in an error
	It("SendMessage - Inner SendMessage fails - Error", func() {

		// First, create a new logger and discard its output
		logger := utils.NewLogger("testd", "test")
		logger.Discard()

		// Next, inject our test config and logger into a new SQS connection
		client := NewSQSConnection(cfg, logger)

		// Create our data type and an instance of it that we'll attempt to send to SQS
		item := struct {
			Key   string
			Value string
		}{
			Key:   "test-key",
			Value: "test-value",
		}

		// Now, attempt to send the message to SQS; this should fail
		output, err := client.SendMessage(context.Background(), "fail-queue", item, WithDeduplicationID("derp"),
			WithMessageGroupID("derp1"), WithMessageAttribute("derp", &types.MessageAttributeValue{DataType: aws.String("string"), StringValue: aws.String("derp")}))

		// Finally, verify the error we received
		Expect(output).Should(BeNil())
		testutils.ErrorVerifier("test", "sqs", "/goutils/awssvc/sqs/conn.go", "SQSConnection", "SendMessage", 74,
			testutils.InnerErrorVerifier("operation error SQS: SendMessage, https response error StatusCode: "+
				"400, RequestID: 00000000-0000-0000-0000-000000000000, api error AWS.SimpleQueueService."+
				"NonExistentQueue: The specified queue does not exist for this wsdl version."), "Failed "+
				"to send SQS message to \"fail-queue\"", "[test] sqs.SQSConnection.SendMessage (/goutils/awssvc/sqs/conn.go 74): "+
				"Failed to send SQS message to \"fail-queue\", Inner:\n\toperation error SQS: SendMessage, "+
				"https response error StatusCode: 400, RequestID: 00000000-0000-0000-0000-000000000000, "+
				"api error AWS.SimpleQueueService.NonExistentQueue: The specified queue does not exist "+
				"for this wsdl version..")(err.(*utils.GError))
	})

	// Tests that, if no errors occur, then calling SendMessage will send the data to SQS to be received
	It("SendMessage - No failures - Sent", func() {

		// First, create a new logger and discard its output
		logger := utils.NewLogger("testd", "test")
		logger.Discard()

		// Next, inject our test config and logger into a new SQS connection
		client := NewSQSConnection(cfg, logger)

		// Create a data type we'll send to SQS
		type Test struct {
			Key   string
			Value string
		}

		// Create an instance of our test type
		item := Test{
			Key:   "test-key",
			Value: "test-value",
		}

		// Now, attempt to send the message to SQS; this should not fail
		output, err := client.SendMessage(context.Background(), queueUrl, item, WithDeduplicationID("derp"),
			WithMessageGroupID("derp1"), WithMessageAttribute("derp", &types.MessageAttributeValue{DataType: aws.String("string"), StringValue: aws.String("derp")}))
		Expect(err).ShouldNot(HaveOccurred())

		// Create a request to receive the message we just sent
		request := sqs.ReceiveMessageInput{
			QueueUrl:              aws.String(queueUrl),
			AttributeNames:        []types.QueueAttributeName{types.QueueAttributeNameAll},
			MessageAttributeNames: []string{string(types.QueueAttributeNameAll)},
		}

		// Finally, attempt to receive the message we just sent; this should not fail
		confirm, err := client.sqs.ReceiveMessage(context.Background(), &request)
		Expect(err).ShouldNot(HaveOccurred())

		// Confirm the body of the message
		Expect(confirm.Messages).Should(HaveLen(1))
		Expect(output.MD5OfMessageAttributes).Should(Equal(confirm.Messages[0].MD5OfMessageAttributes))
		Expect(output.MD5OfMessageBody).Should(Equal(confirm.Messages[0].MD5OfBody))
		Expect(output.MessageId).Should(Equal(confirm.Messages[0].MessageId))

		// Attempt to decode the message body from base-64 string; this should not fail
		decoded, err := base64.StdEncoding.DecodeString(*confirm.Messages[0].Body)
		Expect(err).ShouldNot(HaveOccurred())

		// Attempt to deserialize the message from JSON and verify the data; this should not fail
		var check Test
		err = json.Unmarshal(decoded, &check)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(check.Key).Should(Equal("test-key"))
		Expect(check.Value).Should(Equal("test-value"))
	})

	// Tests that, if the payload cannot be converted to JSON, then calling SendMessages will result in an error
	It("SendMessages - JSON marshal fails - Error", func() {

		// First, create a new logger and discard its output
		logger := utils.NewLogger("testd", "test")
		logger.Discard()

		// Next, inject our test config and logger into a new SQS connection
		client := NewSQSConnection(cfg, logger, WithSendMessagesBatchSize(10))

		// Create a type that will define the data we're sending to SQS
		type Test struct {
			Chan chan error
			Key  int
		}

		// Create the data that we'll attempt to send to SQS
		items := []interface{}{
			Test{
				Chan: make(chan error),
				Key:  42,
			},
			Test{
				Chan: make(chan error),
				Key:  84,
			},
		}

		// Now, attempt to send the messages to SQS; this should fail
		output, err := client.SendMessages(context.Background(), queueUrl, items,
			WithBatchDeduplicationID(func(i int) string { return "derp1" }),
			WithBatchMessageGroupID(func(i int) string { return "derp2" }),
			WithBatchMessageAttribute("derp", func(i int) *types.MessageAttributeValue {
				return &types.MessageAttributeValue{DataType: aws.String("string"), StringValue: aws.String("derp3")}
			}))

		// Finally, verify the error we received
		Expect(output.Failed).Should(BeEmpty())
		Expect(output.Successful).Should(BeEmpty())
		testutils.ErrorVerifier("test", "sqs", "/goutils/awssvc/sqs/conn.go", "SQSConnection", "SendMessages", 109,
			testutils.InnerErrorVerifier("json: unsupported type: chan error"), "Failed to send page "+
				"0 of batched message to \"http://us-east-1.goaws.com:4100/100010001000/test-queue\"",
			"[test] sqs.SQSConnection.SendMessages (/goutils/awssvc/sqs/conn.go 109): Failed to send "+
				"page 0 of batched message to \"http://us-east-1.goaws.com:4100/100010001000/test-queue\", "+
				"Inner:\n\tjson: unsupported type: chan error.")(err.(*utils.GError))
	})

	// Tests that, if the the call to SendMessageBatch fails, then calling SendMessages will result in an error
	It("SendMessages - SendMessageBatch fails - Error", func() {

		// First, create a new logger and discard its output
		logger := utils.NewLogger("testd", "test")
		logger.Discard()

		// Next, inject our test config and logger into a new SQS connection
		client := NewSQSConnection(cfg, logger, WithSendMessagesBatchSize(10))

		// Create a type that will define the data we're sending to SQS
		type Test struct {
			Key   string
			Value string
		}

		// Create the data that we'll attempt to send to SQS
		items := []interface{}{
			Test{
				Key:   "test-key1",
				Value: "test-value2",
			},
			Test{
				Key:   "test-key1",
				Value: "test-value2",
			},
		}

		// Now, attempt to send the messages to SQS; this should fail
		output, err := client.SendMessages(context.Background(), "fail-queue", items,
			WithBatchDeduplicationID(func(i int) string { return "derp1" }),
			WithBatchMessageGroupID(func(i int) string { return "derp2" }),
			WithBatchMessageAttribute("derp", func(i int) *types.MessageAttributeValue {
				return &types.MessageAttributeValue{DataType: aws.String("string"), StringValue: aws.String("derp3")}
			}))

		// Finally, verify the error we received
		Expect(output.Failed).Should(BeEmpty())
		Expect(output.Successful).Should(BeEmpty())
		testutils.ErrorVerifier("test", "sqs", "/goutils/awssvc/sqs/conn.go", "SQSConnection", "SendMessages", 109,
			testutils.InnerErrorVerifier("operation error SQS: SendMessageBatch, https response error StatusCode: "+
				"400, RequestID: 00000000-0000-0000-0000-000000000000, api error AWS.SimpleQueueService."+
				"NonExistentQueue: The specified queue does not exist for this wsdl version."), "Failed "+
				"to send page 0 of batched message to \"fail-queue\"", "[test] sqs.SQSConnection.SendMessages "+
				"(/goutils/awssvc/sqs/conn.go 109): Failed to send page 0 of batched message to \"fail-queue\", "+
				"Inner:\n\toperation error SQS: SendMessageBatch, https response error StatusCode: 400, "+
				"RequestID: 00000000-0000-0000-0000-000000000000, api error AWS.SimpleQueueService.NonExistentQueue: "+
				"The specified queue does not exist for this wsdl version..")(err.(*utils.GError))
	})

	// Tests that, if no errors occur, then calling SendMessages will send the data to SQS to be received
	It("SendMessages - No failures - Sent", func() {

		// First, create a new logger and discard its output
		logger := utils.NewLogger("testd", "test")
		logger.Discard()

		// Next, inject our test config and logger into a new SQS connection
		client := NewSQSConnection(cfg, logger, WithSendMessagesBatchSize(10))

		// Create a data type we'll send to SQS
		type Test struct {
			Key   string
			Value string
		}

		// Create an instance of our test type
		items := []interface{}{
			Test{
				Key:   "test-key1",
				Value: "test-value1",
			},
			Test{
				Key:   "test-key2",
				Value: "test-value2",
			},
		}

		// Now, attempt to send the message to SQS; this should not fail
		output, err := client.SendMessages(context.Background(), queueUrl, items,
			WithBatchDeduplicationID(func(i int) string { return "derp1" }),
			WithBatchMessageGroupID(func(i int) string { return "derp2" }),
			WithBatchMessageAttribute("derp", func(i int) *types.MessageAttributeValue {
				return &types.MessageAttributeValue{DataType: aws.String("string"), StringValue: aws.String("derp3")}
			}))
		Expect(err).ShouldNot(HaveOccurred())

		// Create a request to receive the message we just sent
		request := sqs.ReceiveMessageInput{
			QueueUrl:              aws.String(queueUrl),
			AttributeNames:        []types.QueueAttributeName{types.QueueAttributeNameAll},
			MessageAttributeNames: []string{string(types.QueueAttributeNameAll)},
			MaxNumberOfMessages:   2,
		}

		// Finally, attempt to receive the message we just sent; this should not fail
		confirm, err := client.sqs.ReceiveMessage(context.Background(), &request)
		Expect(err).ShouldNot(HaveOccurred())

		// Confirm the body of the messages that were sent against those received
		Expect(confirm.Messages).Should(HaveLen(2))
		for i := 0; i < len(confirm.Messages); i++ {

			// First, verify that the message attributes, message ID and body match
			Expect(output.Successful[i].MD5OfMessageAttributes).Should(Equal(confirm.Messages[i].MD5OfMessageAttributes))
			Expect(output.Successful[i].MD5OfMessageBody).Should(Equal(confirm.Messages[i].MD5OfBody))
			Expect(output.Successful[i].MessageId).Should(Equal(confirm.Messages[i].MessageId))

			// Next, attempt to decode the message body from base-64 string; this should not fail
			decoded, err := base64.StdEncoding.DecodeString(*confirm.Messages[i].Body)
			Expect(err).ShouldNot(HaveOccurred())

			// Finally, attempt to deserialize the message from JSON and verify the data; this should not fail
			var check Test
			err = json.Unmarshal(decoded, &check)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(check.Key).Should(Equal(fmt.Sprintf("test-key%d", i+1)))
			Expect(check.Value).Should(Equal(fmt.Sprintf("test-value%d", i+1)))
		}
	})
})
