package extensions

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
)

// Create a new test runner we'll use to test all the
// modules in the extensions package
func TestExtensions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Extensions Suite")
}

var _ = Describe("ConvertibleDecimal Marshal/Unmarshal Tests", func() {

	// Test that converting the ConvertibleDecimal enum to JSON works for all values
	It("MarshalJSON Tests", func() {
		c := ConvertibleDecimal{decimal.NewFromFloat(25.99)}
		data, err := json.Marshal(c)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(string(data)).Should(Equal("\"25.99\""))
	})

	// Test that converting the ConvertibleDecimal enum to a CSV column works for all values
	It("MarshalCSV Tests", func() {
		c := ConvertibleDecimal{decimal.NewFromFloat(25.99)}
		data, err := c.MarshalCSV()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(string(data)).Should(Equal("25.99"))
	})

	// Test that converting the ConvertibleDecimal enum to a YAML node works for all values
	It("MarshalYAML - Works", func() {
		c := ConvertibleDecimal{decimal.NewFromFloat(25.99)}
		data, err := c.MarshalYAML()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(data).Should(Equal("25.99"))
	})

	// Test that converting the ConvertibleDecimal enum to a DynamoDB AttributeVAlue works for all values
	It("MarshalDynamoDBAttributeValue - Works", func() {
		c := ConvertibleDecimal{decimal.NewFromFloat(25.99)}
		data, err := attributevalue.Marshal(c)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(data.(*types.AttributeValueMemberN).Value).Should(Equal("25.99"))
	})

	// Test that converting the ConvertibleDecimal enum to an SQL value for all values
	It("Value Tests", func() {
		c := ConvertibleDecimal{decimal.NewFromFloat(25.99)}
		data, err := c.Value()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(data).Should(Equal("25.99"))
	})

	// Test that attempting to deserialize a ConvertibleDecimal will fail and return an error if the value
	// cannot be deserialized from a JSON value to a string
	It("UnmarshalJSON fails - Error", func() {

		// Attempt to convert a non-parseable string value into a ConvertibleDecimal; this should return an error
		c := new(ConvertibleDecimal)
		err := c.UnmarshalJSON([]byte("derp"))

		// Verify the error
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("error decoding string 'derp': can't convert derp to decimal: exponent is not numeric"))
	})

	// Test that attempting to deserialize a ConvertibleDecimal will fail and return an error if the value
	// cannot be converted to either the name value or integer value of the enum option
	It("UnmarshalJSON - Value is invalid - Error", func() {

		// Attempt to convert a fake string value into a ConvertibleDecimal; this should return an error
		c := new(ConvertibleDecimal)
		err := c.UnmarshalJSON([]byte("\"derp\""))

		// Verify the error
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("error decoding string 'derp': can't convert derp to decimal: exponent is not numeric"))
	})

	// Test the conditions under which values should be convertible to a ConvertibleDecimal
	It("UnmarshalJSON Tests", func() {

		// Attempt to convert the string value into a ConvertibleDecimal; this should not fail
		var c ConvertibleDecimal
		err := c.UnmarshalJSON([]byte("\"25.99\""))

		// Verify that the deserialization was successful
		Expect(err).ShouldNot(HaveOccurred())
		Expect(c.String()).Should(Equal("25.99"))
	})

	// Test that attempting to deserialize a ConvertibleDecimal will fail and return an error if the value
	// cannot be converted to either the name value or integer value of the enum option
	It("UnmarshalCSV - Value is empty - Error", func() {

		// Attempt to convert a fake string value into a ConvertibleDecimal; this should return an error
		c := new(ConvertibleDecimal)
		err := c.UnmarshalCSV("")

		// Verify the error
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("can't convert  to decimal"))
	})

	// Test the conditions under which values should be convertible to a ConvertibleDecimal
	It("UnmarshalCSV Tests", func() {

		// Attempt to convert the value into a ConvertibleDecimal; this should not fail
		var c ConvertibleDecimal
		err := c.UnmarshalCSV("25.99")

		// Verify that the deserialization was successful
		Expect(err).ShouldNot(HaveOccurred())
		Expect(c.String()).Should(Equal("25.99"))
	})

	// Test that attempting to deserialize a ConvertibleDecimal will fail and return an error if the YAML
	// node does not represent a scalar value
	It("UnmarshalYAML - Node type is not scalar - Error", func() {
		c := new(ConvertibleDecimal)
		err := c.UnmarshalYAML(&yaml.Node{Kind: yaml.AliasNode})
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("YAML node had an invalid kind (expected scalar value)"))
	})

	// Test that attempting to deserialize a ConvertibleDecimal will fail and return an error if the YAML
	// node value cannot be converted to either the name value or integer value of the enum option
	It("UnmarshalYAML - Parse fails - Error", func() {
		c := new(ConvertibleDecimal)
		err := c.UnmarshalYAML(&yaml.Node{Kind: yaml.ScalarNode, Value: "derp"})
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("can't convert derp to decimal: exponent is not numeric"))
	})

	// Test the conditions under which YAML node values should be convertible to a ConvertibleDecimal
	It("UnmarshalYAML Tests", func() {
		var c ConvertibleDecimal
		err := c.UnmarshalYAML(&yaml.Node{Kind: yaml.ScalarNode, Value: "25.99"})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(c.String()).Should(Equal("25.99"))
	})

	// Tests that, if the attribute type submitted to UnmarshalDynamoDBAttributeValue is not one we
	// recognize, then the function will return an error
	It("UnmarshalDynamoDBAttributeValue - AttributeValue type invalid - Error", func() {
		c := new(ConvertibleDecimal)
		err := attributevalue.Unmarshal(&types.AttributeValueMemberBOOL{Value: true}, &c)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("Attribute value of *types.AttributeValueMemberBOOL could not be converted to a ConvertibleDecimal"))
	})

	// Tests that, if time parsing fails, then calling UnmarshalDynamoDBAttributeValue will return an error
	It("UnmarshalDynamoDBAttributeValue - Parse fails - Error", func() {
		c := new(ConvertibleDecimal)
		err := attributevalue.Unmarshal(&types.AttributeValueMemberS{Value: "derp"}, &c)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("can't convert derp to decimal: exponent is not numeric"))
	})

	// Tests the conditions under which UnmarshalDynamoDBAttributeValue is called and no error is generated
	DescribeTable("UnmarshalDynamoDBAttributeValue - AttributeValue Conditions",
		func(value types.AttributeValue, expected ConvertibleDecimal) {
			var c ConvertibleDecimal
			err := attributevalue.Unmarshal(value, &c)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(c).Should(Equal(expected))
		},
		Entry("Value is []bytes - Works",
			&types.AttributeValueMemberB{Value: []byte("25.99")}, ConvertibleDecimal{decimal.NewFromFloat(25.99)}),
		Entry("Value is number - Works",
			&types.AttributeValueMemberN{Value: "25.99"}, ConvertibleDecimal{decimal.NewFromFloat(25.99)}),
		Entry("Value is NULL - Works", new(types.AttributeValueMemberNULL), nil),
		Entry("Value is string - Works",
			&types.AttributeValueMemberS{Value: "25.99"}, ConvertibleDecimal{decimal.NewFromFloat(25.99)}))

	// Test that attempting to deserialize a ConvertibleDecimal will fial and return an error if the value
	// cannot be converted to either the name value or integer value of the enum option
	It("Scan - Value is invalid - Error", func() {

		// Attempt to convert a fake string value into a ConvertibleDecimal; this should return an error
		c := new(ConvertibleDecimal)
		err := c.Scan("derp")

		// Verify the error
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("can't convert derp to decimal: exponent is not numeric"))
	})

	// Test the conditions under which values should be convertible to a ConvertibleDecimal
	It("Scan Tests", func() {

		// Attempt to convert the value into a ConvertibleDecimal; this should not fail
		var c ConvertibleDecimal
		err := c.Scan("25.99")

		// Verify that the deserialization was successful
		Expect(err).ShouldNot(HaveOccurred())
		Expect(c.String()).Should(Equal("25.99"))
	})
})
