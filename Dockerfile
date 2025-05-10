FROM golang:1.20 AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy && go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY schema.sql ./schema.sql

# Add debugging to see the environment
RUN pwd && ls -la && ls -la cmd/api/

# Build with more verbosity to see any issues
RUN CGO_ENABLED=0 GOOS=linux go build -v -a -ldflags '-extldflags "-static"' -o /app/bin/server ./cmd/api/main.go

# Use a small alpine image for the final container
FROM alpine:latest

WORKDIR /app

# Add CA certificates for HTTPS calls and MySQL client for database initialization
RUN apk --no-cache add ca-certificates mysql-client

# Copy the binary and scripts
COPY --from=builder /app/bin/server /app/bin/server
COPY start-server.sh /app/start-server.sh
COPY start.sh /app/start.sh
COPY init-db.sh /app/init-db.sh
COPY schema.sql /app/schema.sql

# Make the files executable
RUN chmod +x /app/bin/server /app/start-server.sh /app/start.sh /app/init-db.sh

EXPOSE 8080

# Run the start script
ENTRYPOINT ["/app/start.sh"]