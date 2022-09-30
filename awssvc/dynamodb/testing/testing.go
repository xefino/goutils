package testing

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// TestDynamoDBConfig creates an AWS config that can be used for local DynamoDB testing
func TestDynamoDBConfig(ctx context.Context, region string, port int) aws.Config {

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
				Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
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

// EnsureTableExists ensures that the table exists for testing purposes
func EnsureTableExists(ctx context.Context, cfg aws.Config, table *dynamodb.CreateTableInput) error {

	// First, create a connection to our local DynamoDB
	client := dynamodb.NewFromConfig(cfg)

	// Next, attempt to get the table description associated with the table name
	output, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: table.TableName,
	})

	// Now, if we got an error then check if it was a resource-not-found exception. If
	// it was then that means we should create the table; otherwise, it means that something
	// isn't right so return it. If the description was nil, we'll also create the table
	var create bool
	if err != nil {
		if temp := new(types.ResourceNotFoundException); !errors.As(err, &temp) {
			return err
		} else {
			create = true
		}
	} else if output == nil {
		create = true
	}

	// Finally, if we want to create the table then do so here; return any error that occurs
	if create {
		_, err := client.CreateTable(ctx, table)
		if err != nil {
			return err
		}
	}

	return nil
}

// EmptyTable ensures that the table is in pristine condition for testing
func EmptyTable(ctx context.Context, cfg aws.Config, table *dynamodb.CreateTableInput) error {

	// First, create a connection to our local DynamoDB
	client := dynamodb.NewFromConfig(cfg)

	// Next, attempt to delete the table
	_, err := client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: table.TableName,
	})

	// If we got an error, check if it was a resource not found exception. In such cases, the table
	// already doesn't exist so DynamoDB is in the state we want; otherwise, return the error
	if err != nil {
		if _, ok := err.(*types.ResourceNotFoundException); !ok {
			return err
		}
	}

	// Finally, attempt to create the table; return any errors that occur
	_, err = client.CreateTable(ctx, table)
	return err
}
