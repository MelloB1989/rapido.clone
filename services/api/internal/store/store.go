package store

import (
	"context"
	"errors"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// ErrNotFound is returned when a requested item does not exist.
var ErrNotFound = errors.New("not found")

// NewClient builds a DynamoDB client using the default AWS credential chain
// (the Lambda execution role in AWS, your aws-cli profile locally).
func NewClient(ctx context.Context, region string) (*dynamodb.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return dynamodb.NewFromConfig(cfg), nil
}
