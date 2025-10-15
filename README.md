# DevOps BE - Simple Reddit-like API

## Prerequisites

- Go 1.21+
- PostgreSQL 12+

## Database Setup

1. Install PostgreSQL (if not already installed):
```bash
# Install PostgreSQL (macOS)
brew install postgresql
brew services start postgresql
```

2. Set up tables and sample data in your existing postgres database:
```bash
psql -d postgres -f setup.sql
```

**Note**: We use the default `postgres` database that comes with PostgreSQL installation.

## Environment Variables

Set the database connection string (optional):
```bash
export DATABASE_URL="postgres://postgres@localhost:5432/postgres?sslmode=disable"
```

If not set, it defaults to: `postgres://postgres@localhost:5432/postgres?sslmode=disable`

## Running

```bash
cd devops-be
go mod tidy
go run .
```

Server: http://localhost:8080

## Endpoints (JSON Contract)

All requests and responses are in JSON format. `Content-Type: application/json` header must be sent.

### Authentication

- `POST /login` — User authentication.
  - Request body
    ```json
    {"username": "string", "password": "string"}
    ```
  - Response 200 (Success)
    ```json
    {"success": true, "data": {
      "id": 1,
      "name": "John Doe",
      "username": "john_doe",
      "profilePhoto": "https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=1",
      "createdAt": "2023-12-01T09:00:00Z"
    }}
    ```
  - Response 401 (Authentication Failed)
    ```json
    {"success": false, "error": {"code": "authentication_failed", "message": "Invalid username or password"}}
    ```

- `POST /register` — User registration.
  - Request body
    ```json
    {"name": "Full Name", "username": "string", "password": "string"}
    ```
  - Response 201 (Success)
    ```json
    {"success": true, "data": {
      "id": 6,
      "name": "Full Name",
      "username": "newuser",
      "profilePhoto": "https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=6",
      "createdAt": "2023-12-01T11:00:00Z"
    }}
    ```
  - Response 409 (Username Exists)
    ```json
    {"success": false, "error": {"code": "username_exists", "message": "Username is already taken"}}
    ```

### Posts

- `GET /posts` — Returns all posts (newest → oldest).
  - Optional Query Parameter: `userId` — Get posts with user's vote status
  - Response 200 (without userId)
    ```json
    {"success": true, "data": [
      {
        "id": 1,
        "publisherId": 2,
        "publisherName": "John Doe",
        "publisherPhoto": null,
        "header": "First Post",
        "contentText": "Hello world!",
        "contentPhotoId": null,
        "contentPhotoPath": null,
        "voteCount": 5,
        "isLiked": null,
        "createdAt": "2023-12-01T10:00:00Z"
      }
    ]}
    ```
  - Response 200 (with userId)
    ```json
    {"success": true, "data": [
      {
        "id": 1,
        "publisherId": 2,
        "publisherName": "John Doe",
        "publisherPhoto": null,
        "header": "First Post",
        "contentText": "Hello world!",
        "contentPhotoId": null,
        "contentPhotoPath": null,
        "voteCount": 5,
        "isLiked": true,
        "createdAt": "2023-12-01T10:00:00Z"
      }
    ]}
    ```
  - **isLiked Values**:
    - `true`: User has liked (upvoted) the post
    - `false`: User has disliked (downvoted) the post
    - `null`: User hasn't voted or userId not provided
  - No posts: `{"success": true, "data": []}`

- `POST /posts` — Creates a new post.
  - Request body
    ```json
    {"publisherId": 1, "publisherName": "Custom Display Name", "header": "Post Title", "contentText": "string", "contentPhotoId": null}
    ```
  - **Note**: `publisherName` is optional. If empty or not provided, defaults to "Anonymous". `header` is required.
  - Response 201 (Success)
    ```json
    {"success": true, "data": {
      "id": 1,
      "publisherId": 2,
      "publisherName": "John Doe",
      "publisherPhoto": null,
      "header": "Post Title",
      "contentText": "Hello world!",
      "contentPhotoId": null,
      "contentPhotoPath": null,
      "voteCount": 0,
      "isLiked": null,
      "createdAt": "2023-12-01T10:00:00Z"
    }}
    ```
  - Response 400 (Validation Error)
    ```json
    {"success": false, "error": {"code": "validation_error", "message": "Header cannot be empty"}}
    ```
  - Response 400 (Invalid Publisher)
    ```json
    {"success": false, "error": {"code": "invalid_publisher", "message": "Publisher user does not exist"}}
    ```

- `POST /posts/edit` — Edit existing post (owner only).
  - Request body
    ```json
    {"postId": 1, "userId": 2, "header": "Updated Title", "contentText": "Updated content text", "publisherName": "New Display Name"}
    ```
  - **Optional Fields**:
    - `publisherName`: Updates the display name (defaults to "Anonymous" if empty or omitted)
  - **Required Fields**:
    - `header`: Post title/header (must not be empty)
  - **Security**: Only the post owner (publisher) can edit their posts
  - Response 200 (Success): Returns updated post with current data
    ```json
    {"success": true, "data": {
      "id": 1,
      "publisherId": 2,
      "publisherName": "New Display Name",
      "publisherPhoto": null,
      "header": "Updated Title",
      "contentText": "Updated content text",
      "contentPhotoId": null,
      "contentPhotoPath": null,
      "voteCount": 5,
      "isLiked": true,
      "createdAt": "2023-12-01T10:00:00Z"
    }}
    ```
  - Response 403 (Not Owner)
    ```json
    {"success": false, "error": {"code": "not_post_owner", "message": "You are not the owner of this post"}}
    ```
  - Response 404 (Post Not Found)
    ```json
    {"success": false, "error": {"code": "post_not_found", "message": "Post not found"}}
    ```
  - Response 400 (User Not Found)
    ```json
    {"success": false, "error": {"code": "user_not_found", "message": "User not found"}}
    ```
  - Response 400 (Validation Error)
    ```json
    {"success": false, "error": {"code": "validation_error", "message": "Content text cannot be empty"}}
    ```

- `POST /posts/edit-with-image` — Edit existing post with image update capability (owner only).
  - Content-Type: `multipart/form-data`
  - Form fields:
    - `postId`: Post ID (required)
    - `userId`: User ID (required)
    - `header`: Updated post header/title (required)
    - `contentText`: Updated post content (required)
    - `publisherName`: Updated display name (optional, defaults to "Anonymous")
    - `image`: New image file (optional, only JPG/JPEG/PNG allowed)
  - **Image Handling**:
    - If image is provided, it replaces the existing image
    - Old image file is automatically deleted
    - Same file type validation as other image uploads
  - **Security**: Only the post owner (publisher) can edit their posts
  - Response 200 (Success): Returns updated post with current data
    ```json
    {"success": true, "data": {
      "id": 1,
      "publisherId": 2,
      "publisherName": "Updated Name",
      "publisherPhoto": null,
      "header": "Updated Title",
      "contentText": "Updated content with new image",
      "contentPhotoId": 5,
      "contentPhotoPath": "images/uuid_newimage.jpg",
      "voteCount": 5,
      "isLiked": true,
      "createdAt": "2023-12-01T10:00:00Z"
    }}
    ```
  - Response 403 (Not Owner)
    ```json
    {"success": false, "error": {"code": "not_post_owner", "message": "You are not the owner of this post"}}
    ```
  - Response 404 (Post Not Found)
    ```json
    {"success": false, "error": {"code": "post_not_found", "message": "Post not found"}}
    ```
  - Response 400 (User Not Found)
    ```json
    {"success": false, "error": {"code": "user_not_found", "message": "User not found"}}
    ```
  - Response 400 (Invalid File Type)
    ```json
    {"success": false, "error": {"code": "invalid_file_type", "message": "invalid file type. Only JPG, JPEG, and PNG files are allowed"}}
    ```
  - Response 400 (Validation Error)
    ```json
    {"success": false, "error": {"code": "validation_error", "message": "Content text cannot be empty"}}
    ```

- `POST /posts/upload` — Create post with image upload.
  - Content-Type: `multipart/form-data`
  - Form fields:
    - `publisherId`: User ID (required)
    - `publisherName`: Display name (optional, defaults to "Anonymous")
    - `header`: Post title/header (required)
    - `contentText`: Post content (required)
    - `image`: Image file (optional, only JPG/JPEG/PNG allowed)
  - Response 201 (Success): Returns created post with image data
    ```json
    {"success": true, "data": {
      "id": 8,
      "publisherId": 122,
      "publisherName": "Image Tester",
      "publisherPhoto": "https://api.dicebear.com/9.x/adventurer-neutral/svg?seed=122",
      "header": "My Photo Post",
      "contentText": "Post with image!",
      "contentPhotoId": 2,
      "contentPhotoPath": "images/uuid_filename.jpg",
      "voteCount": 0,
      "isLiked": null,
      "createdAt": "2025-09-28T18:29:45.017517Z"
    }}
    ```
  - **Image Processing**: Images are saved with UUID_originalname format in `/images/` directory
  - **File Type Validation**: Only JPG, JPEG, and PNG files are allowed (case-insensitive)
  - **Static Access**: Images accessible via `GET /images/{filename}`

- `POST /upload` — Standalone image upload.
  - Content-Type: `multipart/form-data`
  - Form fields:
    - `image`: Image file (required, only JPG/JPEG/PNG allowed)
  - Response 200 (Success): Returns image data
    ```json
    {"success": true, "data": {
      "id": 1,
      "postId": null,
      "path": "images/uuid_filename.jpg",
      "createdAt": "2025-09-28T18:29:39.001133Z"
    }}
    ```
  - Response 400 (Invalid File Type)
    ```json
    {"success": false, "error": {"code": "invalid_file_type", "message": "invalid file type. Only JPG, JPEG, and PNG files are allowed"}}
    ```

### Voting

- `POST /vote` — Upvote/Downvote with toggle functionality.
  - Request body
    ```json
    {"postId": 1, "userId": 2, "upvoted": true}
    ```
  - **Toggle Behavior**:
    - If user hasn't voted: Creates new vote
    - If user votes same type again: Removes the vote (toggle off)
    - If user votes different type: Changes the vote
  - Response 200 (Success): Returns updated post with current vote status
    ```json
    {"success": true, "data": {
      "id": 1,
      "publisherId": 2,
      "publisherName": "John Doe",
      "publisherPhoto": null,
      "header": "First Post",
      "contentText": "Hello world!",
      "contentPhotoId": null,
      "contentPhotoPath": null,
      "voteCount": 1,
      "isLiked": true,
      "createdAt": "2023-12-01T10:00:00Z"
    }}
    ```
  - Response 404 (Post Not Found)
    ```json
    {"success": false, "error": {"code": "post_not_found", "message": "Post not found"}}
    ```
  - Response 400 (User Not Found)
    ```json
    {"success": false, "error": {"code": "user_not_found", "message": "User not found"}}
    ```

### Users

- `GET /users/{username}` — Get user by username.
  - Response 200
    ```json
    {"success": true, "data": {
      "id": 1,
      "name": "John Doe",
      "username": "john_doe",
      "profilePhoto": null,
      "createdAt": "2023-12-01T09:00:00Z"
    }}
    ```

### Common Error Examples
- **Invalid Content-Type**: `{"success": false, "error": {"code": "invalid_content_type", "message": "Content-Type must be application/json"}}`
- **Invalid JSON**: `{"success": false, "error": {"code": "invalid_json", "message": "Request body is not valid JSON"}}`
- **Invalid File Type**: `{"success": false, "error": {"code": "invalid_file_type", "message": "invalid file type. Only JPG, JPEG, and PNG files are allowed"}}`
- **Method Not Allowed**: `{"success": false, "error": {"code": "method_not_allowed", "message": "Only GET and POST methods are allowed"}}`
- **Server Error**: `{"success": false, "error": {"code": "server_error", "message": "An internal server error occurred"}}`

## Features

- **PostgreSQL Database**: Full database integration with proper relations
- **User System**: Complete user authentication with login/register
- **Auto Profile Photos**: Automatic DiceBear avatar generation for all users  
- **Image Support**: Ready for post images (images table linked to posts)
- **Custom Publisher Names**: Users can post with different display names
- **Enhanced Voting System**: Toggle-based voting with user context
- **Vote States**: Each user can like (upvote), dislike (downvote), or have no vote on posts
- **Toggle Logic**: Same vote removes it, different vote changes it
- **User Vote Context**: Posts include `isLiked` field showing user's current vote status
- **Post Editing**: Users can edit their own posts with ownership validation
- **Image Upload**: Support for image uploads with unique UUID naming, file type validation (JPG/JPEG/PNG only), and static file serving
- **Sorting**: Posts are sorted by creation date (newest first)
- **JOIN Queries**: Efficient data retrieval with publisher names and profile photos
- **Transaction Support**: ACID compliance for vote operations
- **CORS Support**: Full Cross-Origin Resource Sharing for frontend integration

## Testing

The project includes comprehensive unit tests with mock data to validate core functionality without requiring database setup.

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run specific test file
go test -v -run TestAddPostWithMockData
go test -v -run TestEditPostWithMockData
go test -v -run TestVotePostLogic

# Run tests with coverage
go test -cover ./...
```

### Test Coverage

#### Post Tests (`post_test.go`)

**TestAddPostWithMockData** - Post Creation Tests:
- ✅ Valid post creation with all fields
- ✅ Valid post with empty publisher name (defaults to "Anonymous")
- ✅ Valid post with image (contentPhotoId)
- ✅ Invalid publisher ID validation (zero/negative)
- ✅ Empty content text validation
- ✅ Non-existent publisher validation

**TestEditPostWithMockData** - Post Editing Tests:
- ✅ Valid post edit
- ✅ Valid post edit with photo update
- ✅ Edit with empty publisher name (defaults to "Anonymous")
- ✅ Invalid post ID validation (zero/negative)
- ✅ Invalid user ID validation (zero/negative)
- ✅ Empty content text validation
- ✅ Non-existent post validation
- ✅ Non-existent user validation
- ✅ **Security**: User is not the post owner (authorization check)

**TestVotePostLogic** - Voting Logic Tests:
- ✅ New upvote
- ✅ New downvote
- ✅ Toggle upvote off (remove vote)
- ✅ Change from upvote to downvote

#### Main Tests (`main_test.go`)

- TestLoginHandler (requires database setup - currently skipped)
- TestRegisterHandler (requires database setup - currently skipped)

### Test Features

- **Mock Data**: Tests use mock users and posts to avoid database dependencies
- **Validation Testing**: Comprehensive input validation checks
- **Security Testing**: Authorization checks (e.g., post ownership validation)
- **Image Support**: Tests include scenarios with image uploads (contentPhotoId)
- **Edge Cases**: Zero/negative IDs, empty strings, non-existent entities
- **Error Messages**: Verifies correct error codes and messages

### CI/CD Integration

The project uses GitLab CI/CD with the following pipeline:

```yaml
# .gitlab-ci.yml
test:
  stage: test
  script:
    - go test -v ./...
```

Tests run automatically on every commit to ensure code quality and catch regressions early.

## Example Requests (curl)

```bash
# Register new user (gets automatic DiceBear avatar)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"name":"New User", "username":"newuser", "password":"mypassword"}' \
  http://localhost:8080/register

# Login user
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe", "password":"$2a$10$example_hashed_password_1"}' \
  http://localhost:8080/login

# Get user info
curl http://localhost:8080/users/john_doe

# List all posts (without user context)
curl http://localhost:8080/posts

# List posts with user's vote status
curl "http://localhost:8080/posts?userId=1"

# Create post with custom name
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"publisherId":1, "publisherName":"My Custom Name", "header":"My First Post", "contentText":"Hello world from API!"}' \
  http://localhost:8080/posts

# Create post without name (defaults to "Anonymous")
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"publisherId":1, "header":"Anonymous Post", "contentText":"Anonymous post!"}' \
  http://localhost:8080/posts

# Edit a post (only owner can edit)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"postId":1, "userId":2, "header":"Updated Title", "contentText":"Updated post content", "publisherName":"New Display Name"}' \
  http://localhost:8080/posts/edit

# Edit post without publisherName (defaults to Anonymous)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"postId":1, "userId":2, "header":"Updated Title", "contentText":"Updated content only"}' \
  http://localhost:8080/posts/edit

# Edit post with image update
curl -X POST \
  -F "postId=1" \
  -F "userId=2" \
  -F "header=Updated Title" \
  -F "contentText=Updated post content with new image" \
  -F "publisherName=Updated Display Name" \
  -F "image=@newphoto.jpg" \
  http://localhost:8080/posts/edit-with-image

# Edit post with image but without publisherName (defaults to Anonymous)
curl -X POST \
  -F "postId=1" \
  -F "userId=2" \
  -F "header=Updated Title" \
  -F "contentText=Updated content with new image" \
  -F "image=@newphoto.jpg" \
  http://localhost:8080/posts/edit-with-image

# Edit post text only (without changing image)
curl -X POST \
  -F "postId=1" \
  -F "userId=2" \
  -F "header=Updated Title" \
  -F "contentText=Updated text content only" \
  -F "publisherName=Updated Name" \
  http://localhost:8080/posts/edit-with-image

# Upload standalone image
curl -X POST \
  -F "image=@photo.jpg" \
  http://localhost:8080/upload

# Create post with image upload
curl -X POST \
  -F "publisherId=1" \
  -F "publisherName=Image Publisher" \
  -F "header=My Photo" \
  -F "contentText=Check out this image!" \
  -F "image=@photo.jpg" \
  http://localhost:8080/posts/upload

# Upvote a post
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"postId":1, "userId":4, "upvoted": true}' \
  http://localhost:8080/vote

# Downvote a post
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"postId":1, "userId":4, "upvoted": false}' \
  http://localhost:8080/vote

# Toggle vote off (if user already voted, same vote removes it)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"postId":1, "userId":4, "upvoted": true}' \
  http://localhost:8080/vote

# Example error responses
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"invalid_user", "password":"wrong_pass"}' \
  http://localhost:8080/login
# Returns: {"success":false,"error":{"code":"authentication_failed","message":"Invalid username or password"}}
```
