package lambda

import (
	"github.com/aws/aws-lambda-go/events"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transcoding Tests", func() {

	// Tests that the AttributesToJSON function works on all possible data types
	It("AttributesToJSON - Works", func() {

		// First, create a collection of DynamoDB attributes with all our test data
		attrs := map[string]events.DynamoDBAttributeValue{
			"ss": events.NewStringSetAttribute([]string{"a", "b", "c"}),
			"l": events.NewListAttribute([]events.DynamoDBAttributeValue{
				events.NewBooleanAttribute(true), events.NewBooleanAttribute(false), events.NewBooleanAttribute(true)}),
			"null": events.NewNullAttribute(),
			"ns":   events.NewNumberSetAttribute([]string{"42", "556", "72.99", "-14"}),
			"b":    events.NewBinaryAttribute([]byte("01010101")),
			"bb":   events.NewBinarySetAttribute([][]byte{[]byte("1111"), []byte("1100"), []byte("0011"), []byte("0000")}),
			"m": events.NewMapAttribute(map[string]events.DynamoDBAttributeValue{
				"s": events.NewStringAttribute("test"),
				"n": events.NewNumberAttribute("42"),
			}),
		}

		// Next, attempt to convert this data to JSON; this should not fail
		data, err := AttributesToJSON(attrs)

		// Finally, verify the JSON data
		Expect(err).ShouldNot(HaveOccurred())
		Expect(string(data)).Should(Equal("{\"b\":\"MDEwMTAxMDE=\",\"bb\":[\"MTExMQ==\",\"MTEwMA==\"," +
			"\"MDAxMQ==\",\"MDAwMA==\"],\"l\":[true,false,true],\"m\":{\"n\":\"42\",\"s\":\"test\"}," +
			"\"ns\":[\"42\",\"556\",\"72.99\",\"-14\"],\"ss\":[\"a\",\"b\",\"c\"]}"))
	})
})
