FROM golang:1.20 AS builder

WORKDIR /app

# Copy the entire backend directory
COPY backend/ /app/backend/

# Build the application
WORKDIR /app/backend
RUN go mod tidy && go mod download
RUN go build -o /app/bin/server cmd/api/main.go

# Create a smaller final image
FROM debian:bullseye-slim

WORKDIR /app

# Copy the binary and start script
COPY --from=builder /app/bin/server /app/bin/server
COPY start-server.sh /app/start-server.sh

# Make the files executable
RUN chmod +x /app/bin/server /app/start-server.sh

EXPOSE 8080

# Run the binary directly
ENTRYPOINT ["/app/start-server.sh"]