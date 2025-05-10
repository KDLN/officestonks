# Use a minimal golang alpine image to build the app
FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Set environment variables for Go
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

# Check go.mod content
RUN cat go.mod

# Download dependencies with verbose output
RUN go mod tidy -v && go mod download -x

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY schema.sql ./schema.sql

# Add debugging to see the environment
RUN pwd && ls -la && ls -la cmd/api/

# Show environment and list files before building
RUN echo "Current directory: $(pwd)" && ls -la

# Check Go version and environment
RUN go version && go env

# List cmd/api contents to verify files exist
RUN ls -la cmd/api/

# Build with detailed debugging
RUN go build -v -x -a -ldflags '-extldflags "-static"' -o /app/bin/server ./cmd/api/main.go || \
    (echo "===== BUILD FAILED ====="; \
     echo "Directory structure:"; \
     find . -type f -name "*.go" | sort; \
     echo "===== import errors ====="; \
     go build -v ./cmd/api/main.go 2>&1 | grep "import"; \
     echo "===== main.go content ====="; \
     cat cmd/api/main.go; \
     exit 1)

# Use a tiny alpine image for the final container
FROM alpine:3.16

# Install CA certificates and MySQL client for database operations
RUN apk --no-cache add ca-certificates mysql-client

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/bin/server /app/bin/server

# Copy schema and scripts - using absolute paths to be explicit
COPY ./schema.sql /app/schema.sql
COPY ./start.sh /app/start.sh
COPY ./start-server.sh /app/start-server.sh
COPY ./init-db.sh /app/init-db.sh

# Debug: list files to verify they exist
RUN ls -la /app/

# Make files executable
RUN chmod +x /app/bin/server /app/start.sh /app/start-server.sh /app/init-db.sh

# Debug: verify permissions
RUN ls -la /app/

# Expose port
EXPOSE 8080

# Run the start script
ENTRYPOINT ["/app/start.sh"]