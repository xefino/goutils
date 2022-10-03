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
)

var _ = Describe("KMS JWT Generate Tests", func() {

	// Tests that calling the Token function will return an error if the SignedString function returns an error
	It("Token - SignedString fails - Error", func() {

		// First, create our mock KMS client
		client := mockKMSClient{ShouldFail: true}

		// Next, create a new KMS JWT access generator
		generate := NewKMSJWTAccessGenerate(&client, "TEST_KEY_ID", jwtkms.SigningMethodECDSA512)

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

	// This code needs additional testing but I'm not sure how to build around the Sign function so
	// that it performs a decent facsimile of the actual functionality when it's called
})

// Helper type that we'll use to test our KMS client code
type mockKMSClient struct {
	jwtkms.KMSClient
	ShouldFail bool
}

// Mock out the function to sign a JWT, verifying the input options in the process
func (client *mockKMSClient) Sign(ctx context.Context, in *kms.SignInput,
	optFns ...func(*kms.Options)) (*kms.SignOutput, error) {

	// First, verify the input fields and options
	Expect(*in.KeyId).Should(Equal("TEST_KEY_ID"))
	Expect(in.Message).ShouldNot(BeEmpty())
	Expect(in.MessageType).Should(Equal(types.MessageTypeDigest))
	Expect(in.SigningAlgorithm).Should(Equal(types.SigningAlgorithmSpecEcdsaSha512))
	Expect(optFns).Should(BeEmpty())

	// Next, if we want to test around Sign failure then return an error
	if client.ShouldFail {
		return nil, fmt.Errorf("Sign failed")
	}

	// Finally, return a test sign output
	data := append(in.Message, []byte(".TEST_SIGNATURE")...)
	return &kms.SignOutput{
		KeyId:            in.KeyId,
		Signature:        data,
		SigningAlgorithm: types.SigningAlgorithmSpecEcdsaSha512,
	}, nil
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
