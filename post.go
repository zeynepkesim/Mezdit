package main

import (
	"sort"
	"sync"
	"time"
)

type Post struct {
	ID        int       `json:"id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	LikeCount int       `json:"likeCount"`
	CreatedAt time.Time `json:"createdAt"`
}

var (
	posts         []Post
	postIDCounter int
	mu            sync.Mutex
	// postID -> username -> vote type (true=upvote, false=downvote)
	userVotes map[int]map[string]bool
)

func init() {
	userVotes = make(map[int]map[string]bool)
}

func AddPost(author, content string) Post {
	mu.Lock()
	defer mu.Unlock()

	if author == "" {
		author = "anonymous"
	}

	postIDCounter++
	post := Post{
		ID:        postIDCounter,
		Author:    author,
		Content:   content,
		CreatedAt: time.Now(),
	}
	posts = append(posts, post)
	return post
}

func VotePost(id int, like bool, username string) *Post {
	mu.Lock()
	defer mu.Unlock()
	
	for i := range posts {
		if posts[i].ID == id {
			// Check if user has voted on this post
			if userVotes[id] == nil {
				userVotes[id] = make(map[string]bool)
			}
			
			// Has the user voted before?
			if previousVote, exists := userVotes[id][username]; exists {
				// If same vote type, no change
				if previousVote == like {
					return &posts[i]
				}
				// If different vote type, cancel previous vote effect
				if previousVote {
					posts[i].LikeCount-- // cancel previous upvote
				} else {
					posts[i].LikeCount++ // cancel previous downvote
				}
			}
			
			// Record new vote
			userVotes[id][username] = like
			if like {
				posts[i].LikeCount++
			} else {
				posts[i].LikeCount--
			}
			
			return &posts[i]
		}
	}
	return nil
}

func GetPosts() []Post {
	mu.Lock()
	defer mu.Unlock()
	if posts == nil {
		return []Post{}
	}
	// Create copy slice and sort by CreatedAt (newest -> oldest)
	out := make([]Post, len(posts))
	copy(out, posts)
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}
