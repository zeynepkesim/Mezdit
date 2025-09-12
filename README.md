# DevOps BE - Basit Reddit Benzeri API

## Çalıştırma

```bash
cd devops-be
go run .
```

Sunucu: http://localhost:8080

## Endpointler (JSON Contract)

Tüm istekler ve yanıtlar JSON formatındadır. `Content-Type: application/json` header'ı gönderilmelidir.

- `GET /posts` — Tüm postları döner (yeni → eski sıralı).
  - Response 200
    ```json
    {"success": true, "data": [ {"id":1, "author":"...", "content":"...", "likes":0, "dislikes":0, "createdAt":"RFC3339"} ]}
    ```
  - Post yoksa: `{"success": true, "data": []}`

- `POST /posts` — Yeni post oluşturur.
  - Request body
    ```json
    {"author": "string", "content": "string"}
    ```
  - Response 201
    ```json
    {"success": true, "data": {"id":1, "author":"...", "content":"...", "likes":0, "dislikes":0, "createdAt":"RFC3339"}}
    ```
  - Hata (ör. validasyon):
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
  - Hatalar örnekleri:
    - Content-Type eksik/yanlış: `{"success": false, "error": {"code": "invalid_content_type", "message": "Content-Type must be application/json"}}`
    - Validasyon: `{"success": false, "error": {"code": "validation_error", "message": "username field is required"}}`
    - Bulunamadı: `{"success": false, "error": {"code": "not_found", "message": "Post not found"}}`

## Özellikler

- Her kullanıcı bir postu sadece bir kez oylayabilir
- Aynı posta farklı oy türü gönderilirse, önceki oy iptal edilir
- Postlar oluşturulma tarihine göre sıralanır (yeni en üstte)

## Örnek İstekler (curl)

```bash
# Post oluşturma (JSON)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"author":"melih","content":"merhaba"}' \
  http://localhost:8080/posts

# Postları listeleme
curl http://localhost:8080/posts

# Upvote (kullanıcı: ali)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"id":1, "like": true, "username": "ali"}' \
  http://localhost:8080/vote

# Downvote (kullanıcı: ali)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"id":1, "like": false, "username": "ali"}' \
  http://localhost:8080/vote

# Aynı kullanıcı aynı posta tekrar upvote (önceki oy iptal olur)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"id":1, "like": true, "username": "ali"}' \
  http://localhost:8080/vote
```
