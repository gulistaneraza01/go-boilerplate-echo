package aws

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/config"
)

// AWS groups every AWS service client the app uses.
type AWS struct {
	S3 *Storage
}

const startupCheckTimeout = 5 * time.Second

func NewAWS(ctx context.Context, cfg config.AWSConfig) (*AWS, error) {
	awsCfg, err := NewConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	client := NewS3Client(awsCfg, cfg.S3)

	checkCtx, cancel := context.WithTimeout(ctx, startupCheckTimeout)
	defer cancel()
	if _, err := client.HeadBucket(checkCtx, &s3.HeadBucketInput{Bucket: aws.String(cfg.S3.Bucket)}); err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return nil, fmt.Errorf("S3 bucket %q not found: check BOILERPLATE_AWS.S3.BUCKET and BOILERPLATE_AWS.REGION", cfg.S3.Bucket)
		}
		return nil, fmt.Errorf("S3 bucket %q unreachable: %w", cfg.S3.Bucket, err)
	}

	return &AWS{
		S3: NewStorage(client, cfg.S3.Bucket),
	}, nil
}

// NewConfig loads region and credentials once; pass the result to every AWS service client (S3, SES, SQS, ...).
func NewConfig(ctx context.Context, cfg config.AWSConfig) (aws.Config, error) {
	opts := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(cfg.Region)}
	if cfg.AccessKeyID != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load aws config: %w", err)
	}
	return awsCfg, nil
}
