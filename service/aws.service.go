package service

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewAWSService() *s3.Client {
	return s3.NewFromConfig(aws.Config{
		Region:      os.Getenv("AWS_REGION"),
		Credentials: credentials.NewStaticCredentialsProvider(os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET"), ""),
	})
}
