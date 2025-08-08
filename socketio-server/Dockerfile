FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy server files
COPY . .

# Expose port (Railway will override with PORT env var)
EXPOSE 3001

# Start the server
CMD ["npm", "start"]