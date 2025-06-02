# Stage 1: Build the Go application
FROM golang:1.22.0-alpine AS builder

# Set environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Create working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o app ./cmd/main.go

# Stage 2: Minimal runtime container
FROM alpine:latest

# Create working directory
WORKDIR /root/

# Copy built binary and config file
COPY --from=builder /app/app .
COPY config.env .

# Expose the app port (if your app listens on 5000, adjust accordingly)
EXPOSE 5000

# Set entrypoint
CMD ["./app"]
