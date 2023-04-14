package orm

import (
	"database/sql/driver"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("Param Tests", func() {

	// Tests that, if the object Constant was created with has a type we don't recognize, then calling
	// the ModifyQuery function will result in a panic
	It("Constant - ModifyQuery - Type invalid - Panic", func() {

		// Create a new constant with an invalid value
		c := NewConstant(testType{Key: "key", Value: "value"})

		// Attempt to call the ModifyQuery function with our invalid value; this should panic
		gomega.Expect(func() {
			_ = c.ModifyQuery(NewQuery())
		}).Should(gomega.PanicWith("Argument of type orm.testType could not be parsed"))
	})

	// Tests that the ModifyQuery function works as expected when the type of the argument is recognized
	DescribeTable("Constant - ModifyQuery - Types",
		func(value driver.Value, expected string, args ...any) {
			c := NewConstant(value, args...)
			gomega.Expect(c.ModifyQuery(NewQuery())).Should(gomega.Equal(expected))
		},
		Entry("String - Works", "derp", "'derp'"),
		Entry("Bytes - Works", []byte("derp"), "'derp'"),
		Entry("Time, No Layout - Works", time.Date(2022, time.January, 11, 12, 34, 50, 900, time.UTC),
			"'2022-01-11 12:34:50.0000009 +0000 UTC'"),
		Entry("Time, Layout - Works", time.Date(2022, time.January, 11, 12, 34, 50, 900, time.UTC),
			"'2022-01-11T12:34:50'", "2006-01-02T15:04:05"),
		Entry("Int - Works", int(42), "42"),
		Entry("Int8 - Works", int8(120), "120"),
		Entry("Int16 - Works", int16(32000), "32000"),
		Entry("Int32 - Works", int32(400000000), "400000000"),
		Entry("Int64 - Works", int64(1000000000000000000), "1000000000000000000"),
		Entry("Uint - Works", uint(42), "42"),
		Entry("Uint8 - Works", uint8(250), "250"),
		Entry("Uint16 - Works", uint16(60000), "60000"),
		Entry("Uint32 - Works", uint32(4000000000), "4000000000"),
		Entry("Uint64 - Works", uint64(10000000000000000000), "10000000000000000000"),
		Entry("Float32 - Works", float32(5.99), "5.99"),
		Entry("Float64 - Works", float64(6.0001), "6.0001"),
		Entry("Bool, True - Works", true, "TRUE"),
		Entry("Bool, False - Works", false, "FALSE"))

	// Tests that the Argument ModifyQuery function works as expected
	It("Argument - ModifyQuery - Works", func() {

		// First, create a variable and inject it into a new argument
		x := 10
		arg := NewArgument(x)

		// Next, create a query and modify the query with that argument
		query := NewQuery()
		param := arg.ModifyQuery(query)

		// Finally,
		gomega.Expect(param).Should(gomega.Equal("?"))
		gomega.Expect(query.arguments).Should(gomega.HaveLen(1))
		gomega.Expect(query.arguments[0]).Should(gomega.Equal(10))
	})
})
