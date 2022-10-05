package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/matelang/jwt-go-aws-kms/v2/jwtkms"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	lkms "github.com/xefino/goutils/awssvc/kms"
	"github.com/xefino/goutils/awssvc/policy"
	"github.com/xefino/goutils/awssvc/testing"
	"github.com/xefino/goutils/utils"
)

var _ = Describe("KMS JWT Generate Tests", func() {

	// Tests that calling the Token function will return an error if the SignedString function returns an error
	It("Token - SignedString fails - Error", func() {

		// First, create our mock KMS client
		client := new(mockKMSClient)

		// Next, create a new KMS JWT access generator
		generate := NewKMSJWTAccessGenerate(client, "TEST_KEY_ID", jwtkms.SigningMethodECDSA512)

		// Now, create our claims data, including the client, user ID and creation time
		data := generateToken()

		// Finally, create the access token and refresh token; this should fail
		accessToken, refreshToken, err := generate.Token(context.Background(), data, true)

		// Verify the error that occurred
		Expect(accessToken).Should(BeEmpty())
		Expect(refreshToken).Should(BeEmpty())
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("Sign failed"))
	})

	// Tests that calling the Token function will return an access token and refresh token if no error occurs
	// when the Token function is called
	It("Token - No failures - Tokens returned", func() {

		// First, generate the AWS config
		cfg := testing.TestAWSConfig(context.Background(), "eu-west-2", 8080)

		// Next, create our test KMS connection
		logger := utils.NewLogger("testd", "test")
		logger.Discard()
		conn := lkms.NewConnection(cfg, logger)

		// Now, generate the policy for the key we want to use for testing
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

		// Create the KMS key from the policy; this should not fail
		metadata, err := conn.CreateKey(context.Background(), types.KeySpecEccNistP521,
			types.KeyUsageTypeSignVerify, &policy, true)
		Expect(err).ShouldNot(HaveOccurred())

		// Finally, attempt to use the key to generate our access token and refresh token; this should not fail
		generate := NewKMSJWTAccessGenerate(kms.NewFromConfig(cfg), *metadata.KeyId, jwtkms.SigningMethodECDSA512)
		accessToken, refreshToken, err := generate.Token(context.Background(), generateToken(), true)
		Expect(err).ShouldNot(HaveOccurred())

		// Verify the access token and refresh token
		Expect(accessToken).ShouldNot(BeEmpty())
		Expect(refreshToken).ShouldNot(BeEmpty())
	})
})

// Helper type that we'll use to test our KMS client code
type mockKMSClient struct {
	jwtkms.KMSClient
}

// Mock out the function to sign a JWT, verifying the input options in the process
func (client *mockKMSClient) Sign(ctx context.Context, in *kms.SignInput,
	optFns ...func(*kms.Options)) (*kms.SignOutput, error) {

	// Verify the input fields and options
	Expect(*in.KeyId).Should(Equal("TEST_KEY_ID"))
	Expect(in.Message).ShouldNot(BeEmpty())
	Expect(in.MessageType).Should(Equal(types.MessageTypeDigest))
	Expect(in.SigningAlgorithm).Should(Equal(types.SigningAlgorithmSpecEcdsaSha512))
	Expect(optFns).Should(BeEmpty())

	// We want to test around Sign failure then return an error
	return nil, fmt.Errorf("Sign failed")
}

// Helper type that will operate as a test auth client
type testClient struct{}

// Mock out the function to get the client domain
func (c *testClient) GetDomain() string {
	return "test.domain.com"
}

// Mock out the function to get the client ID
func (c *testClient) GetID() string {
	return "test_client"
}

// Mock out the function to get the client secret
func (c *testClient) GetSecret() string {
	return "AAAAAAAAAAAA"
}

// Mock out the function to get the user ID
func (c *testClient) GetUserID() string {
	return "test_user"
}

// Helper function that creates a test token data
func generateToken() *oauth2.GenerateBasic {

	// Create our claims data, including the client, user ID and creation time
	data := oauth2.GenerateBasic{
		Client:    new(testClient),
		UserID:    "test_user",
		CreateAt:  time.Date(2022, time.October, 3, 9, 24, 0, 0, time.UTC),
		TokenInfo: models.NewToken(),
	}

	// Set the token info with our test values
	data.TokenInfo.SetClientID(data.Client.GetID())
	data.TokenInfo.SetUserID(data.Client.GetUserID())
	data.TokenInfo.SetScope("test_scope")
	data.TokenInfo.SetAccessCreateAt(data.CreateAt)
	data.TokenInfo.SetAccessExpiresIn(time.Hour)
	data.TokenInfo.SetRefreshCreateAt(data.CreateAt)
	data.TokenInfo.SetAccessExpiresIn(time.Hour)

	// Return the data
	return &data
}
