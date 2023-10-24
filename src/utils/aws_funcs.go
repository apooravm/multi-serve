package utils

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Download the resource file from the given s3 bucket
func DownloadFile(bucketName string, objPath string, region string) ([]byte, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	output, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objPath),
	})

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer output.Body.Close()

	body, err := io.ReadAll(output.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return body, nil
}
