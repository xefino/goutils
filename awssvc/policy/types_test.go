package policy

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Create a new test runner we'll use to test all the
// modules in the policy package
func TestPolicy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Policy Suite")
}

var _ = Describe("Policy Tests", func() {

	// Test the conditions under which the types will fail to marshal to JSON
	DescribeTable("MarshalJSON - Failures",
		func(principal []string, action Actions, resource []string, message string) {

			// First, create the policy document using our test data
			policy := Policy{
				Version: "2012-10-17",
				ID:      "test-policy",
				Statements: []*Statement{
					{
						ID:            "Fail marshal tests",
						Effect:        Allow,
						PrincipalArns: principal,
						ActionArns:    action,
						ResourceArns:  resource,
					},
				},
			}

			// Next, attempt to marshal the policy to JSON; this should fail
			data, err := json.Marshal(policy)

			// Finally, verify the failure
			Expect(data).Should(BeEmpty())
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(Equal(message))
		},
		Entry("Principal empty - Error", nil, Actions{CreateKey, Sign, Encrypt}, []string{"*"},
			"json: error calling MarshalJSON for type policy.Principals: Principal must contain at least one element"),
		Entry("Action empty - Error", []string{"arn:aws:iam::111122223333:root"}, nil, []string{"*"},
			"json: error calling MarshalJSON for type policy.Actions: Action must contain at least one element"),
		Entry("Resource empty - Error", []string{"arn:aws:iam::111122223333:root"}, Actions{CreateKey, Sign, Encrypt}, nil,
			"json: error calling MarshalJSON for type policy.Resources: Resource must contain at least one element"))

	// Test the conditions determining how data should be converted to JSON
	DescribeTable("MarshalJSON - Conditions",
		func(principal []string, action Actions, resource []string, result string) {

			// First, create the policy document using our test data
			policy := Policy{
				Version: "2012-10-17",
				ID:      "test-policy",
				Statements: []*Statement{
					{
						ID:            "Pass marshal tests",
						Effect:        Allow,
						PrincipalArns: principal,
						ActionArns:    action,
						ResourceArns:  resource,
					},
				},
			}

			// Next, attempt to marshal the policy to JSON; this should not fail
			data, err := json.Marshal(policy)

			// Finally, verify the data
			Expect(string(data)).Should(Equal(result))
			Expect(err).ShouldNot(HaveOccurred())
		},
		Entry("All asterisks - Works", []string{"*"}, Actions{KmsAll}, []string{"*"},
			"{\"Version\":\"2012-10-17\",\"Id\":\"test-policy\",\"Statement\":[{\"Sid\":\"Pass marshal tests\","+
				"\"Effect\":\"Allow\",\"Principal\":{\"AWS\":\"*\"},\"Action\":\"kms:*\",\"Resource\":\"*\"}]}"),
		Entry("Multiple principals - Works",
			[]string{"arn:aws:iam::111122223333:user/ExampleUser", "arn:aws:iam::111122223333:role/ExampleRole"},
			Actions{KmsAll}, []string{"*"}, "{\"Version\":\"2012-10-17\",\"Id\":\"test-policy\",\"Statement\":["+
				"{\"Sid\":\"Pass marshal tests\",\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"arn:aws:iam::"+
				"111122223333:user/ExampleUser\",\"arn:aws:iam::111122223333:role/ExampleRole\"]},"+
				"\"Action\":\"kms:*\",\"Resource\":\"*\"}]}"),
		Entry("Multiple actions - Works",
			[]string{"arn:aws:iam::111122223333:user/ExampleUser", "arn:aws:iam::111122223333:role/ExampleRole"},
			Actions{CreateKey, Sign, Verify}, []string{"*"}, "{\"Version\":\"2012-10-17\",\"Id\":\"test-policy\","+
				"\"Statement\":[{\"Sid\":\"Pass marshal tests\",\"Effect\":\"Allow\",\"Principal\":{\"AWS\":["+
				"\"arn:aws:iam::111122223333:user/ExampleUser\",\"arn:aws:iam::111122223333:role/ExampleRole\"]},"+
				"\"Action\":[\"kms:CreateKey\",\"kms:Sign\",\"kms:Verify\"],\"Resource\":\"*\"}]}"),
		Entry("Multiple resources - Works",
			[]string{"arn:aws:iam::111122223333:user/ExampleUser", "arn:aws:iam::111122223333:role/ExampleRole"},
			Actions{CreateKey, Sign, Verify}, []string{"*", "kms:test"}, "{\"Version\":\"2012-10-17\","+
				"\"Id\":\"test-policy\",\"Statement\":[{\"Sid\":\"Pass marshal tests\",\"Effect\":\"Allow\","+
				"\"Principal\":{\"AWS\":[\"arn:aws:iam::111122223333:user/ExampleUser\",\"arn:aws:iam::111122223333:"+
				"role/ExampleRole\"]},\"Action\":[\"kms:CreateKey\",\"kms:Sign\",\"kms:Verify\"],\"Resource\":"+
				"[\"*\",\"kms:test\"]}]}"))
})
