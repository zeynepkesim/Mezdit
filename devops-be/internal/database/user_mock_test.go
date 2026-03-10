package database

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"
)

// TestLoginUser_Success_Mock tests successful login with mock database
func TestLoginUser_Success_Mock(t *testing.T) {
	// Create mock database
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	// Save original DB and replace with mock
	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	// Generate bcrypt hash for test password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Expect the query
	rows := sqlmock.NewRows([]string{"id", "name", "username", "password", "profile_photo", "created_at"}).
		AddRow(1, "Test User", "testuser", string(hashedPassword), "https://example.com/photo.jpg", time.Now())

	mock.ExpectQuery(`SELECT id, name, username, password, profile_photo, created_at FROM users WHERE username = \$1`).
		WithArgs("testuser").
		WillReturnRows(rows)

	// Test login
	user, err := LoginUser("testuser", "password123")

	if err != nil {
		t.Errorf("Expected successful login, got error: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be returned")
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got '%s'", user.Name)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestLoginUser_WrongPassword_Mock tests login with wrong password
func TestLoginUser_WrongPassword_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	// Hash for "correctpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	rows := sqlmock.NewRows([]string{"id", "name", "username", "password", "profile_photo", "created_at"}).
		AddRow(1, "Test User", "testuser", string(hashedPassword), nil, time.Now())

	mock.ExpectQuery(`SELECT id, name, username, password, profile_photo, created_at FROM users WHERE username = \$1`).
		WithArgs("testuser").
		WillReturnRows(rows)

	// Try login with wrong password
	user, err := LoginUser("testuser", "wrongpassword")

	if err == nil {
		t.Error("Expected error for wrong password")
	}

	if user != nil {
		t.Error("Expected nil user for wrong password")
	}

	if err != nil && err.Error() != "invalid username or password" {
		t.Errorf("Expected 'invalid username or password' error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestLoginUser_UserNotFound_Mock tests login with non-existent user
func TestLoginUser_UserNotFound_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectQuery(`SELECT id, name, username, password, profile_photo, created_at FROM users WHERE username = \$1`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	user, err := LoginUser("nonexistent", "password123")

	if err == nil {
		t.Error("Expected error for non-existent user")
	}

	if user != nil {
		t.Error("Expected nil user for non-existent user")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestCreateUser_Success_Mock tests successful user creation
func TestCreateUser_Success_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	// Mock GetUserByUsername to return user not found
	mock.ExpectQuery(`SELECT id, name, username, password, profile_photo, created_at FROM users WHERE username = \$1`).
		WithArgs("newuser").
		WillReturnError(sql.ErrNoRows)

	// Mock INSERT query
	insertRows := sqlmock.NewRows([]string{"id", "name", "username", "password", "profile_photo", "created_at"}).
		AddRow(5, "New User", "newuser", "hashedpass", nil, time.Now())

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("New User", "newuser", sqlmock.AnyArg()).
		WillReturnRows(insertRows)

	// Mock UPDATE query for profile photo
	mock.ExpectExec(`UPDATE users SET profile_photo = \$1 WHERE id = \$2`).
		WithArgs(sqlmock.AnyArg(), 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	user, err := CreateUser("New User", "newuser", "password123")

	if err != nil {
		t.Errorf("Expected successful creation, got error: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be returned")
	}

	if user.Username != "newuser" {
		t.Errorf("Expected username 'newuser', got '%s'", user.Username)
	}

	if user.ID != 5 {
		t.Errorf("Expected user ID 5, got %d", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestCreateUser_DuplicateUsername_Mock tests creating user with existing username
func TestCreateUser_DuplicateUsername_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	// Mock GetUserByUsername to return existing user
	existingUserRows := sqlmock.NewRows([]string{"id", "name", "username", "password", "profile_photo", "created_at"}).
		AddRow(1, "Existing User", "existinguser", "hashedpass", nil, time.Now())

	mock.ExpectQuery(`SELECT id, name, username, password, profile_photo, created_at FROM users WHERE username = \$1`).
		WithArgs("existinguser").
		WillReturnRows(existingUserRows)

	user, err := CreateUser("New User", "existinguser", "password123")

	if err == nil {
		t.Error("Expected error for duplicate username")
	}

	if user != nil {
		t.Error("Expected nil user for duplicate username")
	}

	if err != nil && err.Error() != "username already exists" {
		t.Errorf("Expected 'username already exists' error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetUserByUsername_Found_Mock tests getting user by username
func TestGetUserByUsername_Found_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	rows := sqlmock.NewRows([]string{"id", "name", "username", "password", "profile_photo", "created_at"}).
		AddRow(10, "John Doe", "johndoe", "hashedpassword", "https://example.com/photo.jpg", time.Now())

	mock.ExpectQuery(`SELECT id, name, username, password, profile_photo, created_at FROM users WHERE username = \$1`).
		WithArgs("johndoe").
		WillReturnRows(rows)

	user, err := GetUserByUsername("johndoe")

	if err != nil {
		t.Errorf("Expected successful query, got error: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be returned")
	}

	if user.Username != "johndoe" {
		t.Errorf("Expected username 'johndoe', got '%s'", user.Username)
	}

	if user.ID != 10 {
		t.Errorf("Expected user ID 10, got %d", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetUserByUsername_NotFound_Mock tests getting non-existent user
func TestGetUserByUsername_NotFound_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectQuery(`SELECT id, name, username, password, profile_photo, created_at FROM users WHERE username = \$1`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	user, err := GetUserByUsername("nonexistent")

	if err == nil {
		t.Error("Expected error for non-existent user")
	}

	if user != nil {
		t.Error("Expected nil user for non-existent user")
	}

	if err != nil && err.Error() != "user not found" {
		t.Errorf("Expected 'user not found' error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetUserByID_Found_Mock tests getting user by ID
func TestGetUserByID_Found_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	rows := sqlmock.NewRows([]string{"id", "name", "username", "password", "profile_photo", "created_at"}).
		AddRow(15, "Jane Smith", "janesmith", "hashedpassword", nil, time.Now())

	mock.ExpectQuery(`SELECT id, name, username, password, profile_photo, created_at FROM users WHERE id = \$1`).
		WithArgs(15).
		WillReturnRows(rows)

	user, err := GetUserByID(15)

	if err != nil {
		t.Errorf("Expected successful query, got error: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user to be returned")
	}

	if user.ID != 15 {
		t.Errorf("Expected user ID 15, got %d", user.ID)
	}

	if user.Name != "Jane Smith" {
		t.Errorf("Expected name 'Jane Smith', got '%s'", user.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestGetUserByID_NotFound_Mock tests getting non-existent user by ID
func TestGetUserByID_NotFound_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	mock.ExpectQuery(`SELECT id, name, username, password, profile_photo, created_at FROM users WHERE id = \$1`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	user, err := GetUserByID(999)

	if err == nil {
		t.Error("Expected error for non-existent user")
	}

	if user != nil {
		t.Error("Expected nil user for non-existent user")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestChangePassword_Success_Mock tests successful password change
func TestChangePassword_Success_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	// Current password hash
	currentPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)

	// Mock GetUserByID
	getUserRows := sqlmock.NewRows([]string{"id", "name", "username", "password", "profile_photo", "created_at"}).
		AddRow(5, "Test User", "testuser", string(currentPasswordHash), nil, time.Now())

	mock.ExpectQuery(`SELECT id, name, username, password, profile_photo, created_at FROM users WHERE id = \$1`).
		WithArgs(5).
		WillReturnRows(getUserRows)

	// Mock UPDATE password
	mock.ExpectExec(`UPDATE users SET password = \$1 WHERE id = \$2`).
		WithArgs(sqlmock.AnyArg(), 5).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = ChangePassword(5, "oldpassword", "newpassword123")

	if err != nil {
		t.Errorf("Expected successful password change, got error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// TestChangePassword_WrongCurrentPassword_Mock tests changing password with wrong current password
func TestChangePassword_WrongCurrentPassword_Mock(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	originalDB := DB
	DB = mockDB
	defer func() { DB = originalDB }()

	// Hash for "correctoldpassword"
	currentPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("correctoldpassword"), bcrypt.DefaultCost)

	getUserRows := sqlmock.NewRows([]string{"id", "name", "username", "password", "profile_photo", "created_at"}).
		AddRow(5, "Test User", "testuser", string(currentPasswordHash), nil, time.Now())

	mock.ExpectQuery(`SELECT id, name, username, password, profile_photo, created_at FROM users WHERE id = \$1`).
		WithArgs(5).
		WillReturnRows(getUserRows)

	err = ChangePassword(5, "wrongoldpassword", "newpassword123")

	if err == nil {
		t.Error("Expected error for wrong current password")
	}

	if err != nil && err.Error() != "current password is incorrect" {
		t.Errorf("Expected 'current password is incorrect' error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
