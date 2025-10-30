.PHONY: run build lint fmt tidy

run:
	go run ./cmd/server

build:
	mkdir -p bin
	go build -o bin/sse-go ./cmd/server

fmt:
	go fmt ./...

tidy:
	go mod tidy

lint:
	go vet ./...
