package models

import "time"

type User struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name         string    `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Username     string    `json:"username" gorm:"column:username;type:varchar(100);uniqueIndex;not null"`
	Password     string    `json:"-" gorm:"column:password;type:text;not null"` // Don't expose password in JSON
	ProfilePhoto *string   `json:"profilePhoto" gorm:"column:profile_photo;type:text"`
	CreatedAt    time.Time `json:"createdAt" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

type Image struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	PostID    *int      `json:"postId" gorm:"column:post_id;index"`
	Path      string    `json:"path" gorm:"column:path;type:text;not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GORM
func (Image) TableName() string {
	return "images"
}

type Post struct {
	ID               int       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	PublisherID      int       `json:"publisherId" gorm:"column:publisher_id;not null;index"`
	PublisherName    string    `json:"publisherName" gorm:"column:publisher_name;type:varchar(255);not null"`    // From posts table
	PublisherPhoto   *string   `json:"publisherPhoto" gorm:"-"`   // JOIN from users table - not a database column
	Header           string    `json:"header" gorm:"column:header;type:varchar(500);not null"`           // Post header/title
	ContentText      string    `json:"contentText" gorm:"column:content_text;type:text;not null"`
	ContentPhotoID   *int      `json:"contentPhotoId" gorm:"column:content_photo_id;index"`
	ContentPhotoPath *string   `json:"contentPhotoPath" gorm:"-"` // JOIN from images table - not a database column
	IsPrivate        bool      `json:"isPrivate" gorm:"column:is_private;default:false"`        // Privacy setting: true=private (only owner can see), false=public
	VoteCount        int       `json:"voteCount" gorm:"column:vote_count;default:0"`
	IsLiked          *bool     `json:"isLiked" gorm:"-"` // User's vote status - not a database column
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GORM
func (Post) TableName() string {
	return "posts"
}

type UserVote struct {
	PostID  int  `json:"postId" gorm:"column:post_id;primaryKey;not null"`
	UserID  int  `json:"userId" gorm:"column:user_id;primaryKey;not null"`
	Upvoted bool `json:"upvoted" gorm:"column:upvoted;not null"`
}

// TableName specifies the table name for GORM
func (UserVote) TableName() string {
	return "user_votes"
}

// Vote is kept for backwards compatibility with existing code
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
	IsPrivate      bool   `json:"isPrivate"`
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
	IsPrivate     bool   `json:"isPrivate"`
}

type ChangePasswordRequest struct {
	UserID          int    `json:"userId"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
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