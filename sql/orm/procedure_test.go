package orm

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("Procedure Call Tests", func() {

	// Test that the Source function works as expected
	It("Source - Works", func() {
		call := NewProcedureCall("sp_get_ohlc", 1, 5, 20, "APPL")
		gomega.Expect(call.Source()).Should(gomega.Equal("sp_get_ohlc"))
	})

	// Test that the String function works as expected
	It("String - Works", func() {
		call := NewProcedureCall("sp_get_ohlc", 1, 5, 20, "APPL")
		gomega.Expect(call.String()).Should(gomega.Equal("CALL sp_get_ohlc(?, ?, ?, ?)"))
	})

	// Test that the Arguments function works as expected
	It("Arguments - Works", func() {
		call := NewProcedureCall("sp_get_ohlc", 1, 5, 20, "APPL")
		args := call.Arguments()
		gomega.Expect(args).Should(gomega.HaveLen(4))
		gomega.Expect(args[0]).Should(gomega.Equal(1))
		gomega.Expect(args[1]).Should(gomega.Equal(5))
		gomega.Expect(args[2]).Should(gomega.Equal(20))
		gomega.Expect(args[3]).Should(gomega.Equal("APPL"))
	})
})
