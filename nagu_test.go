package nagu

import "testing"

func TestS3(t *testing.T) {
	s3 := S3(nil)

	_ = s3.Object("bucket", "key")
}

func TestS3BucketName(t *testing.T) {
	s3 := S3(nil)

	object := s3.Object("bucket", "key")

	if object.BucketName() != "bucket" {
		t.Errorf("Incorrect bucket name in S3 object, Expected: %v, Got: %v", "bucket", object.BucketName())
	}

}

func TestS3Key(t *testing.T) {
	s3 := S3(nil)

	object := s3.Object("bucket", "key")

	if object.Key() != "key" {
		t.Errorf("Incorrect key in S3 object, Expected: %v, Got: %v", "key", object.Key())
	}

}
