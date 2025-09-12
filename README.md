# DevOps BE - Simple Reddit-like API

## Running

```bash
cd devops-be
go run .
```

Server: http://localhost:8080

## Endpoints (JSON Contract)

All requests and responses are in JSON format. `Content-Type: application/json` header must be sent.

- `GET /posts` — Returns all posts (newest → oldest).
  - Response 200
    ```json
    {"success": true, "data": [ {"id":1, "author":"...", "content":"...", "likes":0, "dislikes":0, "createdAt":"RFC3339"} ]}
    ```
  - No posts: `{"success": true, "data": []}`

- `POST /posts` — Creates a new post.
  - Request body
    ```json
    {"author": "string", "content": "string"}
    ```
  - Response 201
    ```json
    {"success": true, "data": {"id":1, "author":"...", "content":"...", "likes":0, "dislikes":0, "createdAt":"RFC3339"}}
    ```
  - Error (e.g. validation):
    ```json
    {"success": false, "error": {"code": "validation_error", "message": "author and content fields cannot be empty"}}
    ```

- `POST /vote` — Upvote/Downvote.
  - Request body
    ```json
    {"id": 1, "like": true, "username": "string"}
    ```
  - Response 200
    ```json
    {"success": true, "data": {"id":1, "author":"...", "content":"...", "likes":1, "dislikes":0, "createdAt":"RFC3339"}}
    ```
  - Error examples:
    - Missing/wrong Content-Type: `{"success": false, "error": {"code": "invalid_content_type", "message": "Content-Type must be application/json"}}`
    - Validation: `{"success": false, "error": {"code": "validation_error", "message": "username field is required"}}`
    - Not found: `{"success": false, "error": {"code": "not_found", "message": "Post not found"}}`

## Features

- Each user can vote on a post only once
- If a different vote type is sent to the same post, the previous vote is cancelled
- Posts are sorted by creation date (newest first)

## Example Requests (curl)

```bash
# Create post (JSON)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"author":"john","content":"hello world"}' \
  http://localhost:8080/posts

# List posts
curl http://localhost:8080/posts

# Upvote (user: alice)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"id":1, "like": true, "username": "alice"}' \
  http://localhost:8080/vote

# Downvote (user: alice)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"id":1, "like": false, "username": "alice"}' \
  http://localhost:8080/vote

# Same user upvotes same post again (previous vote is cancelled)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"id":1, "like": true, "username": "alice"}' \
  http://localhost:8080/vote
```
