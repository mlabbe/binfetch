package objstore

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3DownloadArchive downloads the file at archiveKey from the bucket
// and writes it into the open fileHandle.
func S3DownloadArchive(region, archiveKey, bucket string, fileHandle *os.File) error {
	sess, err := session.NewSession(&aws.Config{
		Region: &region,
	})
	if err != nil {
		return err
	}
	downloader := s3manager.NewDownloader(sess)

	_, err = downloader.Download(fileHandle,
		&s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &archiveKey})
	if err != nil {
		return err
	}

	return nil
}
