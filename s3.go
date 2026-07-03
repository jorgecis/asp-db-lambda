package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// downloadFromS3 downloads the object at bucket/key into a local temp file and
// returns the temp file path. The caller is responsible for removing it.
func downloadFromS3(bucket, key string) (string, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")}))

	downloader := s3manager.NewDownloader(sess)
	file, err := os.CreateTemp("", "aspfile")
	if err != nil {
		log.Printf("Error: %s \r\n", err)
		return "", err
	}

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		log.Printf("Unable to download item %q, %v\r\n", key, err)
		return "", err
	}

	log.Println("Downloaded", file.Name(), numBytes, "bytes")
	return file.Name(), nil
}
