package services

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type UserService struct {
	avatarUrlFormat string
	bucketName      string

	S3Client *s3.Client
}

func (us *UserService) Init(ctx context.Context) error {
	accessKey, found := os.LookupEnv("AWS_ACCESS_KEY")
	if !found {
		return errors.New("missing AWS_ACCESS_KEY environment variable")
	}

	secretAccessKey, found := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if !found {
		return errors.New("missing AWS_SECRET_ACCESS_KEY environment variable")
	}

	bucketName, found := os.LookupEnv("AWS_AVATAR_S3_BUCKET")
	if !found {
		return errors.New("missing AWS_AVATAR_S3_BUCKET environment variable")
	}
	us.bucketName = bucketName

	avatarUrlFormat, found := os.LookupEnv("AWS_AVATAR_URL_FORMAT")
	if !found {
		return errors.New("missing AWS_AVATAR_URL_FORMAT environment variable")
	}
	us.avatarUrlFormat = avatarUrlFormat

	// Get S3 client.
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	cfg.Credentials = aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     accessKey,
			SecretAccessKey: secretAccessKey,
		}, nil
	})

	us.S3Client = s3.NewFromConfig(cfg)

	return nil
}
