package time

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	xstr "github.com/xefino/goutils/strings"
	"gopkg.in/yaml.v3"
)

// Convertible represents a wrapper for a time.Time object, allowing it to be converted to and from
// various formats, including SQL, DynamoDB, JSON, YAML and CSV. The exact format the time should take
// will be determined by the value provided to the Layout field and will operate according to Go's
// documentation for time.Time.
type Convertible struct {
	yaml.Marshaler
	time.Time
	Layout string
}

// WithLayout creates a new Convertible time from an inner time and a layout.
func WithLayout(inner time.Time, layout string) Convertible {
	return Convertible{
		Time:   inner,
		Layout: layout,
	}
}

// MarshalJSON converts the time contained in the Convertible to JSON, using the Layout field associated
// with this Convertible to format the time (or RFC3339 if no layout was provided).
func (c Convertible) MarshalJSON() ([]byte, error) {
	return []byte("\"" + c.marshalInner() + "\""), nil
}

// MarshalCSV converts the time contained in the Convertible to a CSV column, using the Layout field
// associated with this Convertible to format the time (or RFC3339 if no layout was provided).
func (c Convertible) MarshalCSV() (string, error) {
	return c.marshalInner(), nil
}

// MarshalYAML converts the time contained in the Convertible to a YAML node, using the Layout field
// associated with this Convertible to format the time (or RFC3339 if no layout was provided).
func (c Convertible) MarshalYAML() (interface{}, error) {
	return c.marshalInner(), nil
}

// MarshalDynamoDBAttributeValue converts the time contained in the Convertible to a DynamoDB AttributeValue
// object, using the Layout field associated with this Convertible to format the time (or RFC3339 if
// no layout was provided). The Convertible converts to a DynamoDB string attribute.
func (c Convertible) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: c.marshalInner()}, nil
}

// Value converts the time contained in the Convertible to an SQL driver value, using the Layout field
// associated with this Convertible to format the time (or RFC3339 if no layout was provided).
func (c Convertible) Value() (driver.Value, error) {
	return driver.Value(c.marshalInner()), nil
}

// UnmarshalJSON converts a JSON string to a time.Time and sets the inner Time on this Convertible to
// that value, using the Layout field to parse the raw data (or RFC3339 if no layout was provided).
func (c *Convertible) UnmarshalJSON(raw []byte) error {
	return c.unmarshalInner(strings.Trim(string(raw), "\""))
}

// UnmarshalCSV converts a CSV column to a time.Time and sets the inner Time on this Convertible to
// that value, using the Layout field to parse the raw data (or RFC3339 if no layout was provided).
func (c *Convertible) UnmarshalCSV(raw string) error {
	return c.unmarshalInner(raw)
}

// UnmarshalYAML converts a YAML node to a time.Time and sets the inner Time on this Convertible to
// that value, using the Layout field to parse the raw data (or RFC3339 if no layout was provided).
func (c *Convertible) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("YAML node had an invalid kind (expected scalar value)")
	} else {
		return c.unmarshalInner(value.Value)
	}
}

// UnmarshalDynamoDBAttributeValue converts a DynamoDB AttributeValue object to a time.Time and sets
// the inner Time on this Convertible to that value, using the Layout field to parse the raw data (or
// RFC3339 if no layout was provided). String or byte array AttributeValues can be converted and NULL
// AttributeValues will result in no change to the object. Any other type of AttributeValue will result
// in an error.
func (c *Convertible) UnmarshalDynamoDBAttributeValue(value types.AttributeValue) error {
	var asStr string
	switch casted := value.(type) {
	case *types.AttributeValueMemberB:
		asStr = string(casted.Value)
	case *types.AttributeValueMemberNULL:
		return nil
	case *types.AttributeValueMemberS:
		asStr = casted.Value
	default:
		return fmt.Errorf("Attribute value of %T could not be converted to a time.Time", value)
	}

	return c.unmarshalInner(asStr)
}

// Scan converts an SQL driver value to a time.Time and sets the inner Time on this Convertible to that
// value, using the Layout field to parse the raw data (or RFC3339 if no layout was provided). This
// function expects the driver value to be convertible to a string.
func (c *Convertible) Scan(value interface{}) error {
	return c.unmarshalInner(value.(string))
}

// Helper function that retrieves the layout that should be used for conversion operations. This function
// will return the RFC3339 layout if the layout associated with the Convertible is empty
func (c Convertible) getLayout() string {
	if xstr.IsEmpty(c.Layout) {
		return time.RFC3339
	} else {
		return c.Layout
	}
}

// Helper function that attempts to marshal a Convertible object to a string, using the layout associated
// with the object, or RFC3339 if the layout fields was not provided
func (c *Convertible) marshalInner() string {
	return c.Time.Format(c.getLayout())
}

// Helper function that attempts to unmarshal a string to a Convertible object, using the layout associated
// with the object, or RFC3339 if the layout field was not provided
func (c *Convertible) unmarshalInner(raw string) error {
	parsed, err := time.Parse(c.getLayout(), raw)
	if err != nil {
		return fmt.Errorf("Failed to parse %q to time, error: %v", raw, err)
	}

	c.Time = parsed
	return nil
}
