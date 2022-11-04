package time

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Convertible Tests", func() {

	// Tests that the MarshalJSON function works
	It("MarshalJSON - Works", func() {
		data, err := json.Marshal(WithLayout(time.Date(2022, time.November, 4, 13, 37, 0, 0, time.UTC), time.RFC822))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(string(data)).Should(Equal("\"04 Nov 22 13:37 UTC\""))
	})

	// Tests that the MarshalCSV function works
	It("MarshalCSV - Works", func() {
		data, err := WithLayout(time.Date(2022, time.November, 4, 13, 37, 0, 0, time.UTC), "").MarshalCSV()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(data).Should(Equal("2022-11-04T13:37:00Z"))
	})

	// Tests that the MarshalDynamoDBAttributeValue function works
	It("MarshalDynamoDBAttributeValue - Works", func() {
		data, err := attributevalue.Marshal(WithLayout(time.Date(2022, time.November, 4, 13, 37, 0, 0, time.UTC), time.RFC3339))
		Expect(err).ShouldNot(HaveOccurred())
		Expect(data.(*types.AttributeValueMemberS).Value).Should(Equal("2022-11-04T13:37:00Z"))
	})

	// Tests that the Value function works
	It("Value - Works", func() {
		data, err := WithLayout(time.Date(2022, time.November, 4, 13, 37, 0, 0, time.UTC), time.RFC3339).Value()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(data).Should(Equal("2022-11-04T13:37:00Z"))
	})

	// Tests that, if time parsing fails, then calling UnmarshalJSON will return an error
	It("UnmarshalJSON - Parse fails - Error", func() {
		convertible := Convertible{Layout: time.RFC3339}
		err := json.Unmarshal([]byte("\"derp\""), &convertible)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("Failed to parse \"derp\" to time, error: parsing time \"derp\" " +
			"as \"2006-01-02T15:04:05Z07:00\": cannot parse \"derp\" as \"2006\""))
	})

	// Tests that, if time parsing does not fail, then calling UnmarshalJSON will parse the data and
	// assign it to the Time field on the Convertible
	It("UnmarshalJSON - No failures, Quoted - Parsed", func() {
		convertible := Convertible{Layout: time.RFC3339}
		err := json.Unmarshal([]byte("\"2022-11-04T13:37:00Z\""), &convertible)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(convertible.Time).Should(Equal(time.Date(2022, time.November, 4, 13, 37, 0, 0, time.UTC)))
	})

	// Tests that, if time parsing fails, then calling UnmarshalCSV will return an error
	It("UnmarshalCSV - Parse fails - Error", func() {
		convertible := Convertible{Layout: time.RFC3339}
		err := convertible.UnmarshalCSV("derp")
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("Failed to parse \"derp\" to time, error: parsing time \"derp\" " +
			"as \"2006-01-02T15:04:05Z07:00\": cannot parse \"derp\" as \"2006\""))
	})

	// Tests that, if time parsing does not fail, then calling UnmarshalCSV will parse the data and
	// assign it to the Time field on the Convertible
	It("UnmarshalCSV - No failures - Parsed", func() {
		convertible := Convertible{Layout: time.RFC3339}
		err := convertible.UnmarshalCSV("2022-11-04T13:37:00Z")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(convertible.Time).Should(Equal(time.Date(2022, time.November, 4, 13, 37, 0, 0, time.UTC)))
	})

	// Tests that, if the attribute type submitted to UnmarshalDynamoDBAttributeValue is not one we
	// recognize, then the function will return an error
	It("UnmarshalDynamoDBAttributeValue - AttributeValue type invalid - Error", func() {
		convertible := Convertible{Layout: time.RFC3339}
		err := attributevalue.Unmarshal(&types.AttributeValueMemberN{Value: "42"}, &convertible)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("Attribute value of *types.AttributeValueMemberN could not be converted to a time.Time"))
	})

	// Tests that, if time parsing fails, then calling UnmarshalDynamoDBAttributeValue will return an error
	It("UnmarshalDynamoDBAttributeValue - Parse fails - Error", func() {
		convertible := Convertible{Layout: time.RFC3339}
		err := attributevalue.Unmarshal(&types.AttributeValueMemberS{Value: "derp"}, &convertible)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("Failed to parse \"derp\" to time, error: parsing time \"derp\" " +
			"as \"2006-01-02T15:04:05Z07:00\": cannot parse \"derp\" as \"2006\""))
	})

	// Tests the conditions under which UnmarshalDynamoDBAttributeValue is called and no error is generated
	DescribeTable("UnmarshalDynamoDBAttributeValue - AttributeValue Conditions",
		func(value types.AttributeValue, expected time.Time) {
			convertible := Convertible{Layout: time.RFC3339}
			err := attributevalue.Unmarshal(value, &convertible)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(convertible.Time).Should(Equal(expected))
		},
		Entry("Value is NULL - Works", new(types.AttributeValueMemberNULL), time.Time{}),
		Entry("Value is []byte - Works", &types.AttributeValueMemberB{Value: []byte("2022-11-04T13:37:00Z")},
			time.Date(2022, time.November, 4, 13, 37, 0, 0, time.UTC)),
		Entry("Value is string - Works", &types.AttributeValueMemberS{Value: "2022-11-04T13:37:00Z"},
			time.Date(2022, time.November, 4, 13, 37, 0, 0, time.UTC)))

	// Tests that, if time parsing fails, then calling Scan will return an error
	It("Scan - Parse fails - Error", func() {
		convertible := Convertible{Layout: time.RFC3339}
		err := convertible.Scan("derp")
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("Failed to parse \"derp\" to time, error: parsing time \"derp\" " +
			"as \"2006-01-02T15:04:05Z07:00\": cannot parse \"derp\" as \"2006\""))
	})

	// Tests that, if time parsing does not fail, then calling Scan will parse the data and assign
	// it to the Time field on the Convertible
	It("Scan - No failures - Parsed", func() {
		convertible := Convertible{Layout: time.RFC3339}
		err := convertible.Scan("2022-11-04T13:37:00Z")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(convertible.Time).Should(Equal(time.Date(2022, time.November, 4, 13, 37, 0, 0, time.UTC)))
	})
})
