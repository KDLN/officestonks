FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY backend/go.mod backend/go.sum /app/backend/
RUN cd backend && go mod download

# Copy backend source code
COPY backend/ /app/backend/

# Build the application
RUN cd backend && go build -o /app/bin/api cmd/api/main.go

# Create a smaller final image
FROM alpine:3.17

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/api /app/bin/api

# Make the binary executable
RUN chmod +x /app/bin/api

EXPOSE 8080

# Run the binary
CMD ["/app/bin/api"]