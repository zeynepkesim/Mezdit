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
      "profilePhoto": null,
      "createdAt": "2023-12-01T09:00:00Z"
    }}
    ```
  - Response 401 (Authentication Failed)
    ```json
    {"success": false, "error": {"code": "authentication_failed", "message": "Invalid username or password"}}
    ```

### Posts

- `GET /posts` — Returns all posts (newest → oldest).
  - Response 200
    ```json
    {"success": true, "data": [
      {
        "id": 1,
        "publisherId": 2,
        "publisherName": "John Doe",
        "publisherPhoto": null,
        "contentText": "Hello world!",
        "contentPhotoId": null,
        "contentPhotoPath": null,
        "voteCount": 5,
        "createdAt": "2023-12-01T10:00:00Z"
      }
    ]}
    ```
  - No posts: `{"success": true, "data": []}`

- `POST /posts` — Creates a new post.
  - Request body
    ```json
    {"publisherId": 1, "contentText": "string", "contentPhotoId": null}
    ```
  - Response 201 (Success)
    ```json
    {"success": true, "data": {
      "id": 1,
      "publisherId": 2,
      "publisherName": "John Doe",
      "publisherPhoto": null,
      "contentText": "Hello world!",
      "contentPhotoId": null,
      "contentPhotoPath": null,
      "voteCount": 0,
      "createdAt": "2023-12-01T10:00:00Z"
    }}
    ```
  - Response 400 (Validation Error)
    ```json
    {"success": false, "error": {"code": "validation_error", "message": "Content text cannot be empty"}}
    ```
  - Response 400 (Invalid Publisher)
    ```json
    {"success": false, "error": {"code": "invalid_publisher", "message": "Publisher user does not exist"}}
    ```

### Voting

- `POST /vote` — Upvote/Downvote.
  - Request body
    ```json
    {"postId": 1, "userId": 2, "upvoted": true}
    ```
  - Response 200 (Success): Same as post creation response with updated voteCount
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
- **Method Not Allowed**: `{"success": false, "error": {"code": "method_not_allowed", "message": "Only GET and POST methods are allowed"}}`
- **Server Error**: `{"success": false, "error": {"code": "server_error", "message": "An internal server error occurred"}}`

## Features

- **PostgreSQL Database**: Full database integration with proper relations
- **User System**: Users table with authentication-ready structure  
- **Image Support**: Ready for post images (images table linked to posts)
- **Voting System**: Each user can vote on a post only once
- **Vote Logic**: If a different vote type is sent to the same post, the previous vote is cancelled (net `voteCount` accordingly changes and can be negative)
- **Sorting**: Posts are sorted by creation date (newest first)
- **Backward Compatibility**: Legacy API format still works alongside new format
- **JOIN Queries**: Efficient data retrieval with publisher names and image paths
- **Transaction Support**: ACID compliance for vote operations

## Example Requests (curl)

```bash
# Login user
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe", "password":"$2a$10$example_hashed_password_1"}' \
  http://localhost:8080/login

# Get user info
curl http://localhost:8080/users/john_doe

# List all posts
curl http://localhost:8080/posts

# Create post
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"publisherId":1, "contentText":"Hello world from API!"}' \
  http://localhost:8080/posts

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

# Example error responses
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"invalid_user", "password":"wrong_pass"}' \
  http://localhost:8080/login
# Returns: {"success":false,"error":{"code":"authentication_failed","message":"Invalid username or password"}}
```
