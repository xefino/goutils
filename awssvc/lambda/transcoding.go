package lambda

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// AttributesToJSON attempts to convert a mapping of DynamoDB attribute values to a properly-formatted JSON string
func AttributesToJSON(attrs map[string]events.DynamoDBAttributeValue) ([]byte, error) {

	// Attempt to map the DynamoDB attribute value mapping to a map[string]interface{}
	// If this fails then return an error
	keys := make([]string, 0)
	mapping := toJSONInner(attrs, keys...)

	// Attempt to convert this mapping to JSON and return the result
	return json.Marshal(mapping)
}

// Helper function that converts a struct to JSON field-mapping
func toJSONInner(attrs map[string]events.DynamoDBAttributeValue, keys ...string) map[string]interface{} {
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
func toJSONField(attr events.DynamoDBAttributeValue, keys ...string) interface{} {
	attrType := attr.DataType()
	switch attrType {
	case events.DataTypeBinary:
		return attr.Binary()
	case events.DataTypeBinarySet:
		return attr.BinarySet()
	case events.DataTypeBoolean:
		return attr.Boolean()
	case events.DataTypeList:

		// Get the list of items from the attribute value
		list := attr.List()

		// Attempt to convert each item in the list to a JSON mapping
		data := make([]interface{}, len(list))
		for i, item := range list {
			casted := toJSONField(item, keys...)
			data[i] = casted
		}

		// Return the list we created
		return data
	case events.DataTypeMap:
		return toJSONInner(attr.Map(), keys...)
	case events.DataTypeNull:
		return nil
	case events.DataTypeNumber:
		return attr.Number()
	case events.DataTypeNumberSet:
		return attr.NumberSet()
	case events.DataTypeString:
		return attr.String()
	case events.DataTypeStringSet:
		return attr.StringSet()
	default:
		panic(fmt.Sprintf("Attribute at %s had unknown attribute type of %d",
			strings.Join(keys, "."), attrType))
	}
}
