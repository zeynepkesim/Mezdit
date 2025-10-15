package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Password     string    `json:"-"` // Don't expose password in JSON
	ProfilePhoto *string   `json:"profilePhoto"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Image struct {
	ID        int       `json:"id"`
	PostID    *int      `json:"postId"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"createdAt"`
}

type Post struct {
	ID               int       `json:"id"`
	PublisherID      int       `json:"publisherId"`
	PublisherName    string    `json:"publisherName"`    // From posts table
	PublisherPhoto   *string   `json:"publisherPhoto"`   // JOIN from users table
	Header           string    `json:"header"`           // Post header/title
	ContentText      string    `json:"contentText"`
	ContentPhotoID   *int      `json:"contentPhotoId"`
	ContentPhotoPath *string   `json:"contentPhotoPath"` // JOIN from images table
	VoteCount        int       `json:"voteCount"`
	IsLiked          *bool     `json:"isLiked"` // User's vote status: true=liked, false=disliked, null=no vote
	CreatedAt        time.Time `json:"createdAt"`
}

type Vote struct {
	PostID  int  `json:"postId"`
	UserID  int  `json:"userId"`
	Upvoted bool `json:"upvoted"`
}

// Request types
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreatePostRequest struct {
	PublisherID    int    `json:"publisherId"`
	PublisherName  string `json:"publisherName"`
	Name           string `json:"name"`
	Header         string `json:"header"`
	ContentText    string `json:"contentText"`
	ContentPhotoID *int   `json:"contentPhotoId"`
}

type VoteRequest struct {
	PostID  int  `json:"postId"`
	UserID  int  `json:"userId"`
	Upvoted bool `json:"upvoted"`
}

type EditPostRequest struct {
	PostID        int    `json:"postId"`
	UserID        int    `json:"userId"`
	Header        string `json:"header"`
	ContentText   string `json:"contentText"`
	PublisherName string `json:"publisherName"`
}

// Response types
type APISuccess struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type APIErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIError struct {
	Success bool         `json:"success"`
	Error   APIErrorBody `json:"error"`
}