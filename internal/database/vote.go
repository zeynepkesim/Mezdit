package database

import (
	"database/sql"
	"devops-be/internal/models"
	"fmt"
)

func VotePost(postID int, userID int, upvoted bool) (*models.Post, error) {
	if postID <= 0 {
		return nil, fmt.Errorf("invalid post id")
	}

	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	tx, err := DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction")
	}
	defer tx.Rollback()

	// Check if post exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)", postID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("database error occurred")
	}
	if !exists {
		return nil, fmt.Errorf("post not found")
	}

	// Check if user exists
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("database error occurred")
	}
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user has already voted
	var previousVote sql.NullBool
	err = tx.QueryRow(
		"SELECT upvoted FROM user_votes WHERE post_id = $1 AND user_id = $2",
		postID, userID,
	).Scan(&previousVote)

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("database error occurred")
	}

	// Calculate vote count change and decide action
	var voteCountDelta int = 0
	var shouldDeleteVote bool = false
	var shouldUpdateVote bool = false
	var shouldInsertVote bool = false

	if previousVote.Valid {
		// User has voted before
		if previousVote.Bool == upvoted {
			// Same vote (toggle off) - remove the vote
			shouldDeleteVote = true
			if upvoted {
				voteCountDelta = -1 // Remove upvote
			} else {
				voteCountDelta = 1 // Remove downvote
			}
		} else {
			// Different vote - change the vote
			shouldUpdateVote = true
			if previousVote.Bool {
				voteCountDelta = -2 // Remove upvote, add downvote
			} else {
				voteCountDelta = 2 // Remove downvote, add upvote
			}
		}
	} else {
		// New vote
		shouldInsertVote = true
		if upvoted {
			voteCountDelta = 1
		} else {
			voteCountDelta = -1
		}
	}

	// Execute the appropriate database operation
	if shouldDeleteVote {
		_, err = tx.Exec(
			"DELETE FROM user_votes WHERE post_id = $1 AND user_id = $2",
			postID, userID,
		)
	} else if shouldUpdateVote {
		_, err = tx.Exec(
			"UPDATE user_votes SET upvoted = $1 WHERE post_id = $2 AND user_id = $3",
			upvoted, postID, userID,
		)
	} else if shouldInsertVote {
		_, err = tx.Exec(
			"INSERT INTO user_votes (post_id, user_id, upvoted) VALUES ($1, $2, $3)",
			postID, userID, upvoted,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to record vote")
	}

	// Update post vote count
	if voteCountDelta != 0 {
		_, err = tx.Exec(
			"UPDATE posts SET vote_count = vote_count + $1 WHERE id = $2",
			voteCountDelta, postID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update vote count")
		}
	}

	// Get updated post with isLiked status
	var post models.Post
	var contentPhotoIDNull sql.NullInt32
	var contentPhotoPath sql.NullString
	var publisherPhoto sql.NullString
	var isLiked sql.NullBool

	query := `
		SELECT p.id, p.publisher_id, p.publisher_name, p.content_text,
			   p.content_photo_id, p.vote_count, p.created_at,
			   u.profile_photo as publisher_photo,
			   i.path as content_photo_path,
			   uv.upvoted as is_liked
		FROM posts p
		JOIN users u ON p.publisher_id = u.id
		LEFT JOIN images i ON p.content_photo_id = i.id
		LEFT JOIN user_votes uv ON p.id = uv.post_id AND uv.user_id = $2
		WHERE p.id = $1`

	err = tx.QueryRow(query, postID, userID).Scan(
		&post.ID, &post.PublisherID, &post.PublisherName, &post.ContentText,
		&contentPhotoIDNull, &post.VoteCount, &post.CreatedAt,
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