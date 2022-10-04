package kms

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/xefino/goutils/awssvc/policy"
	"github.com/xefino/goutils/utils"
)

// KMSConnection allows the user to handle requests made against KMS in a standard fashion
type KMSConnection struct {
	inner  KMSAPI
	logger *utils.Logger
}

// NewKMSConnection creates a new KMS connection from an AWS session and a logger
func NewKMSConnection(cfg aws.Config, logger *utils.Logger) *KMSConnection {
	return &KMSConnection{
		inner:  kms.NewFromConfig(cfg),
		logger: logger,
	}
}

// CreateKey creates a new key in KMS with the specified spec, usage and policy document
func (conn *KMSConnection) CreateKey(ctx context.Context, spec types.KeySpec, usage types.KeyUsageType,
	doc *policy.Policy, multiRegion bool) (*types.KeyMetadata, error) {

	// First, attempt to serialize the policy to JSON; if this fails then return an error
	policy, err := json.Marshal(doc)
	if err != nil {
		return nil, conn.logger.Error(err, "Failed to convert policy to JSON")
	}

	// Next, create the key input from the spec, usage and policy
	input := kms.CreateKeyInput{
		KeySpec:     spec,
		KeyUsage:    usage,
		MultiRegion: aws.Bool(multiRegion),
		Policy:      aws.String(string(policy)),
	}

	// Now, attempt to create the key from the input; if this fails then return an error
	out, err := conn.inner.CreateKey(ctx, &input)
	if err != nil {
		return nil, conn.logger.Error(err, "Failed to create %s (%s) key in KMS", spec, usage)
	}

	// Finally, return the metadata of the key we created
	return out.KeyMetadata, nil
}
