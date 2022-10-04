package kms

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/xefino/goutils/awssvc/policy"
)

type KMSConnection struct {
	inner kms.Client
}

func (conn *KMSConnection) CreateKey(ctx context.Context, spec types.KeySpec, usage types.KeyUsageType,
	doc *policy.Policy, multiRegion bool) (*types.KeyMetadata, error) {

	policy, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	input := kms.CreateKeyInput{
		KeySpec:     spec,
		KeyUsage:    usage,
		MultiRegion: aws.Bool(multiRegion),
		Policy:      aws.String(string(policy)),
	}

	out, err := conn.inner.CreateKey(ctx, &input)
	if err != nil {
		return nil, err
	}

	return out.KeyMetadata, nil
}
