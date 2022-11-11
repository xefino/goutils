package extensions

import (
	"database/sql/driver"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
)

// ConvertibleDecimal wraps a Decimal object so that we can implement various marshal/unmarshal functionality
type ConvertibleDecimal struct {
	decimal.Decimal
}

// MarshalJSON converts a ConvertibleDecimal value to a JSON value
func (d ConvertibleDecimal) MarshalJSON() ([]byte, error) {
	return d.Decimal.MarshalJSON()
}

// MarshalCSV converts a ConvertibleDecimal value to CSV cell value
func (d ConvertibleDecimal) MarshalCSV() (string, error) {
	return d.Decimal.String(), nil
}

// MarshalYAML converts a ConvertibleDecimal value to a YAML node value
func (d ConvertibleDecimal) MarshalYAML() (interface{}, error) {
	return d.Decimal.String(), nil
}

// MarshalDynamoDBAttributeValue converts a ConvertibleDecimal value to a DynamoDB AttributeValue
func (d ConvertibleDecimal) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberN{Value: d.Decimal.String()}, nil
}

// Value converts a ConvertibleDecimal value to an SQL driver value
func (d ConvertibleDecimal) Value() (driver.Value, error) {
	return d.Decimal.String(), nil
}

// UnmarshalJSON attempts to convert a JSON value to a new ConvertibleDecimal value
func (d *ConvertibleDecimal) UnmarshalJSON(raw []byte) error {
	return d.Decimal.UnmarshalJSON(raw)
}

// UnmarshalCSV attempts to convert a CSV cell value to a new ConvertibleDecimal value
func (d *ConvertibleDecimal) UnmarshalCSV(raw string) error {
	parsed, err := decimal.NewFromString(raw)
	d.Decimal = parsed
	return err
}

// UnmarshalYAML attempts to convert a YAML node to a new ConvertibleDecimal value
func (d *ConvertibleDecimal) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("YAML node had an invalid kind (expected scalar value)")
	} else {
		parsed, err := decimal.NewFromString(value.Value)
		d.Decimal = parsed
		return err
	}
}

// UnmarshalDynamoDBAttributeValue attempts to convert a DynamoDB AttributeVAlue to a ConvertibleDecimal
// value. This function can handle []bytes, numerics, or strings. If the AttributeValue is NULL then
// the FillPolicy value will not be modified.
func (d *ConvertibleDecimal) UnmarshalDynamoDBAttributeValue(value types.AttributeValue) error {
	var parsed decimal.Decimal
	var err error
	switch casted := value.(type) {
	case *types.AttributeValueMemberB:
		parsed, err = decimal.NewFromString(string(casted.Value))
	case *types.AttributeValueMemberN:
		parsed, err = decimal.NewFromString(casted.Value)
	case *types.AttributeValueMemberNULL:
		return nil
	case *types.AttributeValueMemberS:
		parsed, err = decimal.NewFromString(casted.Value)
	default:
		return fmt.Errorf("Attribute value of %T could not be converted to a ConvertibleDecimal", value)
	}

	d.Decimal = parsed
	return err
}

// Scan attempts to convert an SQL driver value to a new ConvertibleDecimal value
func (d *ConvertibleDecimal) Scan(value interface{}) error {
	parsed, err := decimal.NewFromString(value.(string))
	d.Decimal = parsed
	return err
}
