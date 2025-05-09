FROM node:18-alpine as build

WORKDIR /app

# Copy package.json and install dependencies
COPY package.json ./
# Force install without using package-lock to resolve dependency conflicts
RUN npm install

# Copy the rest of the application code
COPY . ./

# Build the application for production
ENV REACT_APP_API_URL=https://web-production-1e26.up.railway.app
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built assets from the build stage
COPY --from=build /app/build /usr/share/nginx/html

# Copy nginx configuration and startup script
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY start.sh /start.sh

# Make startup script executable
RUN chmod +x /start.sh

# Create a health check file
RUN echo "OK" > /usr/share/nginx/html/health

# Expose port 8080 (Railway expects this port by default)
EXPOSE 8080

# Add healthcheck
HEALTHCHECK --interval=5s --timeout=3s --retries=3 CMD curl -f http://localhost:8080/health || exit 1

# Use startup script as entrypoint
ENTRYPOINT ["/start.sh"]