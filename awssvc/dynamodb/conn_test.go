package dynamodb

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/xefino/goutils/awssvc/testing"
	"github.com/xefino/goutils/testutils"
	"github.com/xefino/goutils/utils"
)

var _ = Describe("Database Connection Tests", Ordered, func() {

	// Ensure that the AWS config is created before each test; this could be set as a global variable
	var cfg aws.Config
	BeforeAll(func() {
		cfg = testing.TestAWSConfig(context.Background(), "us-east-1", 9000)
	})

	// Create our test table definition that we'll use for all module tests
	testTable := dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("sort_key"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("sort_key"),
				KeyType:       types.KeyTypeRange,
			},
		},
		TableName:   aws.String("TEST_TABLE"),
		BillingMode: types.BillingModeProvisioned,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableClass: types.TableClassStandard,
	}

	// Esnure that the table exists before the start of each test
	BeforeEach(func() {
		if err := testing.EnsureTableExists(context.Background(), cfg, &testTable); err != nil {
			panic(err)
		}
	})

	// Ensure that the table is empty at the end of each test (not strictly necessary if test data is isolated)
	AfterEach(func() {
		if err := testing.EmptyTable(context.Background(), cfg, &testTable); err != nil {
			panic(err)
		}
	})

	// Tests the conditions under which doRetry will attempt to retry a DynamoDB reques
	DescribeTable("doRetry - Retry Conditions",
		func(inner error, retried bool, verifier func(*utils.GError)) {

			// First, create our test failed connection with a logger and backoff conditions
			logger := utils.NewLogger("testd", "test")
			logger.Discard()
			conn := FromClient(&failureDynamoDBClient{err: inner}, logger,
				WithBackoffStart(1), WithBackoffEnd(5), WithBackoffMaxElapsed(10))

			// Next, attempt to retry a GetItem request until the maximum bakoff time is exceeded
			count := 0
			err := conn.doRetry(context.Background(), "TEST_TABLE", "GET", func() error {
				count++
				var inner error
				_, inner = conn.db.GetItem(context.Background(), &dynamodb.GetItemInput{
					TableName:              aws.String("TEST_TABLE"),
					ConsistentRead:         aws.Bool(false),
					ReturnConsumedCapacity: types.ReturnConsumedCapacityNone,
					Key: map[string]types.AttributeValue{
						"id":       &types.AttributeValueMemberS{Value: "test_id"},
						"sort_key": &types.AttributeValueMemberS{Value: "test|sort|key"}}})
				return inner
			})

			// Finally, verify the resulting error
			casted := err.(*Error)
			Expect(err).Should(HaveOccurred())
			Expect(casted.TableName).Should(Equal("TEST_TABLE"))
			verifier(casted.GError)
			if retried {
				Expect(count).Should(BeNumerically(">", 1))
			} else {
				Expect(count).Should(Equal(1))
			}
		},
		Entry("ProvisionedThroughputExceededException - Retried",
			&types.ProvisionedThroughputExceededException{Message: aws.String("")}, true,
			testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/conn_test.go", "glob",
				"", 85, testutils.InnerErrorVerifier("operation error : , ProvisionedThroughputExceededException: "),
				"GET request to TEST_TABLE in DynamoDB failed", "[test] dynamodb.glob. "+
					"(/goutils/awssvc/dynamodb/conn_test.go 85): GET request to TEST_TABLE in DynamoDB failed, "+
					"Inner: operation error : , ProvisionedThroughputExceededException: .")),
		Entry("RequestLimitExceeded - Retried",
			&types.RequestLimitExceeded{Message: aws.String("")}, true,
			testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/conn_test.go", "glob",
				"", 85, testutils.InnerErrorVerifier("operation error : , RequestLimitExceeded: "),
				"GET request to TEST_TABLE in DynamoDB failed", "[test] dynamodb.glob. "+
					"(/goutils/awssvc/dynamodb/conn_test.go 85): GET request to TEST_TABLE in DynamoDB failed, "+
					"Inner: operation error : , RequestLimitExceeded: .")),
		Entry("InternalServerError - Retried",
			&types.InternalServerError{Message: aws.String("")}, true,
			testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/conn_test.go", "glob",
				"", 85, testutils.InnerErrorVerifier("operation error : , InternalServerError: "),
				"GET request to TEST_TABLE in DynamoDB failed", "[test] dynamodb.glob. "+
					"(/goutils/awssvc/dynamodb/conn_test.go 85): GET request to TEST_TABLE in DynamoDB failed, "+
					"Inner: operation error : , InternalServerError: .")),
		Entry("ResourceNotFoundException - Not Retried",
			&types.ResourceNotFoundException{Message: aws.String("")}, false,
			testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/conn_test.go", "glob",
				"", 85, testutils.InnerErrorVerifier("operation error : , ResourceNotFoundException: "),
				"GET request to TEST_TABLE in DynamoDB failed", "[test] dynamodb.glob. "+
					"(/goutils/awssvc/dynamodb/conn_test.go 85): GET request to TEST_TABLE in DynamoDB failed, "+
					"Inner: operation error : , ResourceNotFoundException: .")))

	// Test that, if the inner PutItem request fails, then calling PutItem will return an error
	It("PutItem - Fails - Error", func() {

		// First, create our test database connection from our test config
		conn := createTestConnection(cfg)

		// Next, create our test object with some fake data
		data := testObject{
			ID:      "test_id",
			SortKey: "test|sort|key",
			Data:    1,
		}

		// Attempt to marshal the test object into a DynamoDB item structure
		attrs, err := attributevalue.MarshalMapWithOptions(&data,
			func(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Now, create our put-item input from our attribute data
		input := dynamodb.PutItemInput{
			Item:                        attrs,
			TableName:                   aws.String("FAKE_TABLE"),
			ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
			ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone,
			ReturnValues:                types.ReturnValueAllNew,
		}

		// Finally, attempt to put the item to the database; this should fail
		output, err := conn.PutItem(context.Background(), &input)

		// Verify the failure
		casted := err.(*Error)
		Expect(output).Should(BeNil())
		Expect(err).Should(HaveOccurred())
		Expect(casted.TableName).Should(Equal("FAKE_TABLE"))
		testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/conn.go", "DatabaseConnection",
			"PutItem", 58, testutils.InnerErrorPrefixSuffixVerifier("operation error DynamoDB: PutItem, "+
				"https response error StatusCode: 400, RequestID: ", ", ResourceNotFoundException: "),
			"PUT request to FAKE_TABLE in DynamoDB failed", "[test] dynamodb.DatabaseConnection.PutItem "+
				"(/goutils/awssvc/dynamodb/conn.go 58): PUT request to FAKE_TABLE in DynamoDB failed, Inner: "+
				"operation error DynamoDB: PutItem, https response error StatusCode: 400, RequestID: ",
			", ResourceNotFoundException: .")(casted.GError)
	})

	// Test that, if no failure occurs, then calling PutItem will result in the item being written
	// to the associated table in the database
	It("PutItem - No failures - Data exists", func() {

		// First, create our test database connection from our test config
		conn := createTestConnection(cfg)

		// Create our test object with some fake data
		data := testObject{
			ID:      "test_id",
			SortKey: "test|sort|key",
			Data:    1,
		}

		// Attempt to marshal the test object into a DynamoDB item structure
		attrs, err := attributevalue.MarshalMapWithOptions(&data,
			func(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Next, create our put-item input from our attribute data
		input := dynamodb.PutItemInput{
			Item:                        attrs,
			TableName:                   aws.String("TEST_TABLE"),
			ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
			ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone,
			ReturnValues:                types.ReturnValueNone,
		}

		// Now, attempt to put the item to the database; this should not fail
		_, err = conn.PutItem(context.Background(), &input)
		Expect(err).ShouldNot(HaveOccurred())

		// Finally, attempt to retrieve the item as it exists in the database; this should not fail
		gOut, err := conn.db.GetItem(context.Background(), &dynamodb.GetItemInput{
			TableName: aws.String("TEST_TABLE"),
			Key: map[string]types.AttributeValue{
				"id":       &types.AttributeValueMemberS{Value: "test_id"},
				"sort_key": &types.AttributeValueMemberS{Value: "test|sort|key"}}})
		Expect(err).ShouldNot(HaveOccurred())

		// Attempt to extract our test object from the response
		var read *testObject
		err = attributevalue.UnmarshalMapWithOptions(gOut.Item, &read,
			func(eo *attributevalue.DecoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Verify the data on the test object
		Expect(read.ID).Should(Equal("test_id"))
		Expect(read.SortKey).Should(Equal("test|sort|key"))
		Expect(read.Data).Should(Equal(1))
	})

	// Test that, if the inner GetItem request fails, then calling GetItem will return an error
	It("GetItem - Fails - Error", func() {

		// First, create our test database connection from our test config
		conn := createTestConnection(cfg)

		// Create our test object with some fake data
		data := testObject{
			ID:      "test_id",
			SortKey: "test|sort|key",
			Data:    1,
		}

		// Attempt to marshal the test object into a DynamoDB item structure
		attrs, err := attributevalue.MarshalMapWithOptions(&data,
			func(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Next, attempt to write a test object to the database; this should not fail
		_, err = conn.db.PutItem(context.Background(), &dynamodb.PutItemInput{
			Item:                        attrs,
			TableName:                   aws.String("TEST_TABLE"),
			ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
			ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone})
		Expect(err).ShouldNot(HaveOccurred())

		// Now, attempt to get an item from the database; this should fail
		output, err := conn.GetItem(context.Background(), &dynamodb.GetItemInput{
			TableName:              aws.String("FAKE_TABLE"),
			ConsistentRead:         aws.Bool(false),
			ReturnConsumedCapacity: types.ReturnConsumedCapacityNone,
			Key: map[string]types.AttributeValue{
				"id":       &types.AttributeValueMemberS{Value: "test_id"},
				"sort_key": &types.AttributeValueMemberS{Value: "test|sort|key"},
			}})

		// Finally, verify the details of the error
		casted := err.(*Error)
		Expect(output).Should(BeNil())
		Expect(err).Should(HaveOccurred())
		Expect(casted.TableName).Should(Equal("FAKE_TABLE"))
		testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/conn.go", "DatabaseConnection",
			"GetItem", 74, testutils.InnerErrorPrefixSuffixVerifier("operation error DynamoDB: GetItem, "+
				"https response error StatusCode: 400, RequestID: ", ", ResourceNotFoundException: "),
			"GET request to FAKE_TABLE in DynamoDB failed", "[test] dynamodb.DatabaseConnection.GetItem "+
				"(/goutils/awssvc/dynamodb/conn.go 74): GET request to FAKE_TABLE in DynamoDB failed, Inner: "+
				"operation error DynamoDB: GetItem, https response error StatusCode: 400, RequestID: ",
			", ResourceNotFoundException: .")(casted.GError)
	})

	// Test that, if no failure occurs, then calling GetItem will result in the item being read
	// from the associated table in the database
	It("GetItem - No failures - Data returned", func() {

		// First, create our test database connection from our test config
		conn := createTestConnection(cfg)

		// Create our test object with some fake data
		data := testObject{
			ID:      "test_id",
			SortKey: "test|sort|key",
			Data:    1,
		}

		// Attempt to marshal the test object into a DynamoDB item structure
		attrs, err := attributevalue.MarshalMapWithOptions(&data,
			func(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Next, attempt to write a test object to the database; this should not fail
		_, err = conn.db.PutItem(context.Background(), &dynamodb.PutItemInput{
			Item:                        attrs,
			TableName:                   aws.String("TEST_TABLE"),
			ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
			ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone})
		Expect(err).ShouldNot(HaveOccurred())

		// Now, attempt to get an item from the database; this should not fail
		output, err := conn.GetItem(context.Background(), &dynamodb.GetItemInput{
			TableName:              aws.String("TEST_TABLE"),
			ConsistentRead:         aws.Bool(false),
			ReturnConsumedCapacity: types.ReturnConsumedCapacityNone,
			Key: map[string]types.AttributeValue{
				"id":       &types.AttributeValueMemberS{Value: "test_id"},
				"sort_key": &types.AttributeValueMemberS{Value: "test|sort|key"}}})
		Expect(err).ShouldNot(HaveOccurred())

		// Finally, unmarshal the output response into a test object
		var written *testObject
		err = attributevalue.UnmarshalMapWithOptions(output.Item, &written,
			func(eo *attributevalue.DecoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Verify the data on the test object
		Expect(written.ID).Should(Equal("test_id"))
		Expect(written.SortKey).Should(Equal("test|sort|key"))
		Expect(written.Data).Should(Equal(1))
	})

	// Test that, if the inner UpdateItem request fails, then calling UpdateItem will return an error
	It("UpdateItem - Fails - Error", func() {

		// First, create our test database connection from our test config
		conn := createTestConnection(cfg)

		// Create our test object with some fake data
		data := testObject{
			ID:      "test_id",
			SortKey: "test|sort|key",
			Data:    1,
		}

		// Attempt to marshal the test object into a DynamoDB item structure
		attrs, err := attributevalue.MarshalMapWithOptions(&data,
			func(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Next, attempt to write a test object to the database; this should not fail
		_, err = conn.db.PutItem(context.Background(), &dynamodb.PutItemInput{
			Item:                        attrs,
			TableName:                   aws.String("TEST_TABLE"),
			ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
			ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone})
		Expect(err).ShouldNot(HaveOccurred())

		// Now, attempt to update the item we just put into the table; this should fail
		output, err := conn.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
			TableName:                   aws.String("FAKE_TABLE"),
			UpdateExpression:            aws.String("SET data = :val"),
			ExpressionAttributeValues:   map[string]types.AttributeValue{":val": &types.AttributeValueMemberS{Value: "test2"}},
			ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
			ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone,
			ReturnValues:                types.ReturnValueAllNew,
			Key: map[string]types.AttributeValue{
				"id":       &types.AttributeValueMemberS{Value: "test_id"},
				"sort_key": &types.AttributeValueMemberS{Value: "test|sort|key"}}})

		// Finally, verify the details of the error
		casted := err.(*Error)
		Expect(output).Should(BeNil())
		Expect(err).Should(HaveOccurred())
		Expect(casted.TableName).Should(Equal("FAKE_TABLE"))
		testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/conn.go", "DatabaseConnection",
			"UpdateItem", 90, testutils.InnerErrorPrefixSuffixVerifier("operation error DynamoDB: UpdateItem, "+
				"https response error StatusCode: 400, RequestID: ", ", ResourceNotFoundException: "),
			"UPDATE request to FAKE_TABLE in DynamoDB failed", "[test] dynamodb.DatabaseConnection.UpdateItem "+
				"(/goutils/awssvc/dynamodb/conn.go 90): UPDATE request to FAKE_TABLE in DynamoDB failed, Inner: "+
				"operation error DynamoDB: UpdateItem, https response error StatusCode: 400, RequestID: ",
			", ResourceNotFoundException: .")(casted.GError)
	})

	// Test that, if no failure occurs, then calling UpdateItem will result in the item being
	// updated in the associated table in the database
	It("UpdateItem - No failures - Data updated", func() {

		// First, create our test database connection from our test config
		conn := createTestConnection(cfg)

		// Create our test object with some fake data
		data := testObject{
			ID:      "test_id",
			SortKey: "test|sort|key",
			Data:    1,
		}

		// Attempt to marshal the test object into a DynamoDB item structure
		attrs, err := attributevalue.MarshalMapWithOptions(&data,
			func(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Next, attempt to write a test object to the database; this should not fail
		_, err = conn.db.PutItem(context.Background(), &dynamodb.PutItemInput{
			Item:                        attrs,
			TableName:                   aws.String("TEST_TABLE"),
			ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
			ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone})
		Expect(err).ShouldNot(HaveOccurred())

		// Now, attempt to update the item we just put into the table; this should not fail
		output, err := conn.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
			TableName:                   aws.String("TEST_TABLE"),
			UpdateExpression:            aws.String("SET #d = :val"),
			ExpressionAttributeNames:    map[string]string{"#d": "data"},
			ExpressionAttributeValues:   map[string]types.AttributeValue{":val": &types.AttributeValueMemberN{Value: "2"}},
			ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
			ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone,
			ReturnValues:                types.ReturnValueAllNew,
			Key: map[string]types.AttributeValue{
				"id":       &types.AttributeValueMemberS{Value: "test_id"},
				"sort_key": &types.AttributeValueMemberS{Value: "test|sort|key"}}})
		Expect(err).ShouldNot(HaveOccurred())

		// Finally, unmarshal the output response into a test object
		var updated *testObject
		err = attributevalue.UnmarshalMapWithOptions(output.Attributes, &updated,
			func(eo *attributevalue.DecoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Verify the data on the test object
		Expect(updated.ID).Should(Equal("test_id"))
		Expect(updated.SortKey).Should(Equal("test|sort|key"))
		Expect(updated.Data).Should(Equal(2))
	})

	// Test that, if the inner DeleteItem request fails, then calling DeleteItem will return an error
	It("DeleteItem - Fails - Error", func() {

		// First, create our test database connection from our test config
		conn := createTestConnection(cfg)

		// Create our test object with some fake data
		data := testObject{
			ID:      "test_id",
			SortKey: "test|sort|key",
			Data:    1,
		}

		// Attempt to marshal the test object into a DynamoDB item structure
		attrs, err := attributevalue.MarshalMapWithOptions(&data,
			func(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Next, attempt to write a test object to the database; this should not fail
		_, err = conn.db.PutItem(context.Background(), &dynamodb.PutItemInput{
			Item:                        attrs,
			TableName:                   aws.String("TEST_TABLE"),
			ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
			ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone})
		Expect(err).ShouldNot(HaveOccurred())

		// Now, attempt to delete the item we just put into the table; this should fail
		output, err := conn.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
			TableName:                   aws.String("FAKE_TABLE"),
			ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
			ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone,
			ReturnValues:                types.ReturnValueNone,
			Key: map[string]types.AttributeValue{
				"id":       &types.AttributeValueMemberS{Value: "test_id"},
				"sort_key": &types.AttributeValueMemberS{Value: "test|sort|key"},
			},
		})

		// Finally, verify the details of the error
		casted := err.(*Error)
		Expect(output).Should(BeNil())
		Expect(err).Should(HaveOccurred())
		Expect(casted.TableName).Should(Equal("FAKE_TABLE"))
		testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/conn.go", "DatabaseConnection",
			"DeleteItem", 106, testutils.InnerErrorPrefixSuffixVerifier("operation error DynamoDB: DeleteItem, "+
				"https response error StatusCode: 400, RequestID: ", ", ResourceNotFoundException: "),
			"DELETE request to FAKE_TABLE in DynamoDB failed", "[test] dynamodb.DatabaseConnection.DeleteItem "+
				"(/goutils/awssvc/dynamodb/conn.go 106): DELETE request to FAKE_TABLE in DynamoDB failed, Inner: "+
				"operation error DynamoDB: DeleteItem, https response error StatusCode: 400, RequestID: ",
			", ResourceNotFoundException: .")(casted.GError)
	})

	// Test that, if no failure occurs, then calling DeleteItem will result in the item being
	// removed from the associated table in the database
	It("DeleteItem - No failures - Data removed", func() {

		// First, create our test database connection from our test config
		conn := createTestConnection(cfg)

		// Create our test object with some fake data
		data := testObject{
			ID:      "test_id",
			SortKey: "test|sort|key",
			Data:    1,
		}

		// Attempt to marshal the test object into a DynamoDB item structure
		attrs, err := attributevalue.MarshalMapWithOptions(&data,
			func(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Next, attempt to write a test object to the database; this should not fail
		_, err = conn.db.PutItem(context.Background(), &dynamodb.PutItemInput{
			Item:                        attrs,
			TableName:                   aws.String("TEST_TABLE"),
			ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
			ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone})
		Expect(err).ShouldNot(HaveOccurred())

		// Now, attempt to delete the item we just put into the table; this should not fail
		dOut, err := conn.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
			TableName:                   aws.String("TEST_TABLE"),
			ReturnConsumedCapacity:      types.ReturnConsumedCapacityNone,
			ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsNone,
			ReturnValues:                types.ReturnValueAllOld,
			Key: map[string]types.AttributeValue{
				"id":       &types.AttributeValueMemberS{Value: "test_id"},
				"sort_key": &types.AttributeValueMemberS{Value: "test|sort|key"}}})
		Expect(err).ShouldNot(HaveOccurred())

		// Attempt to unmarshal the output response into a test object
		var updated *testObject
		err = attributevalue.UnmarshalMapWithOptions(dOut.Attributes, &updated,
			func(eo *attributevalue.DecoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Verify the data on the test object
		Expect(updated.ID).Should(Equal("test_id"))
		Expect(updated.SortKey).Should(Equal("test|sort|key"))
		Expect(updated.Data).Should(Equal(1))

		// Finally, attempt to get the deleted item from the table; this should return no data
		gOut, err := conn.db.GetItem(context.Background(), &dynamodb.GetItemInput{
			TableName:              aws.String("TEST_TABLE"),
			ReturnConsumedCapacity: types.ReturnConsumedCapacityNone,
			ConsistentRead:         aws.Bool(true),
			Key: map[string]types.AttributeValue{
				"id":       &types.AttributeValueMemberS{Value: "test_id"},
				"sort_key": &types.AttributeValueMemberS{Value: "test|sort|key"}}})
		Expect(err).ShouldNot(HaveOccurred())

		// Verify that we retrieved no data
		Expect(gOut.Item).Should(BeEmpty())
	})

	// Test that, if the BatchWriteItem request fails, then calling BatchWrite will return an error
	It("BatchWrite - Fails - Error", func() {

		// First, create our test database connection from our test config
		conn := createTestConnection(cfg)

		// Next, create our test object with some fake data
		data := testObject{
			ID:      "test_id",
			SortKey: "test|sort|key",
			Data:    1,
		}

		// Now, attempt to marshal the test object into a DynamoDB item structure
		attrs, err := attributevalue.MarshalMapWithOptions(&data,
			func(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Finally, attempt to batch-write our requests to DynamoDB; this should fail
		err = conn.BatchWrite(context.Background(), "FAKE_TABLE",
			types.WriteRequest{PutRequest: &types.PutRequest{Item: attrs}})

		// Verify the details of the error
		casted := err.(*Error)
		Expect(err).Should(HaveOccurred())
		Expect(casted.TableName).Should(Equal("FAKE_TABLE"))
		testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/conn.go", "DatabaseConnection",
			"batchWriteInner", 215, testutils.InnerErrorPrefixSuffixVerifier("operation error DynamoDB: BatchWriteItem, "+
				"https response error StatusCode: 400, RequestID: ", ", ResourceNotFoundException: "),
			"BATCH WRITE request to FAKE_TABLE in DynamoDB failed", "[test] dynamodb.DatabaseConnection.batchWriteInner "+
				"(/goutils/awssvc/dynamodb/conn.go 215): BATCH WRITE request to FAKE_TABLE in DynamoDB failed, Inner: "+
				"operation error DynamoDB: BatchWriteItem, https response error StatusCode: 400, RequestID: ",
			", ResourceNotFoundException: .")(casted.GError)
	})

	// Test that, if no failure occurs, then calling BatchWrite will result in the items being written
	// to the table in DynamoDB
	It("BatchWrite - No failures - Data written", func() {

		// First, create our test database connection from our test config
		conn := createTestConnection(cfg)

		// Next, create our list of write requests
		requests := make([]types.WriteRequest, 47)
		for i := 0; i < 47; i++ {

			// First, create our test object with some fake data
			data := testObject{
				ID:      "test_id",
				SortKey: fmt.Sprintf("test|sort|key|%d", i),
				Data:    i,
			}

			// Next, attempt to marshal the test object into a DynamoDB item structure
			attrs, err := attributevalue.MarshalMapWithOptions(&data,
				func(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" })
			Expect(err).ShouldNot(HaveOccurred())

			// Finally, create the write request and add it to our list of such requests
			requests[i] = types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: attrs,
				},
			}
		}

		// Now, attempt to batch-write our requests to DynamoDB; this should not fail
		err := conn.BatchWrite(context.Background(), "TEST_TABLE", requests...)
		Expect(err).ShouldNot(HaveOccurred())

		// Finally, do a query to retrieve all the items we wrote; this should not fail
		output, err := conn.db.Query(context.Background(), &dynamodb.QueryInput{
			TableName:                aws.String("TEST_TABLE"),
			ConsistentRead:           aws.Bool(true),
			KeyConditionExpression:   aws.String("#id = :id AND begins_with(#sk, :sk)"),
			ExpressionAttributeNames: map[string]string{"#id": "id", "#sk": "sort_key"},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":id": &types.AttributeValueMemberS{Value: "test_id"},
				":sk": &types.AttributeValueMemberS{Value: "test|sort|key|"}}})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(output.Count).Should(Equal(int32(47)))

		// Attempt to deserialize the results into a list of test objects; this should not fail
		var results []*testObject
		err = attributevalue.UnmarshalListOfMapsWithOptions(output.Items, &results,
			func(eo *attributevalue.DecoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Sort the results by sort key to ensure that the test either passes or fails deterministicallyu
		sort.Slice(results, func(i, j int) bool {
			return results[i].Data < results[j].Data
		})

		// Verify the results that we retrieved with the query
		for i := 0; i < 47; i++ {
			Expect(results[i].ID).Should(Equal("test_id"))
			Expect(results[i].SortKey).Should(Equal(fmt.Sprintf("test|sort|key|%d", i)))
			Expect(results[i].Data).Should(Equal(i))
		}
	})

	// Test that, if the inner Query request fails, then calling Query will return an error
	It("Query - Fails - Error", func() {

		// First, create our test database connection from our test config
		conn := createTestConnection(cfg)

		// Next, attempt to query items from the table; this should fail
		items, err := conn.Query(context.Background(), &dynamodb.QueryInput{
			TableName:                aws.String("FAKE_TABLE"),
			ConsistentRead:           aws.Bool(true),
			KeyConditionExpression:   aws.String("#id = :id AND begins_with(#sk, :sk)"),
			ExpressionAttributeNames: map[string]string{"#id": "id", "#sk": "sort_key"},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":id": &types.AttributeValueMemberS{Value: "test_id"},
				":sk": &types.AttributeValueMemberS{Value: "test|sort|key|"},
			},
		})

		// Finally, verify the failure
		casted := err.(*Error)
		Expect(items).Should(BeEmpty())
		Expect(err).Should(HaveOccurred())
		Expect(casted.TableName).Should(Equal("FAKE_TABLE"))
		testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/conn.go", "DatabaseConnection",
			"Query", 170, testutils.InnerErrorPrefixSuffixVerifier("operation error DynamoDB: Query, "+
				"https response error StatusCode: 400, RequestID: ", ", ResourceNotFoundException: "),
			"QUERY(0) request to FAKE_TABLE in DynamoDB failed", "[test] dynamodb.DatabaseConnection.Query "+
				"(/goutils/awssvc/dynamodb/conn.go 170): QUERY(0) request to FAKE_TABLE in DynamoDB failed, Inner: "+
				"operation error DynamoDB: Query, https response error StatusCode: 400, RequestID: ",
			", ResourceNotFoundException: .")(casted.GError)
	})

	// Test that, if no failure occurs, then calling Query will result in the items being retrieved
	// from the table in DynamoDB
	It("Query - No failures - Data returned", func() {

		// First, create our test database connection from our test config
		conn := createTestConnection(cfg)

		// Next, create our list of write requests
		size := 50
		requests := make([]types.WriteRequest, size)
		for i := 0; i < size; i++ {

			// First, create our test object with some fake data
			data := testObject{
				ID:      "test_id",
				SortKey: fmt.Sprintf("test|sort|key|%d", i),
				Data:    i,
			}

			// Next, attempt to marshal the test object into a DynamoDB item structure
			attrs, err := attributevalue.MarshalMapWithOptions(&data,
				func(eo *attributevalue.EncoderOptions) { eo.TagKey = "json" })
			Expect(err).ShouldNot(HaveOccurred())

			// Finally, create the write request and add it to our list of such requests
			requests[i] = types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: attrs,
				},
			}
		}

		// Now, attempt to batch-write our requests to DynamoDB; this should not fail
		err := conn.BatchWrite(context.Background(), "TEST_TABLE", requests...)
		Expect(err).ShouldNot(HaveOccurred())

		// Finally, attempt to query the data in DynamoDB; this should not fail
		items, err := conn.Query(context.Background(), &dynamodb.QueryInput{
			TableName:                aws.String("TEST_TABLE"),
			ConsistentRead:           aws.Bool(true),
			Limit:                    aws.Int32(25),
			KeyConditionExpression:   aws.String("#id = :id AND begins_with(#sk, :sk)"),
			ExpressionAttributeNames: map[string]string{"#id": "id", "#sk": "sort_key"},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":id": &types.AttributeValueMemberS{Value: "test_id"},
				":sk": &types.AttributeValueMemberS{Value: "test|sort|key|"}}})

		// Attempt to deserialize the results into a list of test objects; this should not fail
		var results []*testObject
		err = attributevalue.UnmarshalListOfMapsWithOptions(items, &results,
			func(eo *attributevalue.DecoderOptions) { eo.TagKey = "json" })
		Expect(err).ShouldNot(HaveOccurred())

		// Sort the results by sort key to ensure that the test either passes or fails deterministicallyu
		sort.Slice(results, func(i, j int) bool {
			return results[i].Data < results[j].Data
		})

		// Verify the results that we retrieved with the query
		for i := 0; i < size; i++ {
			Expect(results[i].ID).Should(Equal("test_id"))
			Expect(results[i].SortKey).Should(Equal(fmt.Sprintf("test|sort|key|%d", i)))
			Expect(results[i].Data).Should(Equal(i))
		}
	})
})

// Helper function that creates a test connection from an AWS config for
func createTestConnection(cfg aws.Config) *DatabaseConnection {
	logger := utils.NewLogger("testd", "test")
	logger.Discard()
	return NewDatabaseConnection(cfg, logger,
		WithBackoffStart(1), WithBackoffEnd(5), WithBackoffMaxElapsed(10))
}

// Helper type that is used to test various error conditions as returned by DynamoDB
type failureDynamoDBClient struct {
	DynamoDBAPI
	err error
}

// Mocks out the GetItem function so that it returns an error similar to what the actual AWS
// operation would generate
func (client *failureDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput,
	optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return nil, &smithy.OperationError{
		Err: client.err,
	}
}

// Helper type that we'll use to test DynamoDB functionality
type testObject struct {
	ID      string `json:"id"`
	SortKey string `json:"sort_key"`
	Data    int    `json:"data"`
}
