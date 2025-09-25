CMD          ?= ./cmd/broadcast-server
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT       ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE         ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
  -X 'main.version=$(VERSION)' \
  -X 'main.commit=$(COMMIT)' \
  -X 'main.buildDate=$(DATE)'


.PHONY: run-server
run-server:
	go run $(CMD) start --port 8080
