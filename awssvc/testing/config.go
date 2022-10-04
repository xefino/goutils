package testing

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// TestAWSConfig creates an AWS config that can be used for local testing
func TestAWSConfig(ctx context.Context, region string, port int) aws.Config {

	// First, attempt to create the config with our loader functions
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{URL: fmt.Sprintf("http://localhost:%d", port)}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
				Source: "Hard-coded credentials; values are irrelevant for local AWS",
			},
		}),
	)

	// Next, check if the creation failed; if it did then panic
	if err != nil {
		panic(err)
	}

	// Finally, return the config
	return cfg
}
