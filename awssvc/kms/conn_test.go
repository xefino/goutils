package kms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/xefino/goutils/awssvc/policy"
	"github.com/xefino/goutils/awssvc/testing"
	"github.com/xefino/goutils/testutils"
	"github.com/xefino/goutils/utils"
)

var _ = Describe("KMS Connection Tests", Ordered, func() {

	// Ensure that the AWS config is created before each test; this could be set as a global variable
	var cfg aws.Config
	BeforeAll(func() {
		cfg = testing.TestAWSConfig(context.Background(), "eu-west-2", 8080)
	})

	// Tests that, if the policy document fails to marshal to JSON when CreateKey is called, then an
	// error will be returned
	It("CreateKey - MarshalJSON fails - Error", func() {

		// First, create our test KMS connection
		logger := utils.NewLogger("testd", "test")
		logger.Discard()
		kms := NewKMSConnection(cfg, logger)

		// Next, create our test KMS key policy (this will fail to marshal)
		policy := policy.Policy{
			Version: "2012-10-17",
			ID:      "test-key",
			Statements: []*policy.Statement{
				{
					ID:           "test-failure",
					Effect:       policy.Allow,
					ActionArns:   policy.Actions{policy.KmsAll},
					ResourceArns: []string{"*"},
				},
			},
		}

		// Now, create the KMS key with our policy; this should fail
		meta, err := kms.CreateKey(context.Background(), types.KeySpecEccNistP521,
			types.KeyUsageTypeSignVerify, &policy, true)

		// Finally, verify the details of the error
		Expect(meta).Should(BeNil())
		Expect(err).Should(HaveOccurred())
		testutils.ErrorVerifier("test", "kms", "/goutils/awssvc/kms/conn.go", "KMSConnection", "CreateKey", 35,
			testutils.InnerErrorVerifier("json: error calling MarshalJSON for type policy.Principals: "+
				"Principal must contain at least one element"), "Failed to convert policy to JSON",
			"[test] kms.KMSConnection.CreateKey (/goutils/awssvc/kms/conn.go 35): Failed to convert "+
				"policy to JSON, Inner: json: error calling MarshalJSON for type policy.Principals: "+
				"Principal must contain at least one element.")(err.(*utils.GError))
	})

	// Tests that, if the call to CreateKey in KMS fails, then calling CreateKey will return an error
	It("CreateKey - CreateKey inner fails - Error", func() {

		// First, create our test KMS connection
		logger := utils.NewLogger("testd", "test")
		logger.Discard()
		kms := NewKMSConnection(cfg, logger)

		// Next, create our test KMS key policy (this will fail to marshal)
		policy := policy.Policy{
			Version: "2012-10-17",
			ID:      "test-key",
			Statements: []*policy.Statement{
				{
					ID:            "test-failure",
					Effect:        policy.Allow,
					PrincipalArns: []string{"arn:aws:kms:eu-west-2:111122223333:root"},
					ActionArns:    policy.Actions{policy.KmsAll},
					ResourceArns:  []string{"*"},
				},
			},
		}

		// Now, create the KMS key with our policy; this should fail
		meta, err := kms.CreateKey(context.Background(), types.KeySpecEccNistP521,
			types.KeyUsageTypeEncryptDecrypt, &policy, true)

		// Finally, verify the details of the error
		Expect(meta).Should(BeNil())
		Expect(err).Should(HaveOccurred())
		testutils.ErrorVerifier("test", "kms", "/goutils/awssvc/kms/conn.go", "KMSConnection", "CreateKey", 49,
			testutils.InnerErrorVerifier("operation error KMS: CreateKey, https response error StatusCode: "+
				"400, RequestID: , api error ValidationException: KeyUsage ENCRYPT_DECRYPT is not compatible "+
				"with KeySpec ECC_NIST_P521"), "Failed to create ECC_NIST_P521 (ENCRYPT_DECRYPT) key in KMS",
			"[test] kms.KMSConnection.CreateKey (/goutils/awssvc/kms/conn.go 49): Failed to create "+
				"ECC_NIST_P521 (ENCRYPT_DECRYPT) key in KMS, Inner: operation error KMS: CreateKey, "+
				"https response error StatusCode: 400, RequestID: , api error ValidationException: "+
				"KeyUsage ENCRYPT_DECRYPT is not compatible with KeySpec ECC_NIST_P521.")(err.(*utils.GError))
	})
})
