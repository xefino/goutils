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
		testutils.ErrorVerifier("test", "sqs", "/goutils/awssvc/sqs/conn.go", "SQSConnection", "GetURL", 87,
			testutils.InnerErrorVerifier("operation error SQS: GetQueueUrl, https response error StatusCode: "+
				"400, RequestID: 00000000-0000-0000-0000-000000000000, AWS.SimpleQueueService.NonExistentQueue: "),
			"Failed to retrieve SQS queue URL for queue \"test-fail\"", "[test] sqs.SQSConnection.GetURL "+
				"(/goutils/awssvc/sqs/conn.go 87): Failed to retrieve SQS queue URL for queue \"test-fail\", "+
				"Inner: operation error SQS: GetQueueUrl, https response error StatusCode: 400, RequestID: "+
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
		testutils.ErrorVerifier("test", "sqs", "/goutils/awssvc/sqs/conn.go", "SQSConnection", "SendMessage", 45,
			testutils.InnerErrorVerifier("json: unsupported type: chan error"), "Failed to convert payload to JSON",
			"[test] sqs.SQSConnection.SendMessage (/goutils/awssvc/sqs/conn.go 45): Failed to convert payload to JSON, "+
				"Inner: json: unsupported type: chan error.")(err.(*utils.GError))
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
		testutils.ErrorVerifier("test", "sqs", "/goutils/awssvc/sqs/conn.go", "SQSConnection", "SendMessage", 63,
			testutils.InnerErrorVerifier("operation error SQS: SendMessage, https response error StatusCode: "+
				"400, RequestID: 00000000-0000-0000-0000-000000000000, api error AWS.SimpleQueueService."+
				"NonExistentQueue: The specified queue does not exist for this wsdl version."), "Failed "+
				"to send SQS message to \"fail-queue\"", "[test] sqs.SQSConnection.SendMessage (/goutils/awssvc/sqs/conn.go 63): "+
				"Failed to send SQS message to \"fail-queue\", Inner: operation error SQS: SendMessage, "+
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
		fmt.Printf("Error: %v\n", err)
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
})
