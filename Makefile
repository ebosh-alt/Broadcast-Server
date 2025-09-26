CMD          ?= ./cmd/broadcast-server

run-server:
	go run $(CMD) start --port 8080
