FROM golang:1.20 AS builder

WORKDIR /app

# Copy the entire backend directory
COPY backend/ /app/backend/

# Build the application - using CGO_ENABLED=0 for static binary
WORKDIR /app/backend
RUN go mod tidy && go mod download
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
COPY init-db.sh /app/init-db.sh
COPY backend/schema.sql /app/schema.sql

# Make the files executable
RUN chmod +x /app/bin/server /app/start-server.sh /app/init-db.sh

EXPOSE 8080

# Run the start script
ENTRYPOINT ["/app/start-server.sh"]