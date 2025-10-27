package database

import (
	"testing"
)

// TestStringPtr tests the stringPtr helper function
func TestStringPtr(t *testing.T) {
	tests := []string{
		"test string",
		"",
		"https://example.com/image.jpg",
		"special characters: !@#$%^&*()",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			ptr := stringPtr(tt)

			if ptr == nil {
				t.Error("Expected non-nil pointer")
			}

			if *ptr != tt {
				t.Errorf("Expected value '%s', got '%s'", tt, *ptr)
			}

			// Test that we get a new pointer each time
			ptr2 := stringPtr(tt)
			if ptr == ptr2 {
				t.Error("Expected different pointer addresses")
			}
		})
	}
}

// TestStringPtr_Empty tests stringPtr with empty string
func TestStringPtr_Empty(t *testing.T) {
	ptr := stringPtr("")

	if ptr == nil {
		t.Error("Expected non-nil pointer even for empty string")
	}

	if *ptr != "" {
		t.Errorf("Expected empty string, got '%s'", *ptr)
	}
}

// TestStringPtr_UniquePointers tests that each call returns a unique pointer
func TestStringPtr_UniquePointers(t *testing.T) {
	value := "test"

	ptr1 := stringPtr(value)
	ptr2 := stringPtr(value)
	ptr3 := stringPtr(value)

	// All pointers should point to the same value
	if *ptr1 != value || *ptr2 != value || *ptr3 != value {
		t.Error("All pointers should point to the same value")
	}

	// But they should be different pointers
	if ptr1 == ptr2 || ptr2 == ptr3 || ptr1 == ptr3 {
		t.Error("Each call should return a unique pointer")
	}
}

// TestClose tests the Close function doesn't panic
func TestClose(t *testing.T) {
	// Test that Close doesn't panic when DB is nil
	originalDB := DB
	originalGormDB := GormDB

	DB = nil
	GormDB = nil

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Close() panicked: %v", r)
		}
	}()

	Close()

	// Restore original values
	DB = originalDB
	GormDB = originalGormDB
}

// TestInit_InvalidDatabaseURL tests Init with invalid database URL
func TestInit_InvalidDatabaseURL(t *testing.T) {
	// Save original values
	originalDB := DB
	originalGormDB := GormDB

	// Set invalid DATABASE_URL
	t.Setenv("DATABASE_URL", "invalid://url")

	// Should not panic, just log warnings
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Init() panicked with invalid URL: %v", r)
		}
	}()

	Init()

	// Restore original values
	DB = originalDB
	GormDB = originalGormDB
}

// TestInit_EmptyDatabaseURL tests Init uses default when DATABASE_URL is empty
func TestInit_EmptyDatabaseURL(t *testing.T) {
	// This test just ensures Init doesn't panic when DATABASE_URL is empty
	// The actual database connection may fail if PostgreSQL is not running locally

	// Save original values
	originalDB := DB
	originalGormDB := GormDB

	// Clear DATABASE_URL
	t.Setenv("DATABASE_URL", "")

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Init() panicked: %v", r)
		}
	}()

	Init()

	// Restore original values
	DB = originalDB
	GormDB = originalGormDB
}

// TestCreateTables tests createTables function doesn't panic
func TestCreateTables(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("createTables() panicked: %v", r)
		}
	}()

	createTables()
}

// TestDatabaseConstants tests that exported variables exist
func TestDatabaseConstants(t *testing.T) {
	// This test verifies that the exported variables are accessible
	// We don't check their values since they depend on initialization

	// Just access them to ensure they're exported
	_ = DB
	_ = GormDB

	// Test passes if we can access these variables without compile error
}

// TestInitLegacyDB_InvalidURL tests initLegacyDB with invalid URL
func TestInitLegacyDB_InvalidURL(t *testing.T) {
	err := initLegacyDB("invalid://url")

	if err == nil {
		t.Error("Expected error for invalid database URL")
	}
}

// TestInitLegacyDB_UnreachableHost tests initLegacyDB with unreachable host
func TestInitLegacyDB_UnreachableHost(t *testing.T) {
	// Use a URL that's syntactically valid but unreachable
	err := initLegacyDB("postgres://user:password@unreachable-host-12345.invalid:5432/testdb?sslmode=disable")

	// Should return an error (connection failure)
	if err == nil {
		t.Error("Expected error for unreachable database host")
	}

	// DB should be nil after failed connection
	if DB != nil {
		t.Error("Expected DB to be nil after connection failure")
	}
}

// TestInitGormDB_InvalidURL tests initGormDB with invalid URL
func TestInitGormDB_InvalidURL(t *testing.T) {
	err := initGormDB("invalid://url")

	if err == nil {
		t.Error("Expected error for invalid database URL")
	}
}

// TestInitGormDB_UnreachableHost tests initGormDB with unreachable host
func TestInitGormDB_UnreachableHost(t *testing.T) {
	// Use a URL that's syntactically valid but unreachable
	err := initGormDB("postgres://user:password@unreachable-host-12345.invalid:5432/testdb?sslmode=disable")

	// Should return an error (connection failure)
	if err == nil {
		t.Error("Expected error for unreachable database host")
	}
}

// TestAddForeignKeyConstraints_NoDatabase tests foreign key addition without database
func TestAddForeignKeyConstraints_NoDatabase(t *testing.T) {
	// Save original GormDB
	originalGormDB := GormDB

	// Set GormDB to nil
	GormDB = nil

	defer func() {
		// This function will panic with nil pointer, catch it
		if r := recover(); r == nil {
			t.Error("Expected panic when GormDB is nil")
		}
		// Restore original GormDB
		GormDB = originalGormDB
	}()

	// This will panic, which we expect
	_ = addForeignKeyConstraints()
}

// TestSeedDatabase_NoDatabase tests seeding without database
func TestSeedDatabase_NoDatabase(t *testing.T) {
	// Save original GormDB
	originalGormDB := GormDB

	// Set GormDB to nil
	GormDB = nil

	defer func() {
		// This function will panic with nil pointer, catch it
		if r := recover(); r == nil {
			t.Error("Expected panic when GormDB is nil")
		}
		// Restore original GormDB
		GormDB = originalGormDB
	}()

	// This will panic, which we expect
	_ = seedDatabase()
}
