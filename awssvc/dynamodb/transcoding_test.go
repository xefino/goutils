package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transcoding Tests", func() {

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
