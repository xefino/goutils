package dynamodb

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// MarshalMap converts the object to a mapping of DynamoDB attribute values
func (conn *DatabaseConnection) MarshalMap(in interface{}) (map[string]types.AttributeValue, error) {
	attrs, err := attributevalue.MarshalMapWithOptions(in,
		func(options *attributevalue.EncoderOptions) { options.TagKey = conn.tagKey })
	if err != nil {
		return nil, conn.logger.FrameError(2, err, "Failed to marshal %T to DynamoDB attributes", in)
	}

	return attrs, nil
}

// UnmarshalMap converts a mapping of DynamoDB attribute values to an object
func (conn *DatabaseConnection) UnmarshalMap(attrs map[string]types.AttributeValue, out interface{}) error {
	if err := attributevalue.UnmarshalMapWithOptions(attrs, out,
		func(options *attributevalue.DecoderOptions) { options.TagKey = conn.tagKey }); err != nil {
		return conn.logger.FrameError(2, err, "Failed to unmarshal DynamoDB response to %T", out)
	}

	return nil
}

// UnmarshalList converts a list of DynamoDB attribute value mappings to a list of objects
func (conn *DatabaseConnection) UnmarshalList(attrs []map[string]types.AttributeValue, out interface{}) error {
	if err := attributevalue.UnmarshalListOfMapsWithOptions(attrs, out,
		func(options *attributevalue.DecoderOptions) { options.TagKey = conn.tagKey }); err != nil {
		return conn.logger.FrameError(2, err, "Failed to unmarshal DynamoDB response to %T", out)
	}

	return nil
}

// AttributeValuesToJSON attempts to convert a mapping of attribute values to a properly-formatted JSON string
func AttributeValuesToJSON(attrs map[string]types.AttributeValue) ([]byte, error) {

	// Attempt to map the DynamoDB attribute value mapping to a map[string]interface{}
	// If this fails then return an error
	keys := make([]string, 0)
	mapping := toJSONInner(attrs, keys...)

	// Attempt to convert this mapping to JSON and return the result
	return json.Marshal(mapping)
}

// Helper function that converts a struct to JSON field-mapping
func toJSONInner(attrs map[string]types.AttributeValue, keys ...string) map[string]interface{} {
	jsonStr := make(map[string]interface{})
	for key, attr := range attrs {

		// Attempt to convert the field to a JSON mapping; if the value is nil then we'll ignore it and continue
		casted := toJSONField(attr, append(keys, key)...)
		if casted == nil {
			continue
		}

		// Set the field to its associated key in our mapping
		jsonStr[key] = casted
	}

	return jsonStr
}

// Helper function that converts a specific DynamoDB attribute value to its JSON value equivalent
func toJSONField(attr types.AttributeValue, keys ...string) interface{} {
	switch casted := attr.(type) {
	case *types.AttributeValueMemberB:
		return casted.Value
	case *types.AttributeValueMemberBOOL:
		return casted.Value
	case *types.AttributeValueMemberBS:
		return casted.Value
	case *types.AttributeValueMemberL:
		data := make([]interface{}, len(casted.Value))
		for i, item := range casted.Value {
			casted := toJSONField(item, keys...)
			data[i] = casted
		}

		return data
	case *types.AttributeValueMemberM:
		return toJSONInner(casted.Value, keys...)
	case *types.AttributeValueMemberN:
		return casted.Value
	case *types.AttributeValueMemberNS:
		return casted.Value
	case *types.AttributeValueMemberNULL:
		return nil
	case *types.AttributeValueMemberS:
		return casted.Value
	case *types.AttributeValueMemberSS:
		return casted.Value
	default:
		panic(fmt.Sprintf("Attribute at %s had unknown attribute type of %T",
			strings.Join(keys, "."), attr))
	}
}
