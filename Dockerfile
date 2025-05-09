FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go mod file first for better caching
COPY backend/go.mod /app/backend/
# Initialize go modules if no go.sum exists
RUN cd backend && go mod tidy && go mod download

# Copy backend source code
COPY backend/ /app/backend/

# Build the application
RUN cd backend && go build -o /app/bin/server cmd/api/main.go

# Create a smaller final image
FROM alpine:3.17

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/server /app/bin/server

# Make the binary executable
RUN chmod +x /app/bin/server

# Add bash to the image
RUN apk add --no-cache bash

# Copy the start script
COPY backend/start.sh /app/start.sh
RUN chmod +x /app/start.sh

EXPOSE 8080

# Run the server using the start script
CMD ["/app/start.sh"]