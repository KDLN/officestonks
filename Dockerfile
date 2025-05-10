# Use a minimal golang alpine image to build the app
FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
COPY schema.sql ./schema.sql

# Build a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -ldflags '-extldflags "-static"' -o /app/bin/api ./cmd/api/main.go

# Use a tiny alpine image for the final container
FROM alpine:3.16

# Install CA certificates and MySQL client for database operations
RUN apk --no-cache add ca-certificates mysql-client

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/bin/api /app/api

# Copy schema for initialization
COPY schema.sql /app/schema.sql
COPY start.sh /app/start.sh

# Make files executable
RUN chmod +x /app/api /app/start.sh

# Expose port
EXPOSE 8080

# Run the start script
ENTRYPOINT ["/app/start.sh"]