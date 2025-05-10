FROM debian:bullseye-slim

# Install necessary tools
RUN apt-get update && apt-get install -y ca-certificates default-mysql-client && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy pre-built binary and required files
COPY ./server /app/server
COPY ./schema.sql /app/schema.sql
COPY ./start.sh /app/start.sh
COPY ./start-server.sh /app/start-server.sh
COPY ./init-db.sh /app/init-db.sh

# Make all scripts executable
RUN chmod +x /app/server /app/start.sh /app/start-server.sh /app/init-db.sh

# Expose port
EXPOSE 8080

# Run the server directly
CMD ["/app/server"]