.PHONY: fmt vet test build run tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./... -count=1

tidy:
	go mod tidy

build:
	go build -o bin/neurofeed ./cmd/neurofeed

run:
	go run ./cmd/neurofeed

all: fmt vet test build
