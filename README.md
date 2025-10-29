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

2. **Add GitLab CI/CD variables** (Settings → CI/CD → Variables):
   - `SSH_PRIVATE_KEY` - Your private key content
   - `VM_HOST` - VM hostname/IP address
   - `VM_USER` - VM username (e.g., `ubuntu`)

3. **Configure VM** (see [DEPLOYMENT.md](DEPLOYMENT.md) for details):
   - Add SSH public key to VM
   - Create systemd service
   - Grant sudo permissions
   - Configure PostgreSQL

4. **Push to main branch:**
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

## Authentication

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

**Message Format:**
```json
{"type": "POST_CREATED|POST_UPDATED|POST_VOTED|POST_PRIVACY_CHANGED", "payload": {post object}}
```

**Message Types:**
- `POST_CREATED` - New post created (from `POST /posts`, `POST /posts/upload`)
- `POST_UPDATED` - Post edited (from `POST /posts/edit*`)
- `POST_VOTED` - Vote added/changed/removed (from `POST /vote`)
- `POST_PRIVACY_CHANGED` - Reserved for future use

**Features:** Bidirectional WebSocket, multi-client broadcasts, <100ms latency, auto-reconnect, heartbeat (60s pong wait, 54s ping)

**Example:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (event) => {
  const {type, payload} = JSON.parse(event.data);
  if (type === 'POST_CREATED') addPostToUI(payload);
  if (type === 'POST_VOTED') updateVoteCount(payload);
};
```

**Testing:** `wscat -c ws://localhost:8080/ws` (install: `npm i -g wscat`)

### Common Errors
- `invalid_content_type`, `invalid_json`, `invalid_file_type`, `method_not_allowed`, `server_error`

## Features

- **🔒 Security**: Comprehensive security measures
  - **XSS Prevention**: Input sanitization on all user inputs (backend & frontend)
  - **Secrets Detection**: Automated Gitleaks scanning in CI/CD pipeline
  - **Input Validation**: Length limits and pattern validation to prevent attacks
  - **HTML Escaping**: All user-generated content properly escaped
  - **DoS Prevention**: Input size limits on all endpoints
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

