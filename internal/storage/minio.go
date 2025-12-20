package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client
var bucketName string

// InitMinIO initializes the MinIO client
func InitMinIO() error {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucketName = os.Getenv("MINIO_BUCKET_NAME")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	if endpoint == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		return fmt.Errorf("MinIO environment variables not set")
	}

	var err error
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize MinIO client: %v", err)
	}

	// Create bucket if it doesn't exist
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}
		log.Printf("Bucket %s created successfully", bucketName)
	} else {
		log.Printf("Bucket %s already exists", bucketName)
	}

	return nil
}

// UploadImage uploads an image to MinIO and returns the object path
func UploadImage(file multipart.File, filename string) (string, error) {
	// Validate file extension FIRST (before checking MinIO client)
	ext := filepath.Ext(filename)
	ext = strings.ToLower(ext)

	allowedTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
	}

	contentType, ok := allowedTypes[ext]
	if !ok {
		return "", fmt.Errorf("invalid file type. Only JPG, JPEG, and PNG files are allowed")
	}

	// Then check MinIO client
	if minioClient == nil {
		return "", fmt.Errorf("MinIO client not initialized - check MINIO environment variables")
	}

	// Generate unique filename with UUID
	imageUUID := uuid.New().String()
	objectName := fmt.Sprintf("images/%s_%s", imageUUID, filename)

	// Read file content into buffer
	buf := bytes.NewBuffer(nil)
	fileSize, err := io.Copy(buf, file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Upload to MinIO
	ctx := context.Background()
	_, err = minioClient.PutObject(ctx, bucketName, objectName, buf, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to MinIO: %v", err)
	}

	log.Printf("Successfully uploaded %s to MinIO", objectName)
	return objectName, nil
}

// DeleteImage deletes an image from MinIO
func DeleteImage(objectName string) error {
	if minioClient == nil {
		return fmt.Errorf("MinIO client not initialized")
	}

	ctx := context.Background()
	err := minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object from MinIO: %v", err)
	}

	log.Printf("Successfully deleted %s from MinIO", objectName)
	return nil
}

// GetImageURL returns a presigned URL for accessing the image
func GetImageURL(objectName string, expiry time.Duration) (string, error) {
	if minioClient == nil {
		return "", fmt.Errorf("MinIO client not initialized")
	}

	if objectName == "" {
		return "", nil
	}

	ctx := context.Background()
	presignedURL, err := minioClient.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return presignedURL.String(), nil
}
