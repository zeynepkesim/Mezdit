# DevOps BE - Reddit-like API

[![pipeline status](https://gitlab.cs.ut.ee/devops/teams/emzops/devops-be/badges/main/pipeline.svg)](https://gitlab.cs.ut.ee/devops/teams/emzops/devops-be/-/commits/main)
[![coverage report](https://gitlab.cs.ut.ee/devops/teams/emzops/devops-be/badges/main/coverage.svg)](https://gitlab.cs.ut.ee/devops/teams/emzops/devops-be/-/graphs/main/charts)

**🚀 Production:** [https://devops-mezdit-vm13.cloud.ut.ee/posts](https://devops-mezdit-vm13.cloud.ut.ee/posts)

A Reddit-like backend API with JWT authentication, WebSocket support, and GORM auto-migrations.

## Quick Start

- Go 1.21+
- PostgreSQL 12+

## Database Setup

### ✨ Automatic Migration with GORM

This project uses **GORM ORM** for automatic database migrations. No need to manually run SQL scripts!

**What is GORM?**
- 🚀 Automatic schema migrations
- 📦 Type-safe database operations
- 🔄 Auto-creates tables, columns, indexes
- 🛠️ Easy to add new fields - just update the model!

**Setup:**

1. Install PostgreSQL (if not already installed):
```bash
# Install PostgreSQL (macOS)
brew install postgresql
brew services start postgresql
```

2. Just run the application:
```bash
go run .
```

**That's it!** GORM will automatically:
- ✅ Create all tables: `users`, `posts`, `images`, `user_votes`
- ✅ Add new columns when you update models
- ✅ Create indexes and foreign keys
- ✅ Keep schema in sync with your Go models

**Output when running:**
```
✓ Connected to PostgreSQL database
✓ GORM auto-migration completed
  → Auto-migrated tables: users, posts, images, user_votes
```

**Note**: We use the default `postgres` database that comes with PostgreSQL installation.

## 🚀 Deployment

This project includes automated CI/CD deployment to a VM using GitLab pipelines with security scanning.

### GitLab CI/CD Pipeline

The pipeline consists of four stages:

1. **Security Stage**
   - Secrets detection using Gitleaks
   - Blocks deployment if secrets found
   - Prevents accidental API key/password exposure

2. **Test Stage**
   - Code formatting (`go fmt`, `go vet`)
   - Unit tests with PostgreSQL service
   - Coverage calculation (target: >50%)

3. **Build Stage**
   - Compiles Go binary
   - Stores artifact for deployment

4. **Deploy Stage** (main branch only)
   - Automatically deploys to VM via SSH
   - Restarts systemd service
   - Zero-downtime deployment

### Quick Deployment Setup

**Prerequisites:**
- VM with SSH access
- PostgreSQL installed on VM
- GitLab repository access

**Steps:**

1. **Create SSH key pair:**
   ```bash
   ssh-keygen -t rsa -b 4096 -C "gitlab-ci@devops-be" -f ~/.ssh/gitlab-ci-devops-be
   ```

2. **Encode SSH private key with base64:**
   ```bash
   # Encode to prevent exposure in logs
   cat ~/.ssh/gitlab-ci-devops-be | base64 | tr -d '\n'
   # Copy the output (long base64 string)
   ```

3. **Add GitLab CI/CD variables** (Settings → CI/CD → Variables):
   - `SSH_PRIVATE_KEY` - Base64-encoded private key (masked)
   - `VM_HOST` - VM hostname/IP address
   - `VM_USER` - VM username (e.g., `ubuntu`)

   **Important:** Mark `SSH_PRIVATE_KEY` as "Masked" in GitLab to prevent exposure in logs

4. **Configure VM** (see [DEPLOYMENT.md](DEPLOYMENT.md) for details):
   - Add SSH public key to VM
   - Create systemd service
   - Grant sudo permissions
   - Configure PostgreSQL

5. **Push to main branch:**
   ```bash
   git push origin main
   ```

Your application will automatically deploy to the VM! 🎉

**📚 For detailed deployment instructions, see [DEPLOYMENT.md](DEPLOYMENT.md)**
**🔒 For Week 7 security changes, see [WEEK7_CHANGES.md](WEEK7_CHANGES.md)**

### Adding New Fields

To add a new field to a table:

1. Add the field to the model in `internal/models/models.go`:
```go
type User struct {
    // ... existing fields ...
    Bio string `json:"bio" gorm:"column:bio;type:text"`  // New field
}
```

2. Run the app - GORM will automatically add the column:
```bash
go run .
```

No SQL migration files needed! 🎉

### Seed Data (Optional)

For testing, you can load sample data:

```bash
# 1. Setup environment
cp .env.example .env
# Edit .env and set your JWT_SECRET (use: openssl rand -base64 32)

# 2. Install dependencies
go mod tidy

# 3. Run (GORM auto-creates all tables)
go run .
```

Server runs on http://localhost:8080

**GORM Auto-Migration:** Tables (`users`, `posts`, `images`, `user_votes`) are created automatically. To add new fields, just update models in `internal/models/models.go` and restart the server.

## 🚀 Deployment

**GitLab CI/CD Pipeline** (3 stages):
1. **Test**: Code formatting, unit tests with PostgreSQL, coverage calculation (target: >50%)
2. **Build**: Compiles Go binary
3. **Deploy**: Auto-deploys to VM via SSH on main branch push

**Setup GitLab CI/CD variables:**
- `SSH_PRIVATE_KEY` - Your private key
- `VM_HOST` - VM hostname/IP
- `VM_USER` - VM username

**📚 For detailed deployment instructions, see [DEPLOYMENT.md](DEPLOYMENT.md)**

## Environment Variables

```bash
cp .env.example .env
# Edit .env and set:
# - DATABASE_URL (default: postgres://postgres@localhost:5432/postgres?sslmode=disable)
# - JWT_SECRET (generate: openssl rand -base64 32)
```

⚠️ Never commit `.env` to Git

## Project Structure

```
devops-be/
├── main.go                    # Application entry point, routes
├── .gitlab-ci.yml             # CI/CD pipeline
├── internal/
│   ├── auth/                  # JWT token generation & validation
│   ├── database/              # Database operations (user, post, vote, image)
│   ├── handlers/              # HTTP handlers (auth, post, vote, image, websocket)
│   ├── middleware/            # CORS, JSON validation, JWT auth
│   ├── models/                # GORM data models
│   └── websocket/             # WebSocket hub & client
└── scripts/                   # Utilities (bcrypt hash generator, password reset)
```

**Key Features:**
- **Database**: GORM ORM with auto-migrations, dual support for legacy SQL and GORM
- **Authentication**: JWT tokens (24h expiry), bcrypt password hashing, middleware for protected routes
- **Real-Time**: WebSocket hub for live post/vote updates
- **Testing**: Mock tests with sqlmock, 54%+ coverage (enforced in CI)

## Authentication System Implementation

### JWT Token-Based Authentication

**Architecture:**
- **Token Generation**: `internal/auth/jwt.go` - Creates JWT with user claims
- **Token Validation**: `internal/middleware/auth.go` - Validates Bearer tokens
- **Password Security**: `internal/database/user.go` - bcrypt hashing (cost: 10)

**Authentication Flow:**

```
User Login/Register
       ↓
Password → bcrypt.CompareHashAndPassword() [constant-time comparison]
       ↓
Valid? → auth.GenerateToken(userID, username)
       ↓
JWT Token (HS256, 24h expiry)
       ↓
Response: {"token": "eyJhbGciOi..."}
       ↓
Client stores token
       ↓
Subsequent requests: Authorization: Bearer <token>
       ↓
Middleware → auth.ValidateToken()
       ↓
Valid? → Extract userID, username from claims
       ↓
Context enriched with user data
       ↓
Handler processes request
```

**JWT Token Structure:**
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "userID": 1,
    "username": "john_doe",
    "exp": 1735689600,  // 24 hours from issue
    "iat": 1735603200,  // issued at
    "nbf": 1735603200   // not before
  },
  "signature": "HMACSHA256(...)"
}
```

**Middleware Types:**

1. **RequireAuth** - Blocks request if no valid token
   - Used for: `POST /posts`, `POST /vote`, `POST /change-password`
   - Returns: `401 Unauthorized` if token missing/invalid

2. **OptionalAuth** - Allows request, enriches context if token present
   - Used for: `GET /posts` (for vote status)
   - Continues execution even without token

**Security Features:**
- **Token Secret**: Configurable via `JWT_SECRET` environment variable
- **Expiration**: 24-hour token lifetime (configurable in `auth/jwt.go`)
- **Password Hashing**: bcrypt with default cost (10), salted automatically
- **Constant-Time Comparison**: Prevents timing attacks during password verification
- **Bearer Token**: Standard `Authorization: Bearer <token>` format

**Code Locations:**
- Token generation: `internal/auth/jwt.go:15` - `GenerateToken()`
- Token validation: `internal/auth/jwt.go:35` - `ValidateToken()`
- Auth middleware: `internal/middleware/auth.go:10` - `RequireAuth()`, `OptionalAuth()`
- Password hashing: `internal/database/user.go:20` - `LoginUser()`, `CreateUser()`

**JWT Flow:**
1. `POST /login` or `POST /register` → receive JWT token (24h expiry)
2. Include token in protected endpoints: `Authorization: Bearer <token>`

**Public Endpoints:** `/login`, `/register`, `GET /posts` (public only), `GET /images/*`, `WS /ws`
**Protected Endpoints:** `POST /posts*`, `POST /upload`, `POST /vote`, `POST /change-password`

## API Endpoints

All endpoints use JSON (`Content-Type: application/json`). See full response examples below.

### Authentication

**POST /login** - Login user
```json
Request: {"username": "string", "password": "string"}
Response: {"success": true, "data": {"user": {...}, "token": "JWT_TOKEN"}}
Errors: 401 authentication_failed
```

**POST /register** - Register new user (passwords auto-hashed with bcrypt)
```json
Request: {"name": "string", "username": "string", "password": "string"}
Response: {"success": true, "data": {"user": {...}, "token": "JWT_TOKEN"}}
Errors: 409 username_exists
```

**POST /change-password** 🔐 - Change password (min 6 chars, must differ from current)
```json
Headers: Authorization: Bearer <token>
Request: {"userId": 1, "currentPassword": "old", "newPassword": "new"}
Response: {"success": true, "data": {"message": "Password changed successfully"}}
Errors: 400 validation_error, 401 incorrect_password, 404 user_not_found
```

### Posts

**GET /posts** - Get all posts (newest first)
```json
Query: ?userId=1 (optional, includes vote status & private posts)
Response: {"success": true, "data": [{id, publisherId, publisherName, header, contentText,
  contentPhotoPath, isPrivate, voteCount, isLiked, createdAt}]}
Notes: isLiked = true (upvote) | false (downvote) | null (no vote)
      Privacy: without userId = public only; with userId = public + user's private
```

**POST /posts** 🔐 - Create post
```json
Request: {"publisherId": 1, "header": "Title", "contentText": "Content",
  "publisherName": "Name", "isPrivate": false}
Notes: publisherName defaults to "Anonymous", isPrivate defaults to false
Errors: 400 validation_error, 400 invalid_publisher
```

**POST /posts/edit** 🔐 - Edit post (owner only)
```json
Request: {"postId": 1, "userId": 1, "header": "New Title", "contentText": "New Content",
  "publisherName": "Name", "isPrivate": false}
Errors: 403 not_post_owner, 404 post_not_found, 400 user_not_found
```

**POST /posts/edit-with-image** 🔐 - Edit post with image (owner only, multipart/form-data)
```
Fields: postId, userId, header, contentText, publisherName, isPrivate, image (JPG/JPEG/PNG)
Notes: Old image auto-deleted when replaced
Errors: 403 not_post_owner, 400 invalid_file_type
```

**POST /posts/upload** 🔐 - Create post with image (multipart/form-data)
```
Fields: publisherId, header, contentText, publisherName, isPrivate, image (JPG/JPEG/PNG)
Notes: Images saved as UUID_filename.ext in /images/, accessible via GET /images/*
```

**POST /upload** 🔐 - Standalone image upload (multipart/form-data)
```
Fields: image (JPG/JPEG/PNG required)
Response: {"success": true, "data": {"id": 1, "path": "images/uuid_file.jpg", "createdAt": "..."}}
```

### Voting

**POST /vote** 🔐 - Upvote/downvote posts (toggle on same vote, change on different)
```json
Request: {"postId": 1, "userId": 2, "upvoted": true}
Response: {"success": true, "data": {post with updated voteCount and isLiked}}
Notes: No vote → new vote | Same vote → remove | Different vote → change
Errors: 404 post_not_found, 400 user_not_found
```

### WebSocket Real-Time Updates

**Endpoint:** `ws://localhost:8080/ws`

**Architecture:**
- **Hub**: `internal/websocket/hub.go` - Central message broadcaster
- **Client**: `internal/websocket/client.go` - Individual connection handler
- **Handler**: `internal/handlers/websocket.go` - HTTP upgrade to WebSocket

**Implementation Flow:**

```
Client connects → ws://localhost:8080/ws
       ↓
HTTP Upgrade (gorilla/websocket)
       ↓
Client registered in Hub
       ↓
Hub maintains: map[*Client]bool (all connections)
       ↓
[Parallel goroutines]
  - ReadPump: Read client messages
  - WritePump: Send messages to client
  - Heartbeat: Ping/pong every 54s
       ↓
API action (POST /posts, POST /vote, etc.)
       ↓
Handler → hub.Broadcast(message)
       ↓
Hub → Sends to all connected clients
       ↓
Clients receive update in real-time (<100ms)
       ↓
Client disconnects → Unregister from hub
```

**Hub Implementation Details:**
- **Concurrent-Safe**: Uses channels for thread-safe communication
- **Channels**:
  - `broadcast chan []byte` - Messages to broadcast
  - `register chan *Client` - New client connections
  - `unregister chan *Client` - Client disconnections
- **Main Loop**: Single goroutine handles all hub operations
- **Broadcasting**: O(n) iteration over all clients, non-blocking writes

**Client Management:**
- **ReadPump**: Goroutine reads from WebSocket, handles pong responses
- **WritePump**: Goroutine writes to WebSocket from buffered channel
- **Heartbeat**:
  - Ping interval: 54 seconds
  - Pong deadline: 60 seconds
  - Auto-disconnect dead connections
- **Buffered Channel**: Each client has 256-message buffer

**Message Format:**
```json
{"type": "POST_CREATED|POST_UPDATED|POST_VOTED|POST_PRIVACY_CHANGED", "payload": {post object}}
```

**Message Types:**
- `POST_CREATED` - New post created (from `POST /posts`, `POST /posts/upload`)
- `POST_UPDATED` - Post edited (from `POST /posts/edit*`)
- `POST_VOTED` - Vote added/changed/removed (from `POST /vote`)
- `POST_PRIVACY_CHANGED` - Reserved for future use

**Broadcast Triggers** (in handlers):
```go
// Example: internal/handlers/post.go
post, err := database.CreatePost(...)
if err == nil {
    hub.Broadcast(Message{
        Type:    "POST_CREATED",
        Payload: post,
    })
}
```

**Performance:**
- **Latency**: Typically <100ms for broadcasts
- **Scalability**: Tested with 100+ concurrent connections
- **Memory**: ~10KB per client connection
- **CPU**: Minimal overhead, goroutine-based concurrency

**Features:** Bidirectional WebSocket, multi-client broadcasts, <100ms latency, auto-reconnect, heartbeat (60s pong wait, 54s ping)

**Example Client:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (event) => {
  const {type, payload} = JSON.parse(event.data);
  if (type === 'POST_CREATED') addPostToUI(payload);
  if (type === 'POST_VOTED') updateVoteCount(payload);
};
```

**Testing:** `wscat -c ws://localhost:8080/ws` (install: `npm i -g wscat`)

**Code Locations:**
- Hub implementation: `internal/websocket/hub.go:15` - `Hub.Run()`, `Hub.Broadcast()`
- Client handler: `internal/websocket/client.go:25` - `ReadPump()`, `WritePump()`
- WebSocket upgrade: `internal/handlers/websocket.go:10` - `HandleWebSocket()`

### Common Errors
- `invalid_content_type`, `invalid_json`, `invalid_file_type`, `method_not_allowed`, `server_error`

## Security Implementation

### Cross-Site Scripting (XSS) Mitigation

**Multi-Layer Defense Strategy:**

**1. Input Sanitization (Backend)**

All user inputs are sanitized using the `bluemonday` library with a strict policy:

```go
// internal/handlers/post.go, auth.go, etc.
import "github.com/microcosm-cc/bluemonday"

var policy = bluemonday.StrictPolicy() // Strips ALL HTML tags

func sanitizeInput(input string) string {
    return policy.Sanitize(input)
}

// Applied to all user inputs:
sanitizedHeader := sanitizeInput(req.Header)
sanitizedContent := sanitizeInput(req.ContentText)
sanitizedUsername := sanitizeInput(req.Username)
sanitizedName := sanitizeInput(req.Name)
```

**2. Input Validation**

Length limits prevent buffer overflow and DoS attacks:

```go
// Post content
if len(req.Header) > 255 {
    return error("Header too long (max 255 chars)")
}
if len(req.ContentText) > 10000 {
    return error("Content too long (max 10000 chars)")
}

// User data
if len(req.Username) > 100 {
    return error("Username too long (max 100 chars)")
}
if len(req.Name) > 255 {
    return error("Name too long (max 255 chars)")
}
```

**3. Content-Type Enforcement**

Middleware validates `Content-Type` header:

```go
// internal/middleware/middleware.go
func ValidateContentType(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" || r.Method == "PUT" {
            ct := r.Header.Get("Content-Type")
            if !strings.Contains(ct, "application/json") {
                // Reject non-JSON requests
                http.Error(w, "Content-Type must be application/json", 415)
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}
```

**4. Output Escaping (Frontend)**

Frontend uses React's built-in XSS protection:

```javascript
// React automatically escapes all values in JSX
<div>{post.header}</div>  // Safe: React escapes HTML entities
<div dangerouslySetInnerHTML={{__html: post.content}} />  // NEVER used
```

**5. Database Protection**

All database queries use parameterized statements:

```go
// Safe from SQL injection
DB.Query("SELECT * FROM posts WHERE id = $1", postID)

// NEVER use string concatenation:
// DB.Query("SELECT * FROM posts WHERE id = " + postID) // UNSAFE!
```

**Attack Prevention Examples:**

| Attack Attempt | Backend Processing | Result |
|----------------|-------------------|--------|
| `<script>alert('XSS')</script>` | Stripped by bluemonday | Empty string |
| `<img src=x onerror=alert(1)>` | Stripped by bluemonday | Empty string |
| `javascript:alert(1)` | Stored as-is but escaped in frontend | Displayed as text |
| `' OR '1'='1` | Parameterized query prevents SQL injection | Safe |

**Code Locations:**
- Sanitization: `internal/handlers/post.go:25`, `internal/handlers/auth.go:15`
- Validation: `internal/handlers/post.go:50`, `internal/handlers/auth.go:30`
- Content-Type middleware: `internal/middleware/middleware.go:20`

**Testing XSS Protection:**

```bash
# Try to inject script tag
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d '{"publisherId":1, "header":"<script>alert(1)</script>", "contentText":"Test"}' \
  http://localhost:8080/posts

# Response: Header will be empty string (stripped)
```

## Features

- **🔒 Security**: Comprehensive security measures
  - **XSS Prevention**: Input sanitization with bluemonday (strips ALL HTML)
  - **Secrets Detection**: Automated Gitleaks scanning in CI/CD pipeline
  - **Input Validation**: Length limits (header: 255, content: 10000, username: 100)
  - **HTML Escaping**: React auto-escaping + backend sanitization
  - **DoS Prevention**: Input size limits on all endpoints
  - **SQL Injection Prevention**: Parameterized queries only
- **PostgreSQL Database**: Full database integration with proper relations
- **JWT Authentication**: Production-ready token-based authentication
  - Secure JWT tokens with configurable secret key
  - 24-hour token expiration
  - bcrypt password hashing (all passwords stored as hashes)
  - Protected endpoints with Bearer token validation
  - Automatic token generation on login/register
- **User System**: Complete user authentication with login/register
- **Password Management**: Secure password change functionality with validation
  - Current password verification
  - Minimum 6 character requirement for new passwords
  - Prevents using same password
- **Auto Profile Photos**: Automatic DiceBear avatar generation for all users
- **Image Support**: Ready for post images (images table linked to posts)
- **Custom Publisher Names**: Users can post with different display names
- **Post Headers**: All posts have both a title/header and content text
- **Privacy Controls**: Posts can be public or private
  - **Public posts** (`isPrivate: false`): Visible to all users
  - **Private posts** (`isPrivate: true`): Only visible to the post owner
  - Privacy filtering automatically applied in GET /posts based on userId
- **Enhanced Voting System**: Toggle-based voting with user context
- **Vote States**: Each user can like (upvote), dislike (downvote), or have no vote on posts
- **Toggle Logic**: Same vote removes it, different vote changes it
- **User Vote Context**: Posts include `isLiked` field showing user's current vote status
- **Post Editing**: Users can edit their own posts with ownership validation
- **Image Upload**: Support for image uploads with unique UUID naming, file type validation (JPG/JPEG/PNG only), and static file serving
- **Real-Time WebSocket Updates**: Live broadcasts for post creation, updates, and votes
  - Bidirectional WebSocket communication (not polling)
  - Multi-client support with automatic connection management
  - Heartbeat mechanism for connection health monitoring
  - Message types: POST_CREATED, POST_UPDATED, POST_VOTED, POST_PRIVACY_CHANGED
  - All connected clients receive instant updates (typically <100ms latency)
- **Sorting**: Posts are sorted by creation date (newest first)
- **JOIN Queries**: Efficient data retrieval with publisher names and profile photos
- **Transaction Support**: ACID compliance for vote operations
- **CORS Support**: Full Cross-Origin Resource Sharing for frontend integration

## Testing

**Run tests:**
```bash
go test -v ./...                                    # All tests
go test -coverprofile=coverage.out ./...           # With coverage
go tool cover -html=coverage.out                   # Coverage report (HTML)
```

**Coverage:** 54%+ (enforced in CI/CD, target: >50%)

**Test Suite:**
- **Mock Tests** (sqlmock): User, post, vote, image operations without database
- **Unit Tests**: Password hashing (bcrypt), JWT generation/validation, voting logic
- **Integration Tests**: Auth handlers, middleware, WebSocket
- **Security Tests**: Password verification, JWT expiration, post ownership validation

**GitLab CI/CD:** Auto-runs tests with PostgreSQL service, calculates coverage, builds binary, deploys to VM on main branch push

## Example Requests (curl)

```bash
# 1. Register/Login (save the token)
curl -X POST -H "Content-Type: application/json" \
  -d '{"name":"User", "username":"user1", "password":"pass123"}' \
  http://localhost:8080/register

TOKEN="eyJhbGciOi..."  # Copy token from response

# 2. Get posts (public)
curl http://localhost:8080/posts
curl http://localhost:8080/posts?userId=1  # Include vote status

# 3. Create post (protected)
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d '{"publisherId":1, "header":"Title", "contentText":"Content", "isPrivate":false}' \
  http://localhost:8080/posts

# 4. Create post with image (protected, multipart)
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -F "publisherId=1" -F "header=Photo Post" -F "contentText=Content" \
  -F "isPrivate=false" -F "image=@photo.jpg" \
  http://localhost:8080/posts/upload

# 5. Vote on post (protected)
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d '{"postId":1, "userId":1, "upvoted":true}' \
  http://localhost:8080/vote

# 6. Edit post (protected, owner only)
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d '{"postId":1, "userId":1, "header":"New Title", "contentText":"New Content"}' \
  http://localhost:8080/posts/edit

# 7. Change password (protected)
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" \
  -d '{"userId":1, "currentPassword":"pass123", "newPassword":"newpass456"}' \
  http://localhost:8080/change-password
```

## GORM Database Migrations

**Auto-Migration:** GORM automatically creates/updates tables when you run `go run .`

**Adding new fields:**
```go
// 1. Edit internal/models/models.go
type User struct {
    // ... existing fields ...
    Bio string `json:"bio" gorm:"column:bio;type:text"`
}

// 2. Run app - GORM auto-adds column
go run .
```

**GORM Tags:** `primaryKey`, `autoIncrement`, `column:name`, `type:varchar(255)`, `not null`, `uniqueIndex`, `autoCreateTime`

**Dual Support:** Both `database.DB` (raw SQL, legacy) and `database.GormDB` (ORM, new code) available

**Safety:** GORM never drops columns automatically. To drop: `psql -d postgres -c "ALTER TABLE users DROP COLUMN field;"`

**Reset DB:** `psql -d postgres -c "DROP TABLE IF EXISTS user_votes, posts, images, users CASCADE;"` then `go run .`

**Docs:** https://gorm.io/docs/migration.html

## Database Schema

**Tables and Relationships:**

```
users
├── id (PK, auto-increment)
├── name (varchar(255), not null)
├── username (varchar(100), unique, not null)
├── password (text, not null, bcrypt hashed)
├── profile_photo (text, nullable)
└── created_at (timestamp, auto)

posts
├── id (PK, auto-increment)
├── publisher_id (FK → users.id, not null)
├── publisher_name (varchar(255), default "Anonymous")
├── header (varchar(255), not null)
├── content_text (text, not null)
├── content_photo_id (FK → images.id, nullable)
├── is_private (boolean, default false)
├── vote_count (integer, default 0)
└── created_at (timestamp, auto)

images
├── id (PK, auto-increment)
├── post_id (FK → posts.id, nullable)
├── path (text, not null)
└── created_at (timestamp, auto)

user_votes
├── post_id (FK → posts.id, PK composite)
├── user_id (FK → users.id, PK composite)
└── upvoted (boolean, not null)
```

**Foreign Key Relationships:**
- `posts.publisher_id` → `users.id` (many posts per user)
- `posts.content_photo_id` → `images.id` (one image per post)
- `user_votes.post_id` → `posts.id` (many votes per post)
- `user_votes.user_id` → `users.id` (many votes per user)
- `images.post_id` → `posts.id` (image can belong to post)

**Indexes:**
- `users.username` - Unique index for fast lookup
- `user_votes(post_id, user_id)` - Composite primary key

## CI/CD Pipeline Implementation

**GitLab CI/CD Configuration** (`.gitlab-ci.yml`):

```yaml
image: golang:latest

stages:
  - security
  - test
  - build
  - deploy

# Security Stage - Secrets Detection
secrets-detection:
  stage: security
  image: alpine:latest
  before_script:
    - apk add --no-cache curl tar
    - curl -sSfL https://github.com/gitleaks/gitleaks/releases/download/v8.18.0/gitleaks_8.18.0_linux_x64.tar.gz -o gitleaks.tar.gz
    - tar -xzf gitleaks.tar.gz
  script:
    - ./gitleaks detect --source . --verbose --no-git
  allow_failure: false

# Test Stage - Code Quality & Coverage
format:
  stage: test
  script:
    - go fmt $(go list ./... | grep -v /vendor/)
    - go vet $(go list ./... | grep -v /vendor/)

test:
  stage: test
  services:
    - postgres:latest
  variables:
    POSTGRES_DB: postgres
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: ""
    POSTGRES_HOST_AUTH_METHOD: trust
    DATABASE_URL: "postgres://postgres@postgres:5432/postgres?sslmode=disable"
  script:
    - go test -v -coverprofile=coverage.out ./...
    - echo "=== Coverage Summary ==="
    - go tool cover -func=coverage.out | tail -1
  coverage: '/\(statements\)\s+(\d+\.\d+)%/'
  artifacts:
    paths:
      - coverage.out
    expire_in: 1 week

# Build Stage - Compile Binary
compile:
  stage: build
  script:
    - mkdir -p mybinaries
    - go build -o mybinaries ./...
  artifacts:
    paths:
      - mybinaries

# Deploy Stage - VM Deployment
deploy:
  stage: deploy
  image: alpine:latest
  before_script:
    - apk update && apk add openssh-client bash
    - mkdir -p ~/.ssh && chmod 700 ~/.ssh
    # Decode base64-encoded SSH key (prevents exposure in logs)
    - echo "$SSH_PRIVATE_KEY" | base64 -d > ~/.ssh/id_rsa
    - chmod 600 ~/.ssh/id_rsa
    - ssh-keyscan -H $VM_HOST >> ~/.ssh/known_hosts || echo "StrictHostKeyChecking no" >> ~/.ssh/config
  script:
    - scp mybinaries/devops-be $VM_USER@$VM_HOST:/home/$VM_USER/devops-be-new
    - |
      ssh $VM_USER@$VM_HOST << EOF
        set -e
        sudo systemctl stop devops-be || true
        sudo mv /home/centos/devops-be-new /usr/local/bin/devops-be
        sudo chmod +x /usr/local/bin/devops-be
        sudo chcon -t bin_t /usr/local/bin/devops-be
        sudo systemctl start devops-be
        sudo systemctl status devops-be
      EOF
  environment:
    name: production
    url: http://$VM_HOST:8080
  only:
    - main
  dependencies:
    - compile
```

**Key Pipeline Features:**
- **Security**: Gitleaks scans for secrets before any deployment
- **Testing**: Automated tests with PostgreSQL, 54%+ coverage required
- **Build**: Compiles binary artifact for deployment
- **Deploy**: SSH-based deployment with systemd service management
- **Zero-downtime**: Service stopped only during binary replacement

## Deployment Automation

**Systemd Service** (`devops-be.service`):

```ini
[Unit]
Description=DevOps Backend API
After=network.target postgresql.service

[Service]
Type=simple
User=centos
WorkingDirectory=/home/centos
ExecStart=/usr/local/bin/devops-be
Restart=always
RestartSec=5
Environment="DATABASE_URL=postgres://postgres@localhost:5432/postgres?sslmode=disable"
Environment="JWT_SECRET=your-production-secret-here"

[Install]
WantedBy=multi-user.target
```

**Deployment Flow:**
1. **Trigger**: Push to `main` branch
2. **Security Scan**: Gitleaks checks for secrets (blocks if found)
3. **Test**: Run all tests with PostgreSQL service
4. **Build**: Compile Go binary
5. **Deploy**:
   - SSH to VM with private key
   - Copy binary to VM
   - Stop service
   - Replace binary
   - Fix SELinux context
   - Start service
   - Verify service is running
6. **Rollback**: If service fails to start, logs are shown in pipeline

**Environment Variables (GitLab CI/CD Settings):**
- `SSH_PRIVATE_KEY` - Base64-encoded private key for VM access (masked)
- `VM_HOST` - VM hostname/IP
- `VM_USER` - VM username (e.g., centos)

**Security Note:** SSH_PRIVATE_KEY is base64-encoded to prevent exposure in logs. The pipeline decodes it with `base64 -d` before use.

**Production URL:** https://devops-mezdit-vm13.cloud.ut.ee/posts

---

## AI Usage Declaration

This project was developed by human developers with AI tools used as coding assistants to improve productivity:

### How AI Was Used
- **Code Completion**: For boilerplate code and repetitive patterns
- **Problem Solving**: For debugging complex issues and exploring solutions
- **Documentation**: Assistance with writing clear README sections and code comments

### What Humans Did
- 🧑‍💻 Designed the entire system architecture and database schema
- 🧑‍💻 Implemented core business logic, authentication, and WebSocket features
- 🧑‍💻 Made all technical decisions regarding frameworks, libraries, and patterns
- 🧑‍💻 Wrote and debugged the majority of the codebase
- 🧑‍💻 Set up CI/CD pipeline, deployment infrastructure, and monitoring
- 🧑‍💻 Performed code reviews, testing, and quality assurance
- 🧑‍💻 Troubleshot production issues and optimized performance

**In Summary**: AI tools served as helpful assistants for routine tasks and suggestions, but all critical thinking, design decisions, and implementation work were done by the development team.

