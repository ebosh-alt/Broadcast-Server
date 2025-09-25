FROM golang:1.22-alpine AS builder

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=1970-01-01T00:00:00Z

RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -buildvcs=false \
      -ldflags "-s -w \
        -X 'main.version=${VERSION}' \
        -X 'main.commit=${COMMIT}' \
        -X 'main.buildDate=${DATE}'" \
      -o /out/broadcast-server ./cmd/broadcast-server

FROM gcr.io/distroless/base-debian12:nonroot AS final

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /
COPY --from=builder /out/broadcast-server /broadcast-server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/broadcast-server"]
CMD ["start","--port","8080"]


FROM alpine:3.20 AS alpine

RUN apk add --no-cache ca-certificates tzdata curl
WORKDIR /
COPY --from=builder /out/broadcast-server /broadcast-server
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=2s --retries=3 \
  CMD curl -fsS http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/broadcast-server"]
CMD ["start","--port","8080"]
