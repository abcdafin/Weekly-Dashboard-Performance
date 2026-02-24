# ===== Stage 1: Build =====
FROM node:22-alpine AS builder

WORKDIR /app

# Copy package files first for better layer caching
COPY package.json package-lock.json ./
RUN npm ci

# Copy source code
COPY . .

# Set API URL to use relative path (Nginx will proxy to backend)
ENV VITE_API_URL=/api/v1

# Build for production
RUN npm run build

# ===== Stage 2: Runtime =====
FROM nginx:alpine

# Remove default Nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom Nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built frontend from builder
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
