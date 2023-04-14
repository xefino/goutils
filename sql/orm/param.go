package orm

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/xefino/goutils/math"
)

// Parameter describes the functionality that should exist for parameter terms
type Parameter interface {
	ModifyQuery(*Query) string
}

// Constant contains the data used to serialize a constant parameter in an SQL query
type Constant struct {
	value     driver.Value
	arguments []any
}

// NewConstant creates a new constant parameter from a driver.Value and a list of arguments used to
// convert the value to a string. For strings, no additional arguments are accepted. For a time.Time,
// the optional arguments conform to the Format function. For integers and unsigned integeters, the
// optional arguments will conform to those submitted to strconv.FormatInt or strconv.FormatUint respectively.
// For floating-point values, the optional arguments conform to strconv.FormatFloat except for the size
// parameter, which will be decided based on the value submitted to the function. Boolean values will
// be converted to TRUE or FALSE depending on the value of the parameter. No other types will be accepted.
func NewConstant(value driver.Value, args ...any) *Constant {
	return &Constant{
		value:     value,
		arguments: args,
	}
}

// ModifyQuery modifies the query, returning the string value of the parameter.
func (c *Constant) ModifyQuery(query *Query) string {
	switch casted := c.value.(type) {
	case string, []byte:
		return fmt.Sprintf("'%s'", casted)
	case time.Time:

		// Check if we have at least one argument. If we do then we'll attempt to format the date according
		// to the first argument we received, if it is a string.
		if len(c.arguments) >= 1 {
			layout, ok := c.arguments[0].(string)
			if ok {
				return "'" + casted.Format(layout) + "'"
			}
		}

		// Otherwise, we'll send the date to fmt.Sprintf and let Go format it for us
		return fmt.Sprintf("'%v'", casted)
	case int:
		return math.FormatInt(casted, c.arguments...)
	case int8:
		return math.FormatInt(casted, c.arguments...)
	case int16:
		return math.FormatInt(casted, c.arguments...)
	case int32:
		return math.FormatInt(casted, c.arguments...)
	case int64:
		return math.FormatInt(casted, c.arguments...)
	case uint:
		return math.FormatUint(casted, c.arguments...)
	case uint8:
		return math.FormatUint(casted, c.arguments...)
	case uint16:
		return math.FormatUint(casted, c.arguments...)
	case uint32:
		return math.FormatUint(casted, c.arguments...)
	case uint64:
		return math.FormatUint(casted, c.arguments...)
	case float32:
		return math.FormatFloat(casted, c.arguments...)
	case float64:
		return math.FormatFloat(casted, c.arguments...)
	case bool:
		if casted {
			return "TRUE"
		} else {
			return "FALSE"
		}
	default:
		panic(fmt.Sprintf("Argument of type %T could not be parsed", c.value))
	}
}

// Argument contains the data used to inject a variable parameter into an SQL query
type Argument struct {
	value any
}

// NewArgument creates a new SQL argument from the variable that should be injected.
func NewArgument(value any) *Argument {
	return &Argument{
		value: value,
	}
}

// ModifyQuery modifies the query, returning the string value of the parameter.
func (a *Argument) ModifyQuery(query *Query) string {
	query.arguments = append(query.arguments, a.value)
	return "?"
}

// Helper function that creates a parameter from a value and a flag showing whether or not the value
// should be injected as an argument, or inserted as a constant value
func param(value any, constant bool) Parameter {
	if constant {
		return NewConstant(value)
	} else {
		return NewArgument(value)
	}
}
