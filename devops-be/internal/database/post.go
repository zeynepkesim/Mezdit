package database

import (
	"database/sql"
	"devops-be/internal/models"
	"fmt"
)

func AddPost(publisherID int, publisherName string, header string, contentText string, contentPhotoID *int, isPrivate bool) (*models.Post, error) {
	if publisherID <= 0 {
		return nil, fmt.Errorf("invalid publisher id")
	}

	// If publisher name is empty, set to Anonymous
	if publisherName == "" {
		publisherName = "Anonymous"
	}

	if header == "" {
		return nil, fmt.Errorf("header cannot be empty")
	}

	if contentText == "" {
		return nil, fmt.Errorf("content text cannot be empty")
	}

	// Check if publisher exists
	user, err := GetUserByID(publisherID)
	if err != nil {
		return nil, fmt.Errorf("publisher not found")
	}

	// Insert post and get image path via JOIN
	query := `
		WITH inserted_post AS (
			INSERT INTO posts (publisher_id, publisher_name, header, content_text, content_photo_id, is_private)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, publisher_id, publisher_name, header, content_text, content_photo_id, is_private, vote_count, created_at
		)
		SELECT
			ip.id, ip.publisher_id, ip.publisher_name, ip.header, ip.content_text,
			ip.content_photo_id, ip.is_private, ip.vote_count, ip.created_at,
			i.path as content_photo_path
		FROM inserted_post ip
		LEFT JOIN images i ON ip.content_photo_id = i.id`

	var post models.Post
	var contentPhotoIDNull sql.NullInt32
	var contentPhotoPath sql.NullString

	err = DB.QueryRow(query, publisherID, publisherName, header, contentText, contentPhotoID, isPrivate).Scan(
		&post.ID, &post.PublisherID, &post.PublisherName, &post.Header, &post.ContentText,
		&contentPhotoIDNull, &post.IsPrivate, &post.VoteCount, &post.CreatedAt,
		&contentPhotoPath,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create post")
	}

	if contentPhotoIDNull.Valid {
		contentPhotoIDInt := int(contentPhotoIDNull.Int32)
		post.ContentPhotoID = &contentPhotoIDInt
	}

	if contentPhotoPath.Valid {
		post.ContentPhotoPath = &contentPhotoPath.String
	}

	// Get publisher photo from user
	if user.ProfilePhoto != nil {
		post.PublisherPhoto = user.ProfilePhoto
	}

	return &post, nil
}

func GetPosts() ([]models.Post, error) {
	return GetPostsForUser(0) // 0 means no user context
}

func GetPostsForUser(userID int) ([]models.Post, error) {
	var query string
	var args []interface{}

	if userID > 0 {
		query = `
			SELECT p.id, p.publisher_id, p.publisher_name, p.header, p.content_text,
				   p.content_photo_id, p.is_private, p.vote_count, p.created_at,
				   u.profile_photo as publisher_photo,
				   i.path as content_photo_path,
				   uv.upvoted as is_liked
			FROM posts p
			JOIN users u ON p.publisher_id = u.id
			LEFT JOIN images i ON p.content_photo_id = i.id
			LEFT JOIN user_votes uv ON p.id = uv.post_id AND uv.user_id = $1
			WHERE (p.is_private = FALSE OR p.publisher_id = $1)
			ORDER BY p.created_at DESC`
		args = []interface{}{userID}
	} else {
		query = `
			SELECT p.id, p.publisher_id, p.publisher_name, p.header, p.content_text,
				   p.content_photo_id, p.is_private, p.vote_count, p.created_at,
				   u.profile_photo as publisher_photo,
				   i.path as content_photo_path,
				   NULL as is_liked
			FROM posts p
			JOIN users u ON p.publisher_id = u.id
			LEFT JOIN images i ON p.content_photo_id = i.id
			WHERE p.is_private = FALSE
			ORDER BY p.created_at DESC`
		args = []interface{}{}
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve posts")
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var publisherPhoto sql.NullString
		var contentPhotoIDNull sql.NullInt32
		var contentPhotoPath sql.NullString
		var isLiked sql.NullBool

		err := rows.Scan(
			&post.ID, &post.PublisherID, &post.PublisherName, &post.Header, &post.ContentText,
			&contentPhotoIDNull, &post.IsPrivate, &post.VoteCount, &post.CreatedAt,
			&publisherPhoto, &contentPhotoPath, &isLiked,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to process post data")
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

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error while reading posts")
	}

	if posts == nil {
		return []models.Post{}, nil
	}

	return posts, nil
}

func EditPost(postID int, userID int, newHeader string, newContentText string, newPublisherName string, isPrivate bool) (*models.Post, error) {
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

	tx, err := DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction")
	}
	defer tx.Rollback()

	// Check if post exists and get its publisher_id
	var publisherID int
	err = tx.QueryRow("SELECT publisher_id FROM posts WHERE id = $1", postID).Scan(&publisherID)
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

	// Update the post header, content, publisher name, and privacy
	_, err = tx.Exec(
		"UPDATE posts SET header = $1, content_text = $2, publisher_name = $3, is_private = $4 WHERE id = $5",
		newHeader, newContentText, newPublisherName, isPrivate, postID,
	)
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