FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -o /out/broadcast-server ./cmd/broadcast-server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/broadcast-server /app/broadcast-server
EXPOSE 8080
CMD ["/app/broadcast-server","-port","8080"]
