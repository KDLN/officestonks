FROM golang:1.19

WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the codebase
COPY . .

# Build the application
RUN go build -o /app/bin/server ./cmd/api

# Expose the port the server listens on
EXPOSE 8080

# Command to run the executable
CMD ["/app/bin/server"]