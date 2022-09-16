package testutils

import (
	"database/sql/driver"

	. "github.com/onsi/gomega"
)

// StringMatch allows for matching a string argument based on a prefix and suffix
type StringMatch struct {
	Prefix   string
	Suffix   string
	Contains []string
}

// NewStringMatch creates a new string match from the expected prefix, suffix
// and any values that should be contained in the string
func NewStringMatch(prefix string, suffix string, contains ...string) StringMatch {
	return StringMatch{
		Prefix:   prefix,
		Suffix:   suffix,
		Contains: contains,
	}
}

// Match checks that an SQL value matches the string provided
func (match StringMatch) Match(v driver.Value) bool {

	// First, check that the value can be converted
	value, ok := v.(string)
	Expect(ok).Should(BeTrue())

	// Next, verify the prefix and suffix
	Expect(value).Should(HavePrefix(match.Prefix))
	Expect(value).Should(HaveSuffix(match.Suffix))

	// Finally, verify all the contain terms
	for _, term := range match.Contains {
		Expect(value).Should(ContainSubstring(term))
	}

	return ok
}
