package reflection

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Create a new test runner we'll use to test all the
// modules in the reflection package
func TestReflection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reflection Suite")
}

var _ = Describe("Fields Tests", func() {

	// Tests that the GetFieldInfo function works as expected when the value is not cached
	It("GetTypeInfo - Not cached - Works", func() {

		// First, get the field info for the type
		info := GetTypeInfo[testStruct]()

		// Next, verify the FieldInfo that were returned
		verifyType(info, "testStruct", "reflection",
			verifyField("StringType", "string",
				tagVerifier("json", "string_type", "string_type"),
				tagVerifier("sql", "STR_TYPE", "STR_TYPE")),
			verifyField("IntType", "int",
				tagVerifier("json", "int_type,omitempty", "int_type", "omitempty")),
			verifyField("FloatType", "float64",
				tagVerifier("sql", "float_type,omitempty", "float_type", "omitempty")),
			verifyField("UntaggedType", "string"))

		// Now, attempt to verify the data in the cache
		total := 0
		fieldsCache.Range(func(rawKey any, rawValue any) bool {
			total += 1

			// Check that the key corresponds to our test struct
			Expect(rawKey).Should(HaveSuffix("reflection.testStruct"))

			// Verify that the data has the value we expect
			info := rawValue.(*TypeInfo)
			verifyType(info, "testStruct", "reflection",
				verifyField("StringType", "string",
					tagVerifier("json", "string_type", "string_type"),
					tagVerifier("sql", "STR_TYPE", "STR_TYPE")),
				verifyField("IntType", "int",
					tagVerifier("json", "int_type,omitempty", "int_type", "omitempty")),
				verifyField("FloatType", "float64",
					tagVerifier("sql", "float_type,omitempty", "float_type", "omitempty")),
				verifyField("UntaggedType", "string"))
			return true
		})

		// Finally, verify the number of entries in the cache
		Expect(total).Should(Equal(1))
	})

	// Tests that the GetFieldInfo function returns the data in the cache if the struct has
	// already been mapped
	It("GetFieldInfo - Cached - Works", func() {

		// First, create fake field info values and add them to the cache
		typ := reflect.TypeOf(*new(testStruct))
		fieldsCache.Store(fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name()), &TypeInfo{
			Name: "fake",
			Path: "fake/path",
			Fields: []*FieldInfo{
				{
					Tags: map[string]*TagInfo{
						"fake": {
							Raw:       "fake_raw",
							Name:      "fake_name",
							Modifiers: []string{"fake_modifiers"},
						},
					},
				},
			},
		})

		// Next, get the field info for the type
		info := GetTypeInfo[testStruct]()

		// Finally, verify that the returned value is our fake value
		Expect(info.Name).Should(Equal("fake"))
		Expect(info.Path).Should(Equal("fake/path"))
		Expect(info.Fields).Should(HaveLen(1))
		Expect(info.Fields[0].Tags).Should(HaveLen(1))
		tagVerifier("fake", "fake_raw", "fake_name", "fake_modifiers")(info.Fields[0].Tags)
	})
})

// Fake type we'll use for testing reflection library
type testStruct struct {
	StringType   string  `json:"string_type" sql:"STR_TYPE"`
	IntType      int     `json:"int_type,omitempty"`
	FloatType    float64 `sql:"float_type,omitempty"`
	UntaggedType string
}

// Helper function that verifies the fields on a TypeInfo object
func verifyType(info *TypeInfo, name string, pathSuffix string, fieldVerifiers ...func(*FieldInfo)) {
	Expect(info.Name).Should(Equal(name))
	Expect(info.Path).Should(HaveSuffix(pathSuffix))
	Expect(info.Fields).Should(HaveLen(len(fieldVerifiers)))
	for i, field := range info.Fields {
		fieldVerifiers[i](field)
	}
}

// Helper function that verifies the fields on a FieldInfo object
func verifyField(fieldName string, typeName string, tagVerifiers ...func(map[string]*TagInfo)) func(*FieldInfo) {
	return func(field *FieldInfo) {
		Expect(field.Name).Should(Equal(fieldName))
		Expect(field.Type.Name()).Should(Equal(typeName))
		Expect(field.Tags).Should(HaveLen(len(tagVerifiers)))
		for _, verifier := range tagVerifiers {
			verifier(field.Tags)
		}
	}
}

// Helper function that verifies the data associated with a particular tag key and value
func tagVerifier(key string, rawValue string, name string,
	modifiers ...string) func(map[string]*TagInfo) {
	return func(tags map[string]*TagInfo) {
		info := tags[key]
		Expect(tags).Should(HaveKey(key))
		Expect(info.Name).Should(Equal(name))
		Expect(info.Raw).Should(Equal(rawValue))
		Expect(info.Modifiers).Should(Equal(modifiers))
	}
}
