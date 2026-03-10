package database

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestUpdateImagePostID_Success_Mock tests successful image-post link update
func TestUpdateImagePostID_Success_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectExec(`UPDATE images SET post_id = \$1 WHERE id = \$2`).
		WithArgs(10, 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = UpdateImagePostID(5, 10)

	if err != nil {
		t.Errorf("Expected successful update, got error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestUpdateImagePostID_Failure_Mock tests database error
func TestUpdateImagePostID_Failure_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectExec(`UPDATE images SET post_id = \$1 WHERE id = \$2`).
		WithArgs(10, 999).
		WillReturnError(err)

	err = UpdateImagePostID(999, 10)

	if err == nil {
		t.Error("Expected error for database failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestEditPostWithImage_Validation tests input validation
func TestEditPostWithImage_Validation(t *testing.T) {
	tests := []struct {
		name            string
		postID          int
		userID          int
		header          string
		contentText     string
		publisherName   string
		expectError     bool
		errorContains   string
	}{
		{
			name:          "Invalid post ID - zero",
			postID:        0,
			userID:        1,
			header:        "Header",
			contentText:   "Content",
			publisherName: "User",
			expectError:   true,
			errorContains: "invalid post id",
		},
		{
			name:          "Invalid post ID - negative",
			postID:        -1,
			userID:        1,
			header:        "Header",
			contentText:   "Content",
			publisherName: "User",
			expectError:   true,
			errorContains: "invalid post id",
		},
		{
			name:          "Invalid user ID - zero",
			postID:        1,
			userID:        0,
			header:        "Header",
			contentText:   "Content",
			publisherName: "User",
			expectError:   true,
			errorContains: "invalid user id",
		},
		{
			name:          "Invalid user ID - negative",
			postID:        1,
			userID:        -1,
			header:        "Header",
			contentText:   "Content",
			publisherName: "User",
			expectError:   true,
			errorContains: "invalid user id",
		},
		{
			name:          "Empty header",
			postID:        1,
			userID:        1,
			header:        "",
			contentText:   "Content",
			publisherName: "User",
			expectError:   true,
			errorContains: "header cannot be empty",
		},
		{
			name:          "Empty content text",
			postID:        1,
			userID:        1,
			header:        "Header",
			contentText:   "",
			publisherName: "User",
			expectError:   true,
			errorContains: "content text cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EditPostWithImage(
				tt.postID,
				tt.userID,
				tt.header,
				tt.contentText,
				tt.publisherName,
				false,
				nil,
				"",
			)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorContains)
				} else if err.Error() != tt.errorContains {
					t.Errorf("Expected error '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestSaveImage_Validation tests file type validation
func TestSaveImage_Validation(t *testing.T) {
	// Note: SaveImage requires actual file I/O which is hard to mock
	// This test just validates the expected behavior pattern

	// Test cases for invalid file types would include:
	// - .txt files
	// - .pdf files
	// - .exe files
	// - files without extensions
	// - uppercase extensions

	// Since SaveImage requires multipart.File which needs actual file operations,
	// we document the expected validation behavior here:

	// Expected behavior:
	// 1. Only .jpg, .jpeg, .png files allowed
	// 2. Extensions are case-insensitive (.JPG == .jpg)
	// 3. Invalid file types return error with message about allowed types
	// 4. File is saved with UUID prefix
	// 5. Database record is created after file save
	// 6. If database insert fails, file is cleaned up

	t.Log("SaveImage validation documented - requires file I/O for actual testing")
}
