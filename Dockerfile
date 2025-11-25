# Single stage build - avoids pulling multiple base images
FROM golang:1.23-alpine
WORKDIR /app

# Install ca-certificates
RUN apk --no-cache add ca-certificates

# Copy everything first
COPY . .

# Clean and rebuild go.mod/go.sum for Go 1.23 compatibility
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

EXPOSE 8080
CMD ["./main"]