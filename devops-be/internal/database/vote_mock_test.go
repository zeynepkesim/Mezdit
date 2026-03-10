package database

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestVotePost_NewVote_Upvote_Mock tests adding new upvote
func TestVotePost_NewVote_Upvote_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectBegin()

	// Check post exists
	postExistsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM posts WHERE id = \$1\)`).
		WithArgs(10).
		WillReturnRows(postExistsRows)

	// Check user exists
	userExistsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id = \$1\)`).
		WithArgs(5).
		WillReturnRows(userExistsRows)

	// Check previous vote (none)
	mock.ExpectQuery(`SELECT upvoted FROM user_votes WHERE post_id = \$1 AND user_id = \$2`).
		WithArgs(10, 5).
		WillReturnError(sql.ErrNoRows)

	// Insert new vote
	mock.ExpectExec(`INSERT INTO user_votes \(post_id, user_id, upvoted\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(10, 5, true).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Update vote count
	mock.ExpectExec(`UPDATE posts SET vote_count = vote_count \+ \$1 WHERE id = \$2`).
		WithArgs(1, 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Get updated post
	updatedPostRows := sqlmock.NewRows([]string{
		"id", "publisher_id", "publisher_name", "header", "content_text",
		"content_photo_id", "is_private", "vote_count", "created_at",
		"publisher_photo", "content_photo_path", "is_liked",
	}).AddRow(10, 1, "Test User", "Test Post", "Content", nil, false, 1, time.Now(), nil, nil, true)

	mock.ExpectQuery(`SELECT p.id, p.publisher_id`).
		WithArgs(10, 5).
		WillReturnRows(updatedPostRows)

	mock.ExpectCommit()

	post, err := VotePost(10, 5, true)

	if err != nil {
		t.Errorf("Expected successful vote, got error: %v", err)
	}

	if post == nil {
		t.Fatal("Expected post to be returned")
	}

	if post.VoteCount != 1 {
		t.Errorf("Expected vote count 1, got %d", post.VoteCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestVotePost_NewVote_Downvote_Mock tests adding new downvote
func TestVotePost_NewVote_Downvote_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectBegin()

	postExistsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM posts WHERE id = \$1\)`).
		WithArgs(10).
		WillReturnRows(postExistsRows)

	userExistsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id = \$1\)`).
		WithArgs(5).
		WillReturnRows(userExistsRows)

	mock.ExpectQuery(`SELECT upvoted FROM user_votes WHERE post_id = \$1 AND user_id = \$2`).
		WithArgs(10, 5).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO user_votes \(post_id, user_id, upvoted\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(10, 5, false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE posts SET vote_count = vote_count \+ \$1 WHERE id = \$2`).
		WithArgs(-1, 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updatedPostRows := sqlmock.NewRows([]string{
		"id", "publisher_id", "publisher_name", "header", "content_text",
		"content_photo_id", "is_private", "vote_count", "created_at",
		"publisher_photo", "content_photo_path", "is_liked",
	}).AddRow(10, 1, "Test User", "Test Post", "Content", nil, false, -1, time.Now(), nil, nil, false)

	mock.ExpectQuery(`SELECT p.id, p.publisher_id`).
		WithArgs(10, 5).
		WillReturnRows(updatedPostRows)

	mock.ExpectCommit()

	post, err := VotePost(10, 5, false)

	if err != nil {
		t.Errorf("Expected successful vote, got error: %v", err)
	}

	if post == nil {
		t.Fatal("Expected post to be returned")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestVotePost_ChangeVote_Mock tests changing vote from upvote to downvote
func TestVotePost_ChangeVote_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectBegin()

	postExistsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM posts WHERE id = \$1\)`).
		WithArgs(10).
		WillReturnRows(postExistsRows)

	userExistsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id = \$1\)`).
		WithArgs(5).
		WillReturnRows(userExistsRows)

	// User previously upvoted
	previousVoteRows := sqlmock.NewRows([]string{"upvoted"}).AddRow(true)
	mock.ExpectQuery(`SELECT upvoted FROM user_votes WHERE post_id = \$1 AND user_id = \$2`).
		WithArgs(10, 5).
		WillReturnRows(previousVoteRows)

	// Update vote to downvote
	mock.ExpectExec(`UPDATE user_votes SET upvoted = \$1 WHERE post_id = \$2 AND user_id = \$3`).
		WithArgs(false, 10, 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Update vote count (-2: remove upvote, add downvote)
	mock.ExpectExec(`UPDATE posts SET vote_count = vote_count \+ \$1 WHERE id = \$2`).
		WithArgs(-2, 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updatedPostRows := sqlmock.NewRows([]string{
		"id", "publisher_id", "publisher_name", "header", "content_text",
		"content_photo_id", "is_private", "vote_count", "created_at",
		"publisher_photo", "content_photo_path", "is_liked",
	}).AddRow(10, 1, "Test User", "Test Post", "Content", nil, false, -1, time.Now(), nil, nil, false)

	mock.ExpectQuery(`SELECT p.id, p.publisher_id`).
		WithArgs(10, 5).
		WillReturnRows(updatedPostRows)

	mock.ExpectCommit()

	post, err := VotePost(10, 5, false)

	if err != nil {
		t.Errorf("Expected successful vote change, got error: %v", err)
	}

	if post == nil {
		t.Fatal("Expected post to be returned")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestVotePost_ToggleOff_Mock tests removing vote
func TestVotePost_ToggleOff_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectBegin()

	postExistsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM posts WHERE id = \$1\)`).
		WithArgs(10).
		WillReturnRows(postExistsRows)

	userExistsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id = \$1\)`).
		WithArgs(5).
		WillReturnRows(userExistsRows)

	// User previously upvoted, clicking upvote again (toggle off)
	previousVoteRows := sqlmock.NewRows([]string{"upvoted"}).AddRow(true)
	mock.ExpectQuery(`SELECT upvoted FROM user_votes WHERE post_id = \$1 AND user_id = \$2`).
		WithArgs(10, 5).
		WillReturnRows(previousVoteRows)

	// Delete vote
	mock.ExpectExec(`DELETE FROM user_votes WHERE post_id = \$1 AND user_id = \$2`).
		WithArgs(10, 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Update vote count (-1: remove upvote)
	mock.ExpectExec(`UPDATE posts SET vote_count = vote_count \+ \$1 WHERE id = \$2`).
		WithArgs(-1, 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updatedPostRows := sqlmock.NewRows([]string{
		"id", "publisher_id", "publisher_name", "header", "content_text",
		"content_photo_id", "is_private", "vote_count", "created_at",
		"publisher_photo", "content_photo_path", "is_liked",
	}).AddRow(10, 1, "Test User", "Test Post", "Content", nil, false, 0, time.Now(), nil, nil, nil)

	mock.ExpectQuery(`SELECT p.id, p.publisher_id`).
		WithArgs(10, 5).
		WillReturnRows(updatedPostRows)

	mock.ExpectCommit()

	post, err := VotePost(10, 5, true)

	if err != nil {
		t.Errorf("Expected successful vote toggle, got error: %v", err)
	}

	if post == nil {
		t.Fatal("Expected post to be returned")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestVotePost_PostNotFound_Mock tests voting on non-existent post
func TestVotePost_PostNotFound_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectBegin()

	postExistsRows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM posts WHERE id = \$1\)`).
		WithArgs(999).
		WillReturnRows(postExistsRows)

	mock.ExpectRollback()

	post, err := VotePost(999, 5, true)

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

// TestVotePost_UserNotFound_Mock tests voting with non-existent user
func TestVotePost_UserNotFound_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectBegin()

	postExistsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM posts WHERE id = \$1\)`).
		WithArgs(10).
		WillReturnRows(postExistsRows)

	userExistsRows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE id = \$1\)`).
		WithArgs(999).
		WillReturnRows(userExistsRows)

	mock.ExpectRollback()

	post, err := VotePost(10, 999, true)

	if err == nil {
		t.Error("Expected error for non-existent user")
	}

	if post != nil {
		t.Error("Expected nil post for non-existent user")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
