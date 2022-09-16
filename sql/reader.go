package sql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/xefino/goutils/collections"
	"github.com/xefino/goutils/reflection"
)

// Get the type associated with the sql.Scanner interface so we can check whether or not the type
// associated with a given field implements it
var scannerType = func() reflect.Type {
	var scanner sql.Scanner
	return reflect.TypeOf(&scanner).Elem()
}()

// Cache to use for field info so we don't have to reuse reflection as much
var fieldInfoCache = new(sync.Map)

// Helper type we'll use to hold additional information about each field on the object
type fieldInfo struct {
	Exported          bool
	ImplementsScanner bool
	Tag               string
}

// ReadRows extracts all the data from a single SQL result set into a slice of the specified type.
// This function attempts to map each column in the result set to an associated field on the item,
// based on the value of an "sql" tag, a "json" tag or the field name in that order. The mapping is
// case-insensitive. If any column can't be mapped then an error will be returned.
func ReadRows[T any](rows *sql.Rows) ([]*T, error) {
	defer rows.Close()

	// First, get the column names from the result set and map them to their indices
	columnNames, _ := rows.Columns()
	columnMapping := collections.IndexWithFunction(columnNames, func(name string) string {
		return strings.ToLower(name)
	})

	// Next, get the type info and attempt to map a field to each column in the result set; if there
	// were any fields that could not be mapped then return an error here
	typeInfo := reflection.GetTypeInfo[T]()
	fieldColumnMapping, implementsScanner, err := mapFields[T](typeInfo, columnMapping)
	if err != nil {
		return nil, err
	}

	// Now, create the scanner function from the field-column mapping that we'll use to read each row
	scanner := func(result *T) error {

		// First, get the value of the object we'll be writing to
		tValue := reflect.ValueOf(result).Elem()

		// Next, iterate over all the fields on the type and create a value the column can be read into
		// as well as an assigner function to transfer the value to the field if the value is nullable
		values := make([]interface{}, len(columnNames))
		assigners := make([]func() error, 0)
		for i, field := range typeInfo.Fields {

			// First, get the index of the column associated with the field. This value may not
			// exist because the field wasn't exported or because there was no column that could be
			// associated with the field
			index, ok := fieldColumnMapping[i]
			if !ok {
				continue
			}

			// Next, attempt to generate a value for the column, an assigner to transfer that value to
			// a field on the object and whether or not this assigner exists. If it does exist we'll add
			// it to our list of such assigners
			value, assigner, ok := generateValuer(tValue.Field(i), field, implementsScanner[i])
			if ok {
				assigners = append(assigners, assigner)
			}

			// Finally, add our value to our list of values at the index associated with the SQL column
			values[index] = value
		}

		// Finally, attempt to scan the row into the values and then attempt to assign them to the
		// field. If the scan fails or any of the assigners fail then return an error
		if err := rows.Scan(values...); err == nil {
			for _, assigner := range assigners {
				if err := assigner(); err != nil {
					return err
				}
			}

			return nil
		} else {
			return err
		}
	}

	// Finally, iterate over all the rows in the result set and attempt to read each into an item
	// of the type sent into the function
	data := make([]*T, 0)
	for rows.Next() {

		// Attempt to scan the row into a new result object; if this fails then return an error
		var result T
		if err := scanner(&result); err != nil {
			return nil, fmt.Errorf("Failed to read row into object of type %T, error: %v", result, err)
		}

		// Add the result to our list of data
		data = append(data, &result)
	}

	// If the rows failed for any reason at the end then return an error
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Row could not be read, error: %v", err)
	}

	return data, nil
}

// Helper function that maps the fields to the SQL columns returned from the query
func mapFields[T any](typeInfo *reflection.TypeInfo, columnMapping map[string]int) (map[int]int, []bool, error) {
	fieldColumnMapping := make(map[int]int)

	// Iterate over the fields on the type and attempt to map the field to its associated column
	// in the result set
	implementsScanner := make([]bool, len(typeInfo.Fields))
	for i, field := range typeInfo.Fields {

		// First, generate additional field info we need; if the field is not exported then there's
		// no additional work to do here because we don't unmarshal to unexported fields
		info := generateFieldInfo[T](typeInfo.Name, typeInfo.Path, field)
		if !info.Exported {
			continue
		}

		// Next, save whether or not the field implements the sql.Scanner interface
		implementsScanner[i] = info.ImplementsScanner

		// Finally, use the tag to map the field to its associated column in the result set. If we
		// do manage to map the field then remove the field from our column mapping
		if index, ok := columnMapping[info.Tag]; ok {
			fieldColumnMapping[i] = index
			delete(columnMapping, info.Tag)
		}
	}

	// If there are any unmapped SQL columns then generate and return an error because we have extra
	// data that cannot be read from the SQL response
	if len(columnMapping) > 0 {
		return nil, nil, fmt.Errorf("Columns %v were not mapped to any field on %T",
			collections.Keys(columnMapping), *new(T))
	}

	// Return the field-column mapping and which columns implement sql.Scanner
	return fieldColumnMapping, implementsScanner, nil
}

// Helper function that generates additional information for each field on the object, such as whether
// the field is exported, whether it implements the sql.Scanner interface and the tag to associate with
// the column on the SQL result set
func generateFieldInfo[T any](name string, pkg string, field *reflection.FieldInfo) *fieldInfo {

	// First, check the cache for the field information. If we already have it saved then return it
	fieldName := fmt.Sprintf("%s.%s.%s", name, pkg, field.Name)
	raw, ok := fieldInfoCache.Load(fieldName)
	if ok {
		return raw.(*fieldInfo)
	}

	// Next, check if the field is exported; if it isn't then we'll return an empty field info as
	// there's no additional work necessary; we don't import into unexported fields
	info := new(fieldInfo)
	if info.Exported = field.IsExported(); !info.Exported {
		return info
	}

	// Now, check if the type of the field, or a pointer to it, implements the sql.Scanner interface
	info.ImplementsScanner = field.Type.Implements(scannerType) ||
		reflect.PointerTo(field.Type).Implements(scannerType)

	// Finally, attempt to get the tag to associate with the SQL column name. If the field has an
	// SQL tag present, then we'll use that. Otherwise, we'll use the JSON tag if present or the
	// field's name if it isn't
	var tag string
	if sqlTag, ok := field.Tags["sql"]; ok {
		tag = sqlTag.Name
	} else if jsonTag, ok := field.Tags["json"]; ok {
		tag = jsonTag.Name
	} else {
		tag = field.Name
	}

	// Ensure that casing is ignored by converting the tag to lowercase (this will match with the SQL column,
	// which is also converted to lowercase)
	info.Tag = strings.ToLower(tag)

	// Store the info in the cache so it can be retrieved later and return it
	fieldInfoCache.Store(fieldName, info)
	return info
}

// Helper function that generates the value to which the SQL data will be written as well as an
// assigner function that will be used to transfer the value to the field on the object if the value
// isn't null or ignore it if it is null
func generateValuer(vField reflect.Value, field *reflection.FieldInfo, implementsScanner bool) (interface{}, func() error, bool) {
	var value interface{}

	if !implementsScanner {
		var checker func(interface{}) (interface{}, bool)
		hasAssigner := true

		switch kind := vField.Kind(); kind {
		case reflect.Bool:
			value = new(sql.NullBool)
			checker = checkBool
		case reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16, reflect.Int,
			reflect.Uint, reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64:
			value = new(sql.NullInt64)
			checker = checkInt64
		case reflect.Float32, reflect.Float64:
			value = new(sql.NullFloat64)
			checker = checkFloat64
		case reflect.String:
			value = new(sql.NullString)
			checker = checkString
		default:
			hasAssigner = false
		}

		if hasAssigner {
			return value, func() error {
				if inner, valid := checker(value); valid {
					vValue := reflect.ValueOf(inner)
					tValue := vValue.Type()
					if tValue.AssignableTo(field.Type) {
						vField.Set(vValue)
					} else if vValue.CanConvert(field.Type) {
						vField.Set(vValue.Convert(field.Type))
					} else {
						return fmt.Errorf("Field %s cannot be assigned a value of type %s",
							field.Name, tValue.Name())
					}
				}

				return nil
			}, true
		}
	}

	value = vField.Addr().Interface()
	return value, nil, false
}

// Helper function that extracts the inner Boolean value and whether or not the SQL value is NULL
// from a nullable Boolean SQL value
func checkBool(raw interface{}) (interface{}, bool) {
	value := raw.(*sql.NullBool)
	return value.Bool, value.Valid
}

// Helper function that extracts the inner 64-bit integer value and whether or not the SQL value
// is NULL from a nullable integer SQL value
func checkInt64(raw interface{}) (interface{}, bool) {
	value := raw.(*sql.NullInt64)
	return value.Int64, value.Valid
}

// Helper function that extracts the inner 64-bit floating point value and whether or not the SQL
// value is NULL from a nullable floating-point SQL value
func checkFloat64(raw interface{}) (interface{}, bool) {
	value := raw.(*sql.NullFloat64)
	return value.Float64, value.Valid
}

// Helper function that extracts the inner string value and whether or not the SQL value is NULL
// from a nullable string SQL value
func checkString(raw interface{}) (interface{}, bool) {
	value := raw.(*sql.NullString)
	return value.String, value.Valid
}
