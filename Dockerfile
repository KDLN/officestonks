FROM alpine:latest

# Install minimal dependencies
RUN apk --no-cache add ca-certificates bash

WORKDIR /app

# Copy pre-built binary and resources
COPY ./server /app/server
COPY ./schema.sql /app/schema.sql

# Make binary executable
RUN chmod +x /app/server

# Expose port
EXPOSE 8080

# Create a simple start script that directly runs the server
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'echo "Starting OfficeStonks server..."' >> /app/start.sh && \
    echo 'echo "Working directory: $(pwd)"' >> /app/start.sh && \
    echo 'echo "Available files: $(ls -la)"' >> /app/start.sh && \
    echo 'echo "Server binary: $(which server || echo not found)"' >> /app/start.sh && \
    echo 'exec /app/server' >> /app/start.sh && \
    chmod +x /app/start.sh

# Use CMD instead of ENTRYPOINT
CMD ["/app/start.sh"]