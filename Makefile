build:
	go build -o tickets ./cmd/tickets

test:
	go test ./... -race -cover

lint:
	golangci-lint run

install:
	go install ./cmd/tickets

tidy:
	go mod tidy

.PHONY: build test lint install tidy
