FROM alpine:latest

# Install minimal dependencies
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy pre-built binary and resources
COPY ./server /app/server
COPY ./schema.sql /app/schema.sql
COPY ./docker-entrypoint.sh /app/docker-entrypoint.sh

# Make binary and entrypoint executable
RUN chmod +x /app/server /app/docker-entrypoint.sh

# Expose port
EXPOSE 8080

# Use the entrypoint script
ENTRYPOINT ["/app/docker-entrypoint.sh"]