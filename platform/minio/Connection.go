package minio

import (
	"context"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

// MinioConnection func for opening minio connection.
func MinioConnection() (*minio.Client, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESSKEY")
	secretAccessKey := os.Getenv("MINIO_SECRETKEY")
	useSSL := false
	// Initialize minio client object.
	if minioClient != nil {
		return minioClient, nil
	}
	log.Println(endpoint)
	client, errInit := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	minioClient = client
	if errInit != nil {
		log.Fatalln(errInit)
	}

	// Make a new bucket called dev-minio.

	return minioClient, errInit
}

func CreateBucket(bucketName string) bool {
	ctx := context.Background()
	location := "us-east-1"

	exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
	if exists {
		return true
	}
	err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		if errBucketExists == nil {
			log.Fatalln(err)
			return false
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}
	return true
}
