# Build Stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go

# Runtime Stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
# - git: required for cloning user repos
# - docker-cli: required for spawning sibling containers
# - ca-certificates: required for HTTPS requests (Gemini API, GitHub)
RUN apk add --no-cache git docker-cli ca-certificates

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/config ./config

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
