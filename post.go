package main

import (
    "sort"
    "sync"
    "time"
)

type Post struct {
    ID       int       `json:"id"`
    Author   string    `json:"author"`
    Content  string    `json:"content"`
    Likes    int       `json:"likes"`
    Dislikes int       `json:"dislikes"`
    CreatedAt time.Time `json:"createdAt"`
}

var (
    posts []Post
    postIDCounter int
    mu    sync.Mutex
    // postID -> kullanıcı adı -> oy türü (true=upvote, false=downvote)
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
        ID: postIDCounter,
        Author: author,
        Content: content,
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
            // Bu post için kullanıcı oy vermiş mi kontrol et
            if userVotes[id] == nil {
                userVotes[id] = make(map[string]bool)
            }
            
            // Kullanıcı daha önce oy vermiş mi?
            if previousVote, exists := userVotes[id][username]; exists {
                // Aynı oy türüyse değişiklik yok
                if previousVote == like {
                    return &posts[i]
                }
                // Farklı oy türüyse önceki oyu geri al
                if previousVote {
                    posts[i].Likes--
                } else {
                    posts[i].Dislikes--
                }
            }
            
            // Yeni oyu kaydet
            userVotes[id][username] = like
            if like {
                posts[i].Likes++
            } else {
                posts[i].Dislikes++
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
    // Kopya slice oluştur ve CreatedAt'e göre (yeni -> eski) sırala
    out := make([]Post, len(posts))
    copy(out, posts)
    sort.Slice(out, func(i, j int) bool {
        return out[i].CreatedAt.After(out[j].CreatedAt)
    })
    return out
}
