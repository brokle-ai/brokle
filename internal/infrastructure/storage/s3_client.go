package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
)

// S3Client wraps AWS S3 SDK for blob storage operations
type S3Client struct {
	client     *s3.Client
	logger     *logrus.Logger
	bucketName string
}

// NewS3Client creates a new S3 client instance
func NewS3Client(cfg *config.BlobStorageConfig, logger *logrus.Logger) (*S3Client, error) {
	// Build AWS config
	var awsCfg aws.Config
	var err error

	// Check if using custom endpoint (MinIO, LocalStack, etc.)
	if cfg.Endpoint != "" {
		// Custom endpoint with static credentials
		awsCfg, err = awsConfig.LoadDefaultConfig(context.Background(),
			awsConfig.WithRegion(cfg.Region),
			awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}

		// Override endpoint for MinIO/custom S3
		awsCfg.BaseEndpoint = aws.String(cfg.Endpoint)
	} else {
		// Standard AWS S3 (uses default credential chain)
		if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
			awsCfg, err = awsConfig.LoadDefaultConfig(context.Background(),
				awsConfig.WithRegion(cfg.Region),
				awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
					cfg.AccessKeyID,
					cfg.SecretAccessKey,
					"",
				)),
			)
		} else {
			// Use default credential chain (IAM role, env vars, etc.)
			awsCfg, err = awsConfig.LoadDefaultConfig(context.Background(),
				awsConfig.WithRegion(cfg.Region),
			)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		// Use path-style addressing for MinIO compatibility
		o.UsePathStyle = cfg.UsePathStyle
	})

	logger.WithFields(logrus.Fields{
		"provider":   cfg.Provider,
		"bucket":     cfg.BucketName,
		"region":     cfg.Region,
		"endpoint":   cfg.Endpoint,
		"path_style": cfg.UsePathStyle,
	}).Info("S3 client initialized")

	return &S3Client{
		client:     s3Client,
		bucketName: cfg.BucketName,
		logger:     logger,
	}, nil
}

// Upload uploads content to S3
func (c *S3Client) Upload(ctx context.Context, key string, content []byte, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	}

	_, err := c.client.PutObject(ctx, input)
	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"bucket": c.bucketName,
			"key":    key,
		}).Error("Failed to upload to S3")
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"bucket": c.bucketName,
		"key":    key,
		"size":   len(content),
	}).Debug("Successfully uploaded to S3")

	return nil
}

// Download downloads content from S3
func (c *S3Client) Download(ctx context.Context, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	result, err := c.client.GetObject(ctx, input)
	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"bucket": c.bucketName,
			"key":    key,
		}).Error("Failed to download from S3")
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read S3 object body")
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"bucket": c.bucketName,
		"key":    key,
		"size":   len(content),
	}).Debug("Successfully downloaded from S3")

	return content, nil
}

// Delete deletes an object from S3
func (c *S3Client) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	_, err := c.client.DeleteObject(ctx, input)
	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"bucket": c.bucketName,
			"key":    key,
		}).Error("Failed to delete from S3")
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"bucket": c.bucketName,
		"key":    key,
	}).Debug("Successfully deleted from S3")

	return nil
}

// Exists checks if an object exists in S3
func (c *S3Client) Exists(ctx context.Context, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	_, err := c.client.HeadObject(ctx, input)
	if err != nil {
		// Check if error is "not found"
		return false, nil
	}

	return true, nil
}

// GetS3URI returns the full S3 URI for a key
func (c *S3Client) GetS3URI(key string) string {
	return fmt.Sprintf("s3://%s/%s", c.bucketName, key)
}
