FROM golang:1.22-alpine

WORKDIR /app
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org
RUN go mod download

COPY . .

# If you rely on gqlgen codegen to produce graph/generated.go before build,
# run `gqlgen generate` locally; the Docker build here will compile source as-is.
RUN go build -o server ./...

EXPOSE 8080
CMD ["./server"]
