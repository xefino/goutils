package utils

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Helper type that we'll use to test error generation from
// within the context of a class
type errorGenerator struct {
	value string
}

// Generate a test error with an inner error or without one
func (e *errorGenerator) generate(hasInner bool) *GError {

	// If we want to test with an inner error then set it here
	var err error
	if hasInner {
		err = fmt.Errorf("derped")
	}

	// Create the test error and return it
	return NewError("test", err, "generated from class, ID: %d, Value: %s", 42, e.value)
}

// Helper function that we'll use to test error generation from
// within the context of a global function
func generate(hasInner bool) *GError {

	// If we want to test with an inner error then set it here
	var err error
	if hasInner {
		err = fmt.Errorf("derped")
	}

	// Create the test error and return it
	return NewError("test", err, "generated from func, ID: %v, Value: herp", 42)
}

var _ = Describe("Errors Tests", func() {

	// Tests the conditions determining how an error is generated from various operating conditions
	DescribeTable("NewError - Conditions",
		func(generator func(bool) *GError, class string, line int,
			hasInner bool, inner string, innerMessage string, message string) {

			// First, generate the error from the generator
			err := generator(hasInner)

			// Next, verify the base fields on the error
			Expect(err.Class).Should(Equal(class))
			Expect(err.Environment).Should(Equal("test"))
			Expect(err.File).Should(Equal("/goutils/utils/error_test.go"))
			Expect(err.Function).Should(Equal("generate"))
			Expect(err.GeneratedAt).ShouldNot(BeNil())
			Expect(err.LineNumber).Should(Equal(line))
			Expect(err.Message).Should(Equal(innerMessage))
			Expect(err.Package).Should(Equal("utils"))
			Expect(err.Error()).Should(HaveSuffix(message))

			// Finally, verify the inner error on the error
			if hasInner {
				Expect(err.Inner).Should(HaveOccurred())
				Expect(err.Inner.Error()).Should(Equal(inner))
			} else {
				Expect(err.Inner).ShouldNot(HaveOccurred())
			}
		},
		EntryDescription("Error generator %v -> Class %s, Line %d, Has Error? %t, Error: %v, Message %s"),
		Entry("Generated from class, No inner error - Generated", (&errorGenerator{value: "herp"}).generate,
			"errorGenerator", 26, false, "", "generated from class, ID: 42, Value: herp",
			"[test] utils.errorGenerator.generate (/goutils/utils/error_test.go 26): generated from class, "+
				"ID: 42, Value: herp."),
		Entry("Generated from class, Inner error - Generated", (&errorGenerator{value: "herp"}).generate,
			"errorGenerator", 26, true, "derped", "generated from class, ID: 42, Value: herp",
			"[test] utils.errorGenerator.generate (/goutils/utils/error_test.go 26): generated from class, "+
				"ID: 42, Value: herp, Inner: derped."),
		Entry("Generated from global scope, No inner error - Generated", generate, "", 40, false, "",
			"generated from func, ID: 42, Value: herp", "[test] utils.generate (/goutils/utils/error_test.go 40): "+
				"generated from func, ID: 42, Value: herp."),
		Entry("Generated from global scope, Inner error - Generated", generate, "", 40, true, "derped",
			"generated from func, ID: 42, Value: herp", "[test] utils.generate (/goutils/utils/error_test.go 40): "+
				"generated from func, ID: 42, Value: herp, Inner: derped."))

	// Test that calling FromErrors with an empty list will return no error
	It("FromErrors - No errors - Nil returned", func() {
		err := FromErrors()
		Expect(err).ShouldNot(HaveOccurred())
	})

	// Tests that the FromErrors function will create an aggregate error from the
	// list of errors that were provided to the function
	It("FromErrors - At least one error - Error returned", func() {

		// Create the aggregate error from the errors provided
		err := FromErrors(fmt.Errorf("Error 1"), fmt.Errorf("Error 2"))

		// Verify that the error was created
		Expect(err).Should(HaveLen(2))
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("Multiple errors occurred: \n\tError 1\n\tError 2"))
	})

	// Test that, if error passed to the As function cannot be casted to the type
	// indicated by the parameter, then the function will panic
	It("As - Conversion not possible - Panic", func() {

		// Create a function to return an error as the error interface and then call it
		err := func() error {
			return &GError{
				Environment: "test",
				Package:     "pack",
				Class:       "class",
				Function:    "func",
				File:        "/test/tmp/file.go",
				LineNumber:  42,
				Message:     "derp",
			}
		}()

		// Create a test error that is not convertible from a BackendError
		type derpError struct {
			error
		}

		// Attempt to cast the value we created to the error type we created
		// This should panic
		Expect(func() {
			_ = As[*derpError](err)
		}).Should(Panic())
	})

	// Test that the As function works as expected when the cast can be performed
	It("As - Conversion possible - Works", func() {

		// Create a function to return an error as the error interface and then call it
		err := func() error {
			return &GError{
				Environment: "test",
				Package:     "pack",
				Class:       "class",
				Function:    "func",
				File:        "/test/tmp/file.go",
				LineNumber:  42,
				Message:     "derp",
			}
		}()

		// Attempt to cast the error to its specific type
		converted := As[*GError](err)

		// Verify the results
		Expect(err).ShouldNot(BeNil())
		Expect(converted).ShouldNot(BeNil())
		Expect(converted.Class).Should(Equal("class"))
		Expect(converted.Environment).Should(Equal("test"))
		Expect(converted.File).Should(Equal("/test/tmp/file.go"))
		Expect(converted.Function).Should(Equal("func"))
		Expect(converted.GeneratedAt).Should(BeZero())
		Expect(converted.Inner).ShouldNot(HaveOccurred())
		Expect(converted.LineNumber).Should(Equal(42))
		Expect(converted.Message).Should(Equal("derp"))
		Expect(converted.Package).Should(Equal("pack"))
	})
})
