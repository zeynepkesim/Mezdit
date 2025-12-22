package storage

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"testing"
	"time"
)

// mockFile implements multipart.File interface for testing
type mockFile struct {
	*bytes.Reader
}

func (m *mockFile) Close() error {
	return nil
}

func newMockFile(content string) multipart.File {
	return &mockFile{Reader: bytes.NewReader([]byte(content))}
}

func TestUploadImage_InvalidFileType(t *testing.T) {
	// Test invalid file extensions
	testCases := []struct {
		filename string
		wantErr  bool
	}{
		{"test.txt", true},
		{"test.pdf", true},
		{"test.gif", true},
		{"test.bmp", true},
		{"test", true},
		{"test.JPG", false},  // uppercase should work
		{"test.JPEG", false}, // uppercase should work
		{"test.PNG", false},  // uppercase should work
		{"test.jpg", false},
		{"test.jpeg", false},
		{"test.png", false},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			mockFile := newMockFile("test content")
			_, err := UploadImage(mockFile, tc.filename)

			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tc.filename)
				}
				if !strings.Contains(err.Error(), "invalid file type") {
					t.Errorf("Expected 'invalid file type' error, got: %v", err)
				}
			} else {
				// For valid file types, we expect MinIO client error since it's not initialized
				if err == nil {
					t.Errorf("Expected MinIO client error for %s, got nil", tc.filename)
				}
				if !strings.Contains(err.Error(), "MinIO client not initialized") {
					t.Errorf("Expected 'MinIO client not initialized' error, got: %v", err)
				}
			}
		})
	}
}

func TestUploadImage_MinIONotInitialized(t *testing.T) {
	// Ensure minioClient is nil
	minioClient = nil

	mockFile := newMockFile("test image content")
	_, err := UploadImage(mockFile, "test.jpg")

	if err == nil {
		t.Fatal("Expected error when MinIO client is not initialized")
	}

	if !strings.Contains(err.Error(), "MinIO client not initialized") {
		t.Errorf("Expected 'MinIO client not initialized' error, got: %v", err)
	}
}

func TestDeleteImage_MinIONotInitialized(t *testing.T) {
	// Ensure minioClient is nil
	minioClient = nil

	err := DeleteImage("test-image.jpg")

	if err == nil {
		t.Fatal("Expected error when MinIO client is not initialized")
	}

	if !strings.Contains(err.Error(), "MinIO client not initialized") {
		t.Errorf("Expected 'MinIO client not initialized' error, got: %v", err)
	}
}

func TestGetImageURL_MinIONotInitialized(t *testing.T) {
	// Ensure minioClient is nil
	minioClient = nil

	url, err := GetImageURL("test-image.jpg", 15*time.Minute)

	if err == nil {
		t.Fatal("Expected error when MinIO client is not initialized")
	}

	if url != "" {
		t.Errorf("Expected empty URL, got: %s", url)
	}

	if !strings.Contains(err.Error(), "MinIO client not initialized") {
		t.Errorf("Expected 'MinIO client not initialized' error, got: %v", err)
	}
}

func TestGetImageURL_EmptyObjectName(t *testing.T) {
	// Ensure minioClient is nil
	minioClient = nil

	url, err := GetImageURL("", 15*time.Minute)

	// When object name is empty, function should check MinIO client first
	// and return error if not initialized
	if err == nil {
		t.Fatal("Expected error when MinIO client is not initialized")
	}

	if !strings.Contains(err.Error(), "MinIO client not initialized") {
		t.Errorf("Expected 'MinIO client not initialized' error, got: %v", err)
	}

	if url != "" {
		t.Errorf("Expected empty URL when error occurs, got: %s", url)
	}
}

func TestInitMinIO_MissingEnvVars(t *testing.T) {
	// Save original env vars
	origEndpoint := os.Getenv("MINIO_ENDPOINT")
	origAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	origSecretKey := os.Getenv("MINIO_SECRET_KEY")
	origBucketName := os.Getenv("MINIO_BUCKET_NAME")

	// Defer restore
	defer func() {
		os.Setenv("MINIO_ENDPOINT", origEndpoint)
		os.Setenv("MINIO_ACCESS_KEY", origAccessKey)
		os.Setenv("MINIO_SECRET_KEY", origSecretKey)
		os.Setenv("MINIO_BUCKET_NAME", origBucketName)
	}()

	testCases := []struct {
		name        string
		endpoint    string
		accessKey   string
		secretKey   string
		bucketName  string
		expectError bool
	}{
		{
			name:        "all missing",
			endpoint:    "",
			accessKey:   "",
			secretKey:   "",
			bucketName:  "",
			expectError: true,
		},
		{
			name:        "missing endpoint",
			endpoint:    "",
			accessKey:   "access",
			secretKey:   "secret",
			bucketName:  "bucket",
			expectError: true,
		},
		{
			name:        "missing access key",
			endpoint:    "localhost:9000",
			accessKey:   "",
			secretKey:   "secret",
			bucketName:  "bucket",
			expectError: true,
		},
		{
			name:        "missing secret key",
			endpoint:    "localhost:9000",
			accessKey:   "access",
			secretKey:   "",
			bucketName:  "bucket",
			expectError: true,
		},
		{
			name:        "missing bucket name",
			endpoint:    "localhost:9000",
			accessKey:   "access",
			secretKey:   "secret",
			bucketName:  "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set test env vars
			os.Setenv("MINIO_ENDPOINT", tc.endpoint)
			os.Setenv("MINIO_ACCESS_KEY", tc.accessKey)
			os.Setenv("MINIO_SECRET_KEY", tc.secretKey)
			os.Setenv("MINIO_BUCKET_NAME", tc.bucketName)
			os.Setenv("MINIO_USE_SSL", "false")

			err := InitMinIO()

			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
			}

			if tc.expectError && err != nil {
				if !strings.Contains(err.Error(), "MinIO environment variables not set") {
					t.Errorf("Expected 'MinIO environment variables not set' error, got: %v", err)
				}
			}
		})
	}
}

func TestUploadImage_FileExtensionCaseInsensitive(t *testing.T) {
	testCases := []struct {
		filename    string
		shouldBeJPG bool
	}{
		{"test.jpg", true},
		{"test.JPG", true},
		{"test.JpG", true},
		{"test.jpeg", true},
		{"test.JPEG", true},
		{"test.JpEg", true},
		{"test.png", true},
		{"test.PNG", true},
		{"test.PnG", true},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			mockFile := newMockFile("test content")
			_, err := UploadImage(mockFile, tc.filename)

			// We expect MinIO client error, not file type error
			if err == nil {
				t.Errorf("Expected MinIO client error for %s, got nil", tc.filename)
			}

			if strings.Contains(err.Error(), "invalid file type") {
				t.Errorf("File %s should be valid, but got invalid file type error", tc.filename)
			}

			if !strings.Contains(err.Error(), "MinIO client not initialized") {
				t.Errorf("Expected 'MinIO client not initialized' error for %s, got: %v", tc.filename, err)
			}
		})
	}
}

func TestUploadImage_ReadFileError(t *testing.T) {
	// Create a file that will error on read
	errorReader := &errorReadCloser{}
	mockFile := &mockFileWithError{Reader: errorReader}

	_, err := UploadImage(mockFile, "test.jpg")

	if err == nil {
		t.Fatal("Expected error when reading file fails")
	}

	// Should get "MinIO client not initialized" error first since we check that before reading
	if !strings.Contains(err.Error(), "MinIO client not initialized") {
		t.Errorf("Expected 'MinIO client not initialized' error, got: %v", err)
	}
}

// errorReadCloser implements io.Reader that always returns an error
type errorReadCloser struct{}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (e *errorReadCloser) Close() error {
	return nil
}

// mockFileWithError implements multipart.File with an error reader
type mockFileWithError struct {
	Reader io.Reader
}

func (m *mockFileWithError) Read(p []byte) (n int, err error) {
	return m.Reader.Read(p)
}

func (m *mockFileWithError) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (m *mockFileWithError) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (m *mockFileWithError) Close() error {
	return nil
}
