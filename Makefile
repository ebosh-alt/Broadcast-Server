SHELL := /bin/bash

APP          ?= broadcast-server
BIN_DIR      ?= bin
BIN          ?= $(BIN_DIR)/$(APP)
PKG          ?= ./...
CMD          ?= ./cmd/broadcast-server
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT       ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE         ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
  -X 'main.version=$(VERSION)' \
  -X 'main.commit=$(COMMIT)' \
  -X 'main.buildDate=$(DATE)'

GOFLAGS   ?=
TESTFLAGS ?= -race -covermode=atomic -coverprofile=coverage.out

.PHONY: all
all: fmt vet lint test build

.PHONY: fmt
fmt:
	go fmt $(PKG)

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test $(TESTFLAGS) $(PKG)
	@echo "Coverage report: coverage.out"
	@go tool cover -func=coverage.out | tail -n1

.PHONY: bench
bench:
	go test -bench=. -benchmem $(PKG)

.PHONY: vulncheck
vulncheck:
	govulncheck $(PKG)

# --- Run (local) ---
.PHONY: run-server
run-server:
	go run $(GOFLAGS) $(CMD) start --port 8080

.PHONY: run-client
run-client:
	go run $(GOFLAGS) $(CMD) connect --addr ws://localhost:8080/ws --name alice

.PHONY: build
build:
	mkdir -p $(BIN_DIR)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) $(CMD)

IMAGE ?= $(APP):$(VERSION)

.PHONY: docker-build
docker-build:
	docker build -f build/Dockerfile -t $(IMAGE) .

.PHONY: compose-up
compose-up:
	docker compose up --build

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) coverage.out

.PHONY: docker-build
docker-build:
	docker build -f build/Dockerfile --target final -t $(IMAGE) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) .

.PHONY: compose-up
compose-up:
	VERSION=$(VERSION) docker compose up --build

.PHONY: compose-up-d
compose-up-d:
	VERSION=$(VERSION) docker compose up --build -d

.PHONY: compose-up-alpine
compose-up-alpine:
	VERSION=$(VERSION) docker compose --profile alpine up --build

.PHONY: compose-up-alpine-d
compose-up-alpine-d:
	VERSION=$(VERSION) docker compose --profile alpine up --build -d

.PHONY: compose-down
compose-down:
	docker compose down -v --remove-orphans

.PHONY: logs
logs:
	docker compose logs -f --tail=200

.PHONY: health
health:
	@which curl >/dev/null 2>&1 && curl -fsS http://127.0.0.1:8080/healthz && echo || \
	echo "curl not found"


.PHONY: run
run: run-server

.PHONY: run-server
run-server:
	go run $(GOFLAGS) ./cmd/broadcast-server start --port 8080

.PHONY: run-client
run-client:
	go run $(GOFLAGS) ./cmd/broadcast-server connect --addr ws://localhost:8080/ws --name alice
