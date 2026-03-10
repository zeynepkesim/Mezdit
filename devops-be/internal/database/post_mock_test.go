package database

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestGetPosts_Mock tests getting all public posts with mock
func TestGetPosts_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	rows := sqlmock.NewRows([]string{
		"id", "publisher_id", "publisher_name", "header", "content_text",
		"content_photo_id", "is_private", "vote_count", "created_at",
		"publisher_photo", "content_photo_path", "is_liked",
	}).
		AddRow(1, 5, "User One", "Public Post", "Public content", nil, false, 10, time.Now(), "photo.jpg", nil, nil).
		AddRow(2, 6, "User Two", "Another Post", "More content", nil, false, 5, time.Now(), nil, nil, nil)

	mock.ExpectQuery(`SELECT (.+) FROM posts (.+) WHERE p.is_private = FALSE`).
		WillReturnRows(rows)

	posts, err := GetPosts()

	if err != nil {
		t.Errorf("Expected successful query, got error: %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(posts))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetPostsForUser_WithUser_Mock tests getting posts for authenticated user
func TestGetPostsForUser_WithUser_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	rows := sqlmock.NewRows([]string{
		"id", "publisher_id", "publisher_name", "header", "content_text",
		"content_photo_id", "is_private", "vote_count", "created_at",
		"publisher_photo", "content_photo_path", "is_liked",
	}).
		AddRow(1, 5, "User One", "My Post", "Content 1", nil, false, 10, time.Now(), "photo.jpg", nil, true).
		AddRow(2, 5, "User One", "Private Post", "Content 2", nil, true, 0, time.Now(), "photo.jpg", nil, false)

	mock.ExpectQuery(`SELECT (.+) FROM posts p(.+)JOIN users u ON p.publisher_id = u.id(.+)LEFT JOIN images i ON p.content_photo_id = i.id(.+)LEFT JOIN user_votes uv ON p.id = uv.post_id AND uv.user_id = \$1(.+)WHERE \(p.is_private = FALSE OR p.publisher_id = \$1\)(.+)ORDER BY p.created_at DESC`).
		WithArgs(5).
		WillReturnRows(rows)

	posts, err := GetPostsForUser(5)

	if err != nil {
		t.Errorf("Expected successful query, got error: %v", err)
	}

	if len(posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(posts))
	}

	if posts[0].IsLiked == nil {
		t.Error("Expected IsLiked to be set")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetPostsForUser_NoUser_Mock tests getting posts without user context
func TestGetPostsForUser_NoUser_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	rows := sqlmock.NewRows([]string{
		"id", "publisher_id", "publisher_name", "header", "content_text",
		"content_photo_id", "is_private", "vote_count", "created_at",
		"publisher_photo", "content_photo_path", "is_liked",
	}).
		AddRow(1, 5, "User One", "Public Post", "Public content", nil, false, 10, time.Now(), nil, nil, nil)

	mock.ExpectQuery(`SELECT (.+) FROM posts (.+) WHERE p.is_private = FALSE`).
		WillReturnRows(rows)

	posts, err := GetPostsForUser(0)

	if err != nil {
		t.Errorf("Expected successful query, got error: %v", err)
	}

	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetPostsForUser_Empty_Mock tests empty result
func TestGetPostsForUser_Empty_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	rows := sqlmock.NewRows([]string{
		"id", "publisher_id", "publisher_name", "header", "content_text",
		"content_photo_id", "is_private", "vote_count", "created_at",
		"publisher_photo", "content_photo_path", "is_liked",
	})

	mock.ExpectQuery(`SELECT (.+) FROM posts`).
		WillReturnRows(rows)

	posts, err := GetPostsForUser(0)

	if err != nil {
		t.Errorf("Expected successful query, got error: %v", err)
	}

	if len(posts) != 0 {
		t.Errorf("Expected 0 posts, got %d", len(posts))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestEditPost_Success_Mock tests successful post editing
func TestEditPost_Success_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectBegin()

	publisherRows := sqlmock.NewRows([]string{"publisher_id"}).AddRow(5)
	mock.ExpectQuery(`SELECT publisher_id FROM posts WHERE id = \$1`).
		WithArgs(10).
		WillReturnRows(publisherRows)

	userExistsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id = \$1\)`).
		WithArgs(5).
		WillReturnRows(userExistsRows)

	mock.ExpectExec(`UPDATE posts SET header = \$1, content_text = \$2, publisher_name = \$3, is_private = \$4 WHERE id = \$5`).
		WithArgs("Updated Header", "Updated content", "Test User", false, 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updatedPostRows := sqlmock.NewRows([]string{
		"id", "publisher_id", "publisher_name", "header", "content_text",
		"content_photo_id", "is_private", "vote_count", "created_at",
		"publisher_photo", "content_photo_path", "is_liked",
	}).AddRow(10, 5, "Test User", "Updated Header", "Updated content", nil, false, 5, time.Now(), nil, nil, false)

	mock.ExpectQuery(`SELECT p.id, p.publisher_id`).
		WithArgs(10, 5).
		WillReturnRows(updatedPostRows)

	mock.ExpectCommit()

	post, err := EditPost(10, 5, "Updated Header", "Updated content", "Test User", false)

	if err != nil {
		t.Errorf("Expected successful edit, got error: %v", err)
	}

	if post == nil {
		t.Fatal("Expected post to be returned")
	}

	if post.Header != "Updated Header" {
		t.Errorf("Expected header 'Updated Header', got '%s'", post.Header)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestEditPost_NotOwner_Mock tests editing post by non-owner
func TestEditPost_NotOwner_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectBegin()

	publisherRows := sqlmock.NewRows([]string{"publisher_id"}).AddRow(5)
	mock.ExpectQuery(`SELECT publisher_id FROM posts WHERE id = \$1`).
		WithArgs(10).
		WillReturnRows(publisherRows)

	mock.ExpectRollback()

	post, err := EditPost(10, 10, "Updated Header", "Updated content", "Test User", false)

	if err == nil {
		t.Error("Expected error for non-owner edit")
	}

	if post != nil {
		t.Error("Expected nil post for non-owner edit")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestEditPost_PostNotFound_Mock tests editing non-existent post
func TestEditPost_PostNotFound_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT publisher_id FROM posts WHERE id = \$1`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectRollback()

	post, err := EditPost(999, 5, "Updated Header", "Updated content", "Test User", false)

	if err == nil {
		t.Error("Expected error for non-existent post")
	}

	if post != nil {
		t.Error("Expected nil post for non-existent post")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
