package nagu

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/otm/nagu/cloudformation"
	"github.com/otm/nagu/s3"
)

// Cloudformation return a nagu cloudformation resource
func Cloudformation(config *aws.Config) *cloudformation.CloudFormation {
	return cloudformation.New(config)
}

// S3 returns a nagp s3 resource
func S3(config *aws.Config) *s3.S3 {
	return s3.New(config)
}
