FROM golang:1.20 AS builder

WORKDIR /app

# Copy the entire backend directory
COPY backend/ /app/backend/

# Build the application - using CGO_ENABLED=0 for static binary
WORKDIR /app/backend
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o /app/bin/server cmd/api/main.go

# Use a small alpine image for the final container
FROM alpine:latest

WORKDIR /app

# Add CA certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Copy the binary and start script
COPY --from=builder /app/bin/server /app/bin/server
COPY start-server.sh /app/start-server.sh

# Make the files executable
RUN chmod +x /app/bin/server /app/start-server.sh

EXPOSE 8080

# Run the binary directly
ENTRYPOINT ["/app/start-server.sh"]