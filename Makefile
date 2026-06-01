VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/krzhapps/GitTickets/internal/cli.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o tickets ./cmd/tickets

test:
	go test ./... -race -cover

lint:
	golangci-lint run

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/tickets

tidy:
	go mod tidy

.PHONY: build test lint install tidy
