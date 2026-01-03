# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

COPY ui/package*.json ./
RUN npm ci

COPY ui/ ./
RUN npm run build

# Runtime stage
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY ui/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
