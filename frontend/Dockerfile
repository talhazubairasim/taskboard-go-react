# Development stage (docker compose up)
FROM node:20-alpine AS development
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# Build stage
FROM node:20-alpine AS build
WORKDIR /app
ARG VITE_GRAPHQL_HTTP_URL=http://localhost:8080/query
ARG VITE_GRAPHQL_WS_URL=ws://localhost:8080/query
ENV VITE_GRAPHQL_HTTP_URL=$VITE_GRAPHQL_HTTP_URL \
    VITE_GRAPHQL_WS_URL=$VITE_GRAPHQL_WS_URL
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Production stage
FROM nginx:1.27-alpine AS production
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
