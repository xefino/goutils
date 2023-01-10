package random

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("String Tests", func() {

	// Test that the RandomNRunes function works as expected
	It("RandomNRunes - Works", func() {
		rand := RandomNRunes(50, Uppercase)
		Expect(rand).Should(HaveLen(50))
		for _, c := range rand {
			Expect(Uppercase).Should(ContainElement(c))
		}
	})
})
