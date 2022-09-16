package reflection

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	wstr "github.com/xefino/goutils/strings"
)

// Cache containing concurrency-safe field info for each type examined
var fieldsCache = new(sync.Map)

type TypeInfo struct {
	Name   string
	Path   string
	Fields []*FieldInfo
}

// FieldInfo contains information concerning a single field on a struct
type FieldInfo struct {
	reflect.StructField
	Tags map[string]*TagInfo
}

// TagInfo contains information concerning a tag on a field
type TagInfo struct {
	Raw       string
	Name      string
	Modifiers []string
}

// GetTypeInfo retrieves information on the type submitted to the function. A list of FieldInfo
// objects will be returned, where each contains the SructField for the associated field along with a
// mapping of tag names to their values. This list will be ordered in the same way as the fields.
// Additionally, the field into will be cached so that subsequent accesses do not incur additional costs
func GetTypeInfo[T any]() *TypeInfo {

	// First, get the reflected type, its package path and name
	tType := reflect.TypeOf(*new(T))
	pkg := tType.PkgPath()
	name := tType.Name()

	// Next, check if we've already looked at this type; if we have then return its associated field info
	typeName := fmt.Sprintf("%s.%s", tType.PkgPath(), tType.Name())
	if raw, ok := fieldsCache.Load(typeName); ok {
		return raw.(*TypeInfo)
	}

	// Now, iterate over all the fields on the type and extract the info for each
	numFields := tType.NumField()
	fields := make([]*FieldInfo, numFields)
	for i := 0; i < numFields; i++ {

		// First, create a new FieldInfo from the field and a new map for tags info
		fields[i] = &FieldInfo{
			StructField: tType.Field(i),
			Tags:        make(map[string]*TagInfo),
		}

		// Check if the tag is empty; if it is then there's nothing more to do here so continue
		if wstr.IsEmpty(fields[i].Tag) {
			continue
		}

		// Next, split the tag into groups by the space character (this is the typical value)
		tagGroups := strings.Split(strings.TrimSpace(string(fields[i].Tag)), " ")

		// Finally, iterate over all the tag groups we extracted and convert each to a TagInfo
		for _, group := range tagGroups {

			// First, split the tag group by a colon to get its key and value
			colonSpplit := strings.Split(group, ":")

			// Next, strip the quotes from the value and split it by a comma to extract modifiers
			raw := strings.Trim(colonSpplit[1], "\"")
			commaSplit := strings.Split(raw, ",")

			// Now, check if we have any modifiers and, if we do, extract them
			var modifiers []string
			if len(commaSplit) > 1 {
				modifiers = commaSplit[1:]
			}

			// Finally, create a new TagInfo with the raw tag, its name and any modifiers and
			// associate it with the tag name
			fields[i].Tags[colonSpplit[0]] = &TagInfo{
				Raw:       raw,
				Name:      commaSplit[0],
				Modifiers: modifiers,
			}
		}
	}

	// Finally, inject the name, package path and fields into a new type info object
	info := &TypeInfo{
		Name:   name,
		Path:   pkg,
		Fields: fields,
	}

	// Store the info object in the cache with the type name and return it
	fieldsCache.Store(typeName, info)
	return info
}
