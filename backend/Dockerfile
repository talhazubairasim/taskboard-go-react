# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# GraphQL code is generated and committed (make generate), so the build is reproducible.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# Runtime stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata wget \
 && adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=builder /out/server /out/migrate ./
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -q -O /dev/null http://localhost:8080/health || exit 1
# The server applies pending migrations on startup.
CMD ["./server"]
