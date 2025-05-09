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

# Copy the binary from the builder stage
COPY --from=builder /app/bin/server /app/bin/server

# Make the binary executable
RUN chmod +x /app/bin/server

EXPOSE 8080

# Run the binary directly
CMD ["/app/bin/server"]