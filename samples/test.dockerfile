# --------------------
# Build stage
# --------------------
FROM node:20-alpine AS build

# Set working directory
WORKDIR /app

# Copy dependency files first (better caching)
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy the rest of the source code
COPY . .

# Build the app
RUN npm run build


# --------------------
# Production stage
# --------------------
FROM nginx:alpine

# Copy built files from build stage
COPY --from=build /app/dist /usr/share/nginx/html

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
