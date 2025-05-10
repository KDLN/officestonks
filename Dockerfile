FROM alpine:latest

# Install minimal dependencies
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy pre-built binary and resources
COPY ./server /app/server
COPY ./schema.sql /app/schema.sql

# Make binary executable
RUN chmod +x /app/server

# Expose port
EXPOSE 8080

# Run the server directly
CMD ["/app/server"]