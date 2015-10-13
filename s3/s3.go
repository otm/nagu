package s3

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

import s3client "github.com/aws/aws-sdk-go/service/s3"

// S3 is the service resource
type S3 struct {
	srv *s3client.S3
}

// New returns a new nagu.Cloudformation object
func New(config *aws.Config) *S3 {
	s3 := &S3{}
	s3.srv = s3client.New(config)
	return s3
}

// Object creates a new S3 object
func (s3 *S3) Object(bucket, key string) *Object {
	o := &Object{
		srv:        s3.srv,
		bucketName: bucket,
		key:        key,
	}

	return o
}

// Object is a S3 object
type Object struct {
	srv        *s3client.S3
	bucketName string
	key        string

	AcceptRanges            string
	CacheControl            string
	ContentDisposition      string
	ContentEncoding         string
	ContentLanguage         string
	ContentLength           int64
	ContentType             string
	DeleteMarker            bool
	ETag                    string
	Expiration              string
	Expires                 string
	LastModified            time.Time
	Metadata                map[string]*string
	MissingMeta             int64
	ReplicationStatus       string
	RequestCharged          string
	Restore                 string
	ServerSideEncryption    string
	SSECustomerAlgorithm    string
	SSECustomerKeyMD5       string
	SSEKMSKeyID             string
	StorageClass            string
	VersionID               string
	WebsiteRedirectLocation string
}

// BucketName returns the name of the bucket where the object resides
func (o *Object) BucketName() string {
	return o.bucketName
}

// Key returns the key of the S3 object in the bucket
func (o *Object) Key() string {
	return o.key
}

func (o *Object) load() error {
	resp, err := o.srv.HeadObject(&s3client.HeadObjectInput{Bucket: &o.bucketName, Key: &o.key})
	if err != nil {
		return err
	}

	o.AcceptRanges = *resp.AcceptRanges
	o.CacheControl = *resp.CacheControl
	o.ContentDisposition = *resp.ContentDisposition
	o.ContentEncoding = *resp.ContentEncoding
	o.ContentLanguage = *resp.ContentLanguage
	o.ContentLength = *resp.ContentLength
	o.ContentType = *resp.ContentType
	o.DeleteMarker = *resp.DeleteMarker
	o.ETag = *resp.ETag
	o.Expiration = *resp.Expiration
	o.Expires = *resp.Expires
	o.LastModified = *resp.LastModified
	o.Metadata = resp.Metadata
	o.MissingMeta = *resp.MissingMeta
	o.ReplicationStatus = *resp.ReplicationStatus
	o.RequestCharged = *resp.RequestCharged
	o.Restore = *resp.Restore
	o.ServerSideEncryption = *resp.ServerSideEncryption
	o.SSECustomerAlgorithm = *resp.SSECustomerAlgorithm
	o.SSECustomerKeyMD5 = *resp.SSECustomerKeyMD5
	o.SSEKMSKeyID = *resp.SSEKMSKeyId
	o.StorageClass = *resp.StorageClass
	o.VersionID = *resp.VersionId
	o.WebsiteRedirectLocation = *resp.WebsiteRedirectLocation
	return nil
}

func (o *Object) reload() error {
	return o.load()
}

func (o *Object) get() (body io.ReadCloser, err error) {
	resp, err := o.srv.GetObject(&s3client.GetObjectInput{Bucket: &o.bucketName, Key: &o.key})
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (o *Object) delete() error {
	_, err := o.srv.DeleteObject(&s3client.DeleteObjectInput{Bucket: &o.bucketName, Key: &o.key})
	return err
}

func (o *Object) put(body io.ReadSeeker) error {
	_, err := o.srv.PutObject(&s3client.PutObjectInput{Bucket: &o.bucketName, Key: &o.key, Body: body})
	return err
}

func (o *Object) copy(dst string) (copy *Object, err error) {
	parts := strings.SplitN(dst, "/", 2)

	_, err = o.srv.CopyObject(&s3client.CopyObjectInput{CopySource: aws.String(o.bucketName + "/" + o.key), Bucket: &parts[0], Key: &parts[1]})

	copy = &Object{srv: o.srv, bucketName: o.bucketName, key: o.key}
	return copy, err
}

func (o *Object) copyFrom(src *Object) error {
	_, err := o.srv.CopyObject(&s3client.CopyObjectInput{CopySource: aws.String(src.bucketName + "/" + src.key), Bucket: &o.bucketName, Key: &o.key})
	o.reload()

	return err
}

func (o *Object) waitUntilExists() error {
	for i := 0; i < 20; i++ {
		_, err := o.srv.HeadObject(&s3client.HeadObjectInput{Bucket: &o.bucketName, Key: &o.key})
		if err == nil {
			o.reload()
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timeout")
}

func (o *Object) waitUntilNotExists() error {
	for i := 0; i < 20; i++ {
		_, err := o.srv.HeadObject(&s3client.HeadObjectInput{Bucket: &o.bucketName, Key: &o.key})
		if err != nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timeout")
}
