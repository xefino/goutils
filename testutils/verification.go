package testutils

import (
	. "github.com/onsi/gomega"
	"github.com/xefino/goutils/utils"
)

// ItemIsNil is a function that verifies that the data sent to it is nil
func ItemIsNil[T any](item T) {
	Expect(item).Should(BeNil())
}

// VerifyList verifies that the number of items and the fields on each item
func VerifyList[T any](actual []T, verifiers ...func(T)) {
	Expect(actual).Should(HaveLen(len(verifiers)))
	for i, item := range actual {
		verifiers[i](item)
	}
}

// NoInnerError verifies that an error did not occur
func NoInnerError() func(error) {
	return func(err error) {
		Expect(err).ShouldNot(HaveOccurred())
	}
}

// InnerErrorVerifier verifies an error message
func InnerErrorVerifier(message string) func(error) {
	return func(err error) {
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal(message))
	}
}

// InnerErrorPrefixSuffixVerifier verifies that an error message has a given
// prefix and suffix
func InnerErrorPrefixSuffixVerifier(prefix string, suffix string) func(error) {
	return func(err error) {
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(HavePrefix(prefix))
		Expect(err.Error()).Should(HaveSuffix(suffix))
	}
}

// InnerGErrorVerifier can be used for the case when the inner error is a GError
func InnerGErrorVerifier(env string, pkg string, file string, class string,
	function string, line int, innerVerifier func(error), message string,
	msgParts ...string) func(error) {
	return func(err error) {
		Expect(err).Should(HaveOccurred())
		gerr, ok := err.(*utils.GError)
		Expect(ok).Should(BeTrue())
		ErrorVerifier(env, pkg, file, class, function, line, innerVerifier,
			message, msgParts...)(gerr)
	}
}

// ErrorVerifier verifies the fields on a backend Error
func ErrorVerifier(env string, pkg string, file string, class string,
	function string, line int, innerVerifier func(error), message string,
	msgParts ...string) func(*utils.GError) {
	return func(err *utils.GError) {
		Expect(err.Class).Should(Equal(class))
		Expect(err.Environment).Should(Equal(env))
		Expect(err.File).Should(Equal(file))
		Expect(err.Function).Should(Equal(function))
		Expect(err.GeneratedAt).ShouldNot(BeNil())
		Expect(err.LineNumber).Should(Equal(line))
		Expect(err.Message).Should(Equal(message))
		Expect(err.Package).Should(Equal(pkg))
		innerVerifier(err.Inner)
		for _, part := range msgParts {
			Expect(err.Error()).Should(ContainSubstring(part))
		}
	}
}
