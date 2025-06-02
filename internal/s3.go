package internal

import (
	"bytes"
	"encoding/json"
	"os"
	"log"
	"io/ioutil"
	"errors"
	"strings"
	"time"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

var once sync.Once


// CreateS3Bucket Is used to create a Bucket in S3 when app is bought up.
// If the bucket already exists it moves on
func CreateS3Bucket() {
	once.Do(func() {
		awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		awsRegion := os.Getenv("AWS_REGION")
		bucketName := os.Getenv("S3_BUCKET")

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(awsRegion),
			Credentials: credentials.NewStaticCredentials(
				awsAccessKey, awsSecretKey, ""),
		})
		if err != nil {
			log.Fatalf("Failed to create AWS session: %v", err)
		}

		s3Client := s3.New(sess)
		log.Printf("S3 client initialized for region: %s", awsRegion)

		// Check if the bucket exists
		_, err = s3Client.HeadBucket(&s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		})

		if err == nil {
			log.Printf("Bucket '%s' already exists.", bucketName)
			return
		}

		// Check if the error is a "NotFound" error
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFound" || awsErr.Code() == "404" {
				log.Printf("Bucket '%s' not found. Creating it now...", bucketName)

				_, err := s3Client.CreateBucket(&s3.CreateBucketInput{
					Bucket: aws.String(bucketName),
					CreateBucketConfiguration: &s3.CreateBucketConfiguration{
						LocationConstraint: aws.String(awsRegion),
					},
				})
				if err != nil {
					log.Fatalf("Failed to create bucket: %v", err)
				}

				log.Printf("Bucket '%s' created successfully.", bucketName)
				return
			}
		}

		log.Fatalf("Error checking bucket: %v", err)
	})
}


//UploadToS3 is used to upload a Json file to an AWS S3 bucket, return error if any
func UploadToS3(filename string, data interface{}) error {

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	log.Printf("Aws access key id is %s...\n", os.Getenv("AWS_ACCESS_KEY_ID"))

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})
	if err != nil {
		return err
	}

	svc := s3.New(sess)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String("logs/" + filename),
		Body:   bytes.NewReader(jsonBytes),
	})

	//Create and upload metadata for the latest ingestion (using S3 bukcet only).
	metadata := map[string]interface{}{
		"filename":    filename,
		"ingested_at": time.Now().UTC().Format(time.RFC3339),
	}
	
	metadataBytes, _ := json.Marshal(metadata)
	err = UploadMetadataToS3("logs/latest_ingestion.json", metadataBytes)
	if err != nil {
		log.Printf("Warning: Failed to write latest_ingestion.json: %v", err)
	}

	return err
}

//UploadMetadataToS3 is used to upload metadata info of latest Ingestion to S3
func UploadMetadataToS3(filename string, data []byte) error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"), 
			os.Getenv("AWS_SECRET_ACCESS_KEY"), 
			"",
		),
	}))

	svc := s3.New(sess)
	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(data),
	})
	return err
}

//ReadJSONFromS3 is used to read the download and extract the json file contents from S3.
func ReadJSONFromS3(objectKey string) (interface{}, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	})
	if err != nil {
		return nil, err
	}

	s3Client := s3.New(sess)

	result, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String(objectKey),
	})
	
	if err != nil {
		// This handles file not found error.
		log.Printf("error is %s", err)
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == s3.ErrCodeNoSuchKey {
				return nil, errors.New("no_such_key")
			}
		}
		return nil, err
	}

	log.Printf("objectkey is %s",objectKey)
	log.Printf("result is %s",result)

	defer result.Body.Close()

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	// unmarshalling Based on the file (metatdata file or logs file).
	if strings.HasSuffix(objectKey, "logs/latest_ingestion.json") {

		var jsonData map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			log.Printf("Error unmarshalling object JSON: %v", err)
			return nil, err
		}
		log.Printf("jsonData is %s",jsonData)
		log.Printf("Read JSON from s3://%s/%s", os.Getenv("S3_BUCKET"), objectKey)
		return jsonData, nil

	} else {
		var jsonData []map[string]interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			log.Printf("Error unmarshalling array JSON: %v", err)
			return nil, err
		}
		log.Printf("jsonData is %s",jsonData)
		log.Printf("Read JSON from s3://%s/%s", os.Getenv("S3_BUCKET"), objectKey)
		return jsonData, nil
	}
}

//ListFilesFromS3 is used to List all the contents of the AWS S3 bucket
func ListFilesFromS3(prefix string) ([]string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	})
	if err != nil {
		return nil, err
	}

	s3Client := s3.New(sess)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Prefix: aws.String(prefix),
	}

	result, err := s3Client.ListObjectsV2(input)
	if err != nil {
		return nil, err
	}

	var filenames []string
	for _, item := range result.Contents {
		name := strings.TrimPrefix(*item.Key, prefix)
		if name != "" {
			filenames = append(filenames, name)
		}
	}

	return filenames, nil
}


// DeleteFileFromS3 is used to Delete a file from the specified AWS S3 bucket
func DeleteFileFromS3(filename string) error {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	s3Client := s3.New(sess)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String("logs/" + filename),
	}

	_, err = s3Client.DeleteObject(input)
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	log.Printf("Deleted s3://%s/logs/%s", os.Getenv("S3_BUCKET"), filename)
	return nil
}

