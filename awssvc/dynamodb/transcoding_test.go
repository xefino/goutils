package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/xefino/goutils/awssvc/testing"
	"github.com/xefino/goutils/testutils"
	"github.com/xefino/goutils/utils"
)

var _ = Describe("Transcoding Tests", func() {

	// Contains tests for marshalling and unmarshalling DynamoDB attribute values data
	Context("Marshal/Unmarshal Tests", Ordered, func() {

		// Ensure that the AWS config is created before each test; this could be set as a global variable
		var cfg aws.Config
		BeforeAll(func() {
			cfg = testing.TestAWSConfig(context.Background(), "us-east-1", 9000)
		})

		// Tests that, if the call to MarshalMapWithOptions fails, then the MarshalMap function will return an error
		It("MarshalMap - MarshalMapWithOptions fails - Error", func() {

			// First, create an item which should fail to serialize
			item := new(failData)

			// Next, create our test database connection from our test config
			conn := createTestConnection(cfg)

			// Now, attempt to serialize the data; this should fail
			attrs, err := conn.MarshalMap(item)

			// Verify that we got no data and verify the error
			Expect(attrs).Should(BeEmpty())
			testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/transcoding.go",
				"DatabaseConnection", "MarshalMap", 17, testutils.InnerErrorVerifier("failData cannot be marshalled"),
				"Failed to marshal *dynamodb.failData to DynamoDB attributes", "[test] dynamodb.DatabaseConnection.MarshalMap "+
					"(/goutils/awssvc/dynamodb/transcoding.go 17): Failed to marshal *dynamodb.failData "+
					"to DynamoDB attributes, Inner:\n\tfailData cannot be marshalled.")(err.(*utils.GError))
		})

		// Tests that, if the call to MarshalMapWithOptions does not fail, then the MarshalMap function
		// will convert the data to a mapping between field values and DyanmoDB attribute values
		It("MarshalMap - No failures - No error", func() {

			// First, create an item which should be serialized
			item := testData{
				Key:   42,
				Value: "test-value",
			}

			// Next, create our test database connection from our test config
			conn := createTestConnection(cfg)

			// Now, attempt to serialize the data; this should not fail
			attrs, err := conn.MarshalMap(item)
			Expect(err).ShouldNot(HaveOccurred())

			// Finally, verify the attributes we created
			Expect(attrs).Should(HaveLen(2))
			Expect(attrs).Should(HaveKey("key"))
			Expect(attrs["key"].(*types.AttributeValueMemberN).Value).Should(Equal("42"))
			Expect(attrs).Should(HaveKey("value"))
			Expect(attrs["value"].(*types.AttributeValueMemberS).Value).Should(Equal("test-value"))
		})

		// Tests that, if the call to UnmarshalMapWithOptions fails, then the UnmarshalMap function will return an error
		It("UnmarshalMap - UnmarshalMapWithOptions fails - Error", func() {

			// First, create our test attributes data
			attrs := map[string]types.AttributeValue{
				"key":   &types.AttributeValueMemberN{Value: "derp"},
				"value": &types.AttributeValueMemberS{Value: "derp"},
			}

			// Next, create our test database connection from our test config
			conn := createTestConnection(cfg)

			// Now, attempt to deserialize the data; this should fail
			var data testData
			err := conn.UnmarshalMap(attrs, &data)

			// Finally, verify that we got no data and verify the error
			testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/transcoding.go", "DatabaseConnection",
				"UnmarshalMap", 27, testutils.InnerErrorVerifier("strconv.ParseInt: parsing \"derp\": invalid syntax"),
				"Failed to unmarshal DynamoDB response to *dynamodb.testData", "[test] dynamodb.DatabaseConnection.UnmarshalMap "+
					"(/goutils/awssvc/dynamodb/transcoding.go 27): Failed to unmarshal DynamoDB response "+
					"to *dynamodb.testData, Inner:\n\tstrconv.ParseInt: parsing \"derp\": invalid syntax.")(err.(*utils.GError))
		})

		// Tests that, if the call to UnmarshalMapWithOptions does not fail, then the UnmarshalMap function
		// will convert the mapping between field values and DyanmoDB attribute values to our data object
		It("UnmarshalMap - No failures - No error", func() {

			// First, create our test attributes data
			attrs := map[string]types.AttributeValue{
				"key":   &types.AttributeValueMemberN{Value: "69"},
				"value": &types.AttributeValueMemberS{Value: "derp"},
			}

			// Next, create our test database connection from our test config
			conn := createTestConnection(cfg)

			// Now, attempt to deserialize the data; this should not fail
			var data *testData
			err := conn.UnmarshalMap(attrs, &data)
			Expect(err).ShouldNot(HaveOccurred())

			// Finally, verify that the data was deserialized successfully
			Expect(data.Key).Should(Equal(69))
			Expect(data.Value).Should(Equal("derp"))
		})

		// Tests that, if the call to UnmarshalListOfMapsWithOptions fails, then the UnmarshalList function will return an error
		It("UnmarshalList - UnmarshalListOfMapsWithOptions fails - Error", func() {

			// First, create our test attributes data
			attrs := []map[string]types.AttributeValue{
				{
					"key":   &types.AttributeValueMemberN{Value: "69"},
					"value": &types.AttributeValueMemberS{Value: "derp"},
				},
				{
					"key":   &types.AttributeValueMemberN{Value: "derp"},
					"value": &types.AttributeValueMemberS{Value: "derp"},
				},
			}

			// Next, create our test database connection from our test config
			conn := createTestConnection(cfg)

			// Now, attempt to deserialize the data; this should fail
			var data []*testData
			err := conn.UnmarshalList(attrs, &data)

			// Finally, verify that we got no data and verify the error
			testutils.ErrorVerifier("test", "dynamodb", "/goutils/awssvc/dynamodb/transcoding.go", "DatabaseConnection",
				"UnmarshalList", 37, testutils.InnerErrorVerifier("strconv.ParseInt: parsing \"derp\": invalid syntax"),
				"Failed to unmarshal DynamoDB response to *[]*dynamodb.testData", "[test] dynamodb.DatabaseConnection.UnmarshalList "+
					"(/goutils/awssvc/dynamodb/transcoding.go 37): Failed to unmarshal DynamoDB response to "+
					"*[]*dynamodb.testData, Inner:\n\tstrconv.ParseInt: parsing \"derp\": invalid syntax.")(err.(*utils.GError))
		})

		// Tests that, if the call to UnmarshalListOfMapsWithOptions does not fail, then the UnmarshalList
		// function will convert the list of mappings between field values and DyanmoDB attribute values
		// to our list of data objects
		It("UnmarshalList - No failures - No error", func() {

			// First, create our test attributes data
			attrs := []map[string]types.AttributeValue{
				{
					"key":   &types.AttributeValueMemberN{Value: "69"},
					"value": &types.AttributeValueMemberS{Value: "derp"},
				},
				{
					"key":   &types.AttributeValueMemberN{Value: "70"},
					"value": &types.AttributeValueMemberS{Value: "herp"},
				},
			}

			// Next, create our test database connection from our test config
			conn := createTestConnection(cfg)

			// Now, attempt to deserialize the data; this should not fail
			var data []*testData
			err := conn.UnmarshalList(attrs, &data)
			Expect(err).ShouldNot(HaveOccurred())

			// Finally, verify that the data was deserialized successfully
			Expect(data).Should(HaveLen(2))
			Expect(data[0].Key).Should(Equal(69))
			Expect(data[0].Value).Should(Equal("derp"))
			Expect(data[1].Key).Should(Equal(70))
			Expect(data[1].Value).Should(Equal("herp"))
		})
	})

	// Tests that the AttributeValuesToJSON function works on all possible data types
	It("AttributeValuesToJSON - Works", func() {

		// First, create a collection of DynamoDB attributes with all our test data
		attrs := map[string]types.AttributeValue{
			"ss": &types.AttributeValueMemberSS{Value: []string{"a", "b", "c"}},
			"l": &types.AttributeValueMemberL{Value: []types.AttributeValue{
				&types.AttributeValueMemberBOOL{Value: true},
				&types.AttributeValueMemberBOOL{Value: false},
				&types.AttributeValueMemberBOOL{Value: true}}},
			"null": new(types.AttributeValueMemberNULL),
			"ns":   &types.AttributeValueMemberNS{Value: []string{"42", "556", "72.99", "-14"}},
			"b":    &types.AttributeValueMemberB{Value: []byte("01010101")},
			"bb":   &types.AttributeValueMemberBS{Value: [][]byte{[]byte("1111"), []byte("1100"), []byte("0011"), []byte("0000")}},
			"m": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"s": &types.AttributeValueMemberS{Value: "test"},
				"n": &types.AttributeValueMemberN{Value: "42"},
			}},
		}

		// Next, attempt to convert this data to JSON; this should not fail
		data, err := AttributeValuesToJSON(attrs)

		// Finally, verify the JSON data
		Expect(err).ShouldNot(HaveOccurred())
		Expect(string(data)).Should(Equal("{\"b\":\"MDEwMTAxMDE=\",\"bb\":[\"MTExMQ==\",\"MTEwMA==\"," +
			"\"MDAxMQ==\",\"MDAwMA==\"],\"l\":[true,false,true],\"m\":{\"n\":\"42\",\"s\":\"test\"}," +
			"\"ns\":[\"42\",\"556\",\"72.99\",\"-14\"],\"ss\":[\"a\",\"b\",\"c\"]}"))
	})
})

// Define a test type that we'll use to test marshal failures
type failData struct{}

// Mock out a DynamoDB marshal function that will just return an error
func (failData) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return nil, fmt.Errorf("failData cannot be marshalled")
}

// Define a test type we'll use to test marshal/unmarshal functionality
type testData struct {
	Key   int    `json:"key"`
	Value string `json:"value"`
}
