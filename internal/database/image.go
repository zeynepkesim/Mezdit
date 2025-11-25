package database

import (
	"database/sql"
	"devops-be/internal/models"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

func SaveImage(file multipart.File, filename string) (*models.Image, error) {
	// Get file extension and validate
	ext := filepath.Ext(filename)
	ext = strings.ToLower(ext)

	// Check allowed file types
	allowedTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}

	if !allowedTypes[ext] {
		return nil, fmt.Errorf("invalid file type. Only JPG, JPEG, and PNG files are allowed")
	}

	// Generate UUID for unique filename
	imageUUID := uuid.New().String()

	// Create unique filename: UUID_originalname
	uniqueFilename := fmt.Sprintf("%s_%s", imageUUID, filename)

	// Ensure images directory exists
	imagesDir := "images"
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create images directory: %v", err)
	}

	filePath := filepath.Join(imagesDir, uniqueFilename)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	// Copy uploaded file to destination
	_, err = io.Copy(dst, file)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	// Insert image record to database
	query := `
		INSERT INTO images (path, created_at)
		VALUES ($1, $2)
		RETURNING id, post_id, path, created_at`

	var image models.Image
	var postIDNull sql.NullInt32

	err = DB.QueryRow(query, filePath, time.Now()).Scan(
		&image.ID, &postIDNull, &image.Path, &image.CreatedAt,
	)

	if err != nil {
		// Clean up file if database insert fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save image record: %v", err)
	}

	if postIDNull.Valid {
		postIDInt := int(postIDNull.Int32)
		image.PostID = &postIDInt
	}

	return &image, nil
}

func UpdateImagePostID(imageID int, postID int) error {
	query := `UPDATE images SET post_id = $1 WHERE id = $2`
	_, err := DB.Exec(query, postID, imageID)
	if err != nil {
		return fmt.Errorf("failed to update image post_id: %v", err)
	}
	return nil
}

func EditPostWithImage(postID int, userID int, newHeader string, newContentText string, newPublisherName string, isPrivate bool, imageFile multipart.File, imageFilename string) (*models.Post, error) {
	if postID <= 0 {
		return nil, fmt.Errorf("invalid post id")
	}

	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	if newHeader == "" {
		return nil, fmt.Errorf("header cannot be empty")
	}

	if newContentText == "" {
		return nil, fmt.Errorf("content text cannot be empty")
	}

	// If publisherName is empty, set to Anonymous
	if newPublisherName == "" {
		newPublisherName = "Anonymous"
	}

	fmt.Printf("EditPostWithImage called: postID=%d, userID=%d, hasImage=%v\n", postID, userID, imageFile != nil)

	tx, err := DB.Begin()
	if err != nil {
		fmt.Printf("Failed to begin transaction: %v\n", err)
		return nil, fmt.Errorf("failed to start transaction")
	}
	defer tx.Rollback()

	// Check if post exists and get its publisher_id and current image
	var publisherID int
	var currentImageID sql.NullInt32
	err = tx.QueryRow("SELECT publisher_id, content_photo_id FROM posts WHERE id = $1", postID).Scan(&publisherID, &currentImageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("database error occurred")
	}

	// Security check: verify that the user is the owner of the post
	if publisherID != userID {
		return nil, fmt.Errorf("you are not the owner of this post")
	}

	// Check if user exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("database error occurred")
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	var newImageID *int

	// Handle image update if provided
	if imageFile != nil && imageFilename != "" {
		// Save the new image
		savedImage, err := SaveImage(imageFile, imageFilename)
		if err != nil {
			return nil, fmt.Errorf("failed to save image: %v", err)
		}

		// Update the image to link with this post
		_, err = tx.Exec("UPDATE images SET post_id = $1 WHERE id = $2", postID, savedImage.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to link image to post")
		}

		newImageID = &savedImage.ID

		// Delete old image file if it exists
		if currentImageID.Valid {
			var oldImagePath string
			err = tx.QueryRow("SELECT path FROM images WHERE id = $1", currentImageID.Int32).Scan(&oldImagePath)
			if err == nil && oldImagePath != "" {
				// Delete old image file from filesystem
				os.Remove(oldImagePath)
				// Delete old image record from database
				_, err = tx.Exec("DELETE FROM images WHERE id = $1", currentImageID.Int32)
				if err != nil {
					// Log the error but don't fail the transaction
					fmt.Printf("Warning: failed to delete old image record: %v\n", err)
				}
			}
		}
	}

	// Update the post header, content, publisher name, privacy, and image if provided
	var updateQuery string
	var updateArgs []interface{}

	if newImageID != nil {
		updateQuery = "UPDATE posts SET header = $1, content_text = $2, publisher_name = $3, is_private = $4, content_photo_id = $5 WHERE id = $6"
		updateArgs = []interface{}{newHeader, newContentText, newPublisherName, isPrivate, *newImageID, postID}
	} else {
		updateQuery = "UPDATE posts SET header = $1, content_text = $2, publisher_name = $3, is_private = $4 WHERE id = $5"
		updateArgs = []interface{}{newHeader, newContentText, newPublisherName, isPrivate, postID}
	}

	_, err = tx.Exec(updateQuery, updateArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to update post")
	}

	// Get updated post with all details including isLiked for the user
	var post models.Post
	var contentPhotoIDNull sql.NullInt32
	var contentPhotoPath sql.NullString
	var publisherPhoto sql.NullString
	var isLiked sql.NullBool

	query := `
		SELECT p.id, p.publisher_id, p.publisher_name, p.header, p.content_text,
			   p.content_photo_id, p.is_private, p.vote_count, p.created_at,
			   u.profile_photo as publisher_photo,
			   i.path as content_photo_path,
			   uv.upvoted as is_liked
		FROM posts p
		JOIN users u ON p.publisher_id = u.id
		LEFT JOIN images i ON p.content_photo_id = i.id
		LEFT JOIN user_votes uv ON p.id = uv.post_id AND uv.user_id = $2
		WHERE p.id = $1`

	err = tx.QueryRow(query, postID, userID).Scan(
		&post.ID, &post.PublisherID, &post.PublisherName, &post.Header, &post.ContentText,
		&contentPhotoIDNull, &post.IsPrivate, &post.VoteCount, &post.CreatedAt,
		&publisherPhoto, &contentPhotoPath, &isLiked,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated post")
	}

	if publisherPhoto.Valid {
		post.PublisherPhoto = &publisherPhoto.String
	}
	if contentPhotoIDNull.Valid {
		contentPhotoIDInt := int(contentPhotoIDNull.Int32)
		post.ContentPhotoID = &contentPhotoIDInt
	}
	if contentPhotoPath.Valid {
		post.ContentPhotoPath = &contentPhotoPath.String
	}
	if isLiked.Valid {
		post.IsLiked = &isLiked.Bool
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to save changes")
	}

	return &post, nil
}