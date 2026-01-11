package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

// SESConfig contains configuration for the AWS SES client
type SESConfig struct {
	Region    string
	AccessKey string // Optional - uses default credential chain if empty
	SecretKey string // Optional - uses default credential chain if empty
	FromEmail string
	FromName  string
}

// SESClient implements EmailSender using AWS SES v2
type SESClient struct {
	client    *sesv2.Client
	fromEmail string
	fromName  string
}

// NewSESClient creates a new AWS SES client
func NewSESClient(cfg SESConfig) (*SESClient, error) {
	ctx := context.Background()

	var awsCfg aws.Config
	var err error

	// Load AWS configuration
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		// Use explicit credentials if provided
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					cfg.AccessKey,
					cfg.SecretKey,
					"", // session token (not used for IAM users)
				),
			),
		)
	} else {
		// Use default credential chain (env vars, IAM role, etc.)
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("email: failed to load AWS config: %w", err)
	}

	return &SESClient{
		client:    sesv2.NewFromConfig(awsCfg),
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
	}, nil
}

// Send sends an email using AWS SES
func (c *SESClient) Send(ctx context.Context, params SendEmailParams) error {
	if len(params.To) == 0 {
		return fmt.Errorf("email: recipient list is empty")
	}
	if params.Subject == "" {
		return fmt.Errorf("email: subject is required")
	}
	if params.HTML == "" && params.Text == "" {
		return fmt.Errorf("email: either HTML or Text content is required")
	}

	// Build from address
	fromAddress := c.fromEmail
	if c.fromName != "" {
		fromAddress = fmt.Sprintf("%s <%s>", c.fromName, c.fromEmail)
	}

	// Build email content
	content := &types.EmailContent{
		Simple: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(params.Subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &types.Body{},
		},
	}

	// Add text body if provided
	if params.Text != "" {
		content.Simple.Body.Text = &types.Content{
			Data:    aws.String(params.Text),
			Charset: aws.String("UTF-8"),
		}
	}

	// Add HTML body if provided
	if params.HTML != "" {
		content.Simple.Body.Html = &types.Content{
			Data:    aws.String(params.HTML),
			Charset: aws.String("UTF-8"),
		}
	}

	// Build send request
	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(fromAddress),
		Destination: &types.Destination{
			ToAddresses: params.To,
		},
		Content: content,
	}

	// Add reply-to if provided
	if params.ReplyTo != "" {
		input.ReplyToAddresses = []string{params.ReplyTo}
	}

	// Send email
	_, err := c.client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("email: SES SendEmail failed: %w", err)
	}

	return nil
}

// IsConfigured returns true if the client has required configuration
func (c *SESClient) IsConfigured() bool {
	return c.client != nil && c.fromEmail != ""
}
