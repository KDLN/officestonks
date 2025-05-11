FROM node:18-alpine

WORKDIR /app

# Copy package.json and package-lock.json
COPY package.json ./

# Install dependencies
RUN npm install

# Copy server code
COPY server.js ./

# Expose port
EXPOSE 8080

# Start the server
CMD ["npm", "start"]