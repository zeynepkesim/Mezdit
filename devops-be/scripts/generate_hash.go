package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Test passwords for sample users
	passwords := map[string]string{
		"john_doe":   "password123",
		"jane_smith": "password123",
		"bob_wilson": "password123",
	}

	fmt.Println("-- Bcrypt password hashes for sample users")
	fmt.Println("-- Default password for all users: password123")
	fmt.Println()

	for username, password := range passwords {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Error generating hash for %s: %v", username, err)
		}
		fmt.Printf("-- Username: %s | Password: %s\n", username, password)
		fmt.Printf("'%s',\n", string(hash))
	}
}
