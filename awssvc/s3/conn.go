package s3

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	oaws "github.com/aws/aws-sdk-go/aws"
	"github.com/xefino/goutils/utils"
)

// IS3Connection describes the functionality encapsulated in an S3 connection
type IS3Connection interface {
	DownloadToStream(ctx context.Context, bucket string, key string) (io.Writer, error)
	UploadFromStream(ctx context.Context, bucket string, key string, body io.Reader) error
}

// S3Connection wraps functionality necessary to communicate with AWS S3
type S3Connection struct {
	inner  S3API
	logger *utils.Logger
}

// NewS3Connection creates a new S3 connection from an AWS session and a logger
func NewS3Connection(cfg aws.Config, logger *utils.Logger) *S3Connection {
	return &S3Connection{
		inner:  s3.NewFromConfig(cfg),
		logger: logger,
	}
}

// DownloadToStream retrieves a file from S3 and downloads it to a stream so we can work with it
func (conn *S3Connection) DownloadToStream(ctx context.Context, bucket string, key string) (io.Writer, error) {
	conn.logger.Log("Attempting to download %s from %s in S3...", key, bucket)

	// First, create a new S3 manager from our inner S3 connection
	downloader := manager.NewDownloader(conn.inner)

	// Next, create a new download input with our body, bucket name and key
	input := s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	// Now, download the file from S3; if this fails then generate an error
	buffer := oaws.NewWriteAtBuffer(make([]byte, 0))
	if _, err := downloader.Download(ctx, buffer, &input); err != nil {
		return nil, conn.logger.Error(err, "Failed to download from %s in %s in S3", key, bucket)
	}

	// Finally, write the data in the buffer to a new buffer and return it
	return bytes.NewBuffer(buffer.Bytes()), nil
}

// UploadFromStream writes data in a stream to a file in S3
func (conn *S3Connection) UploadFromStream(ctx context.Context, bucket string, key string, body io.Reader) error {
	conn.logger.Log("Attempting to upload %s to %s in S3...", key, bucket)

	// First, create a new S3 manager from our inner S3 connection
	uploader := manager.NewUploader(conn.inner)

	// Next, create a new upload input with our body, bucket name and key
	input := s3.PutObjectInput{
		Body:   body,
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	// Finally, upload the file to S3; if this fails then generate an error
	_, err := uploader.Upload(ctx, &input)
	if err != nil {
		return conn.logger.Error(err, "Failed to upload %s to %s in S3", key, bucket)
	}

	return nil
}
