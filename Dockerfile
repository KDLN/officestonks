# Use Node.js to build the frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
# Pass environment variables for the build
ARG REACT_APP_SUPABASE_URL
ARG REACT_APP_SUPABASE_ANON_KEY
ARG REACT_APP_BACKEND_URL
ENV REACT_APP_SUPABASE_URL=$REACT_APP_SUPABASE_URL
ENV REACT_APP_SUPABASE_ANON_KEY=$REACT_APP_SUPABASE_ANON_KEY
ENV REACT_APP_BACKEND_URL=$REACT_APP_BACKEND_URL
RUN npm run build

# Use Go to build the backend
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o out ./cmd/api/main.go

# Final runtime image with Node.js for Socket.IO
FROM node:18-alpine
RUN apk --no-cache add ca-certificates bash netcat-openbsd
WORKDIR /app

# Copy Go backend
COPY --from=backend-builder /app/out .
COPY --from=frontend-builder /app/frontend/build ./frontend/build

# Copy Socket.IO server
COPY socketio-server/ ./socketio-server/

# Copy startup script
COPY start.sh .
RUN chmod +x start.sh

# Expose ports (Railway will use PORT env var for main service)
EXPOSE 8080

# Use startup script to run both services
CMD ["./start.sh"]