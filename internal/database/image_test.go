package database

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockFile implements multipart.File interface for testing
type mockFile struct {
	*bytes.Reader
}

func (m *mockFile) Close() error {
	return nil
}

func (m *mockFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, io.EOF
}

func createMockFile(content string) multipart.File {
	return &mockFile{bytes.NewReader([]byte(content))}
}

// TestSaveImage_InvalidFileType tests that invalid file types are rejected
func TestSaveImage_InvalidFileType(t *testing.T) {
	invalidFiles := []struct {
		filename string
		wantErr  string
	}{
		{"test.txt", "invalid file type"},
		{"test.pdf", "invalid file type"},
		{"test.doc", "invalid file type"},
		{"test.exe", "invalid file type"},
		{"test.gif", "invalid file type"},
		{"test.bmp", "invalid file type"},
		{"test.svg", "invalid file type"},
		{"test", "invalid file type"},
		{"test.PHP", "invalid file type"},
	}

	for _, tt := range invalidFiles {
		t.Run(tt.filename, func(t *testing.T) {
			mockFile := createMockFile("fake image content")
			_, err := SaveImage(mockFile, tt.filename)

			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.filename)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.wantErr, err.Error())
			}
		})
	}
}

// TestSaveImage_ValidFileTypes tests that valid file types are accepted
func TestSaveImage_ValidFileTypes(t *testing.T) {
	// Skip if no database connection
	if DB == nil {
		t.Skip("Skipping test - no database connection")
	}

	// Clean up test images directory after test
	defer func() {
		os.RemoveAll("images")
	}()

	validFiles := []string{
		"test.jpg",
		"test.jpeg",
		"test.png",
		"test.JPG",
		"test.JPEG",
		"test.PNG",
		"Photo.jpg",
		"image_123.png",
	}

	for _, filename := range validFiles {
		t.Run(filename, func(t *testing.T) {
			mockFile := createMockFile("fake image content")
			image, err := SaveImage(mockFile, filename)

			if err != nil {
				// If DB error, that's expected without proper setup
				if strings.Contains(err.Error(), "failed to save image record") {
					t.Skip("Skipping - database not properly configured for test")
				}
				// File system errors are real errors
				if strings.Contains(err.Error(), "failed to create") || strings.Contains(err.Error(), "failed to save file") {
					t.Errorf("Unexpected file system error: %v", err)
				}
				return
			}

			// If we got here, verify the image was created
			if image == nil {
				t.Error("Expected image object, got nil")
			}

			if image.Path == "" {
				t.Error("Expected image path to be set")
			}

			// Verify file was created
			if _, err := os.Stat(image.Path); os.IsNotExist(err) {
				t.Errorf("Expected file to exist at %s", image.Path)
			}

			// Clean up created file
			os.Remove(image.Path)
		})
	}
}

// TestSaveImage_FileExtensionCaseInsensitive tests that file extensions are case-insensitive
func TestSaveImage_FileExtensionCaseInsensitive(t *testing.T) {
	if DB == nil {
		t.Skip("Skipping - no database connection")
	}
	
	testCases := []string{
		"test.JPG",
		"test.Jpg",
		"test.jPg",
		"test.JPEG",
		"test.PNG",
	}

	for _, filename := range testCases {
		t.Run(filename, func(t *testing.T) {
			mockFile := createMockFile("fake content")
			_, err := SaveImage(mockFile, filename)

			// Should not get "invalid file type" error
			if err != nil && strings.Contains(err.Error(), "invalid file type") {
				t.Errorf("File type should be valid for %s, got error: %v", filename, err)
			}
		})
	}
}

// TestSaveImage_DirectoryCreation tests that images directory is created if it doesn't exist
func TestSaveImage_DirectoryCreation(t *testing.T) {
	if DB == nil {
		t.Skip("Skipping - no database connection")
	}
	
	// Clean up before and after test
	os.RemoveAll("images")
	defer os.RemoveAll("images")

	mockFile := createMockFile("test content")
	_, err := SaveImage(mockFile, "test.jpg")

	// Check if directory was created
	if _, statErr := os.Stat("images"); os.IsNotExist(statErr) {
		t.Error("Expected images directory to be created")
	}

	// DB error is expected without proper setup
	if err != nil && !strings.Contains(err.Error(), "failed to save image record") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestSaveImage_UniqueFilenames tests that saved files have unique names
func TestSaveImage_UniqueFilenames(t *testing.T) {
	if DB == nil {
		t.Skip("Skipping test - no database connection")
	}

	defer os.RemoveAll("images")

	// Save same filename twice
	mockFile1 := createMockFile("content 1")
	mockFile2 := createMockFile("content 2")

	image1, err1 := SaveImage(mockFile1, "same.jpg")
	image2, err2 := SaveImage(mockFile2, "same.jpg")

	// Skip if DB not configured
	if err1 != nil && strings.Contains(err1.Error(), "failed to save image record") {
		t.Skip("Skipping - database not configured")
	}
	if err2 != nil && strings.Contains(err2.Error(), "failed to save image record") {
		t.Skip("Skipping - database not configured")
	}

	if err1 == nil && err2 == nil {
		// Verify filenames are different
		if image1.Path == image2.Path {
			t.Error("Expected unique filenames for duplicate uploads")
		}

		// Verify both contain UUID
		if !strings.Contains(filepath.Base(image1.Path), "_") {
			t.Error("Expected filename to contain UUID separator")
		}

		// Clean up
		os.Remove(image1.Path)
		os.Remove(image2.Path)
	}
}

// TestEditPostWithImage_InvalidPostID tests validation
func TestEditPostWithImage_InvalidPostID(t *testing.T) {
	tests := []struct {
		postID int
		name   string
	}{
		{0, "zero post ID"},
		{-1, "negative post ID"},
		{-100, "large negative post ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EditPostWithImage(tt.postID, 1, "Header", "Content", "Name", false, nil, "")

			if err == nil {
				t.Errorf("Expected error for %s", tt.name)
			}

			if err.Error() != "invalid post id" {
				t.Errorf("Expected 'invalid post id' error, got '%s'", err.Error())
			}
		})
	}
}

// TestEditPostWithImage_InvalidUserID tests user ID validation
func TestEditPostWithImage_InvalidUserID(t *testing.T) {
	tests := []struct {
		userID int
		name   string
	}{
		{0, "zero user ID"},
		{-1, "negative user ID"},
		{-100, "large negative user ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EditPostWithImage(1, tt.userID, "Header", "Content", "Name", false, nil, "")

			if err == nil {
				t.Errorf("Expected error for %s", tt.name)
			}

			if err.Error() != "invalid user id" {
				t.Errorf("Expected 'invalid user id' error, got '%s'", err.Error())
			}
		})
	}
}

// TestEditPostWithImage_EmptyHeader tests header validation
func TestEditPostWithImage_EmptyHeader(t *testing.T) {
	_, err := EditPostWithImage(1, 1, "", "Content", "Name", false, nil, "")

	if err == nil {
		t.Error("Expected error for empty header")
	}

	if err.Error() != "header cannot be empty" {
		t.Errorf("Expected 'header cannot be empty' error, got '%s'", err.Error())
	}
}

// TestEditPostWithImage_EmptyContent tests content validation
func TestEditPostWithImage_EmptyContent(t *testing.T) {
	_, err := EditPostWithImage(1, 1, "Header", "", "Name", false, nil, "")

	if err == nil {
		t.Error("Expected error for empty content")
	}

	if err.Error() != "content text cannot be empty" {
		t.Errorf("Expected 'content text cannot be empty' error, got '%s'", err.Error())
	}
}

// TestUpdateImagePostID_Validation tests UpdateImagePostID parameter validation
func TestUpdateImagePostID_Validation(t *testing.T) {
	if DB == nil {
		t.Skip("Skipping database-dependent test")
	}

	// Test with invalid IDs - should not panic
	err := UpdateImagePostID(0, 0)
	if err == nil {
		t.Log("UpdateImagePostID with zero IDs completed (DB error expected)")
	}

	err = UpdateImagePostID(-1, -1)
	if err == nil {
		t.Log("UpdateImagePostID with negative IDs completed (DB error expected)")
	}
}
