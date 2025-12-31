.PHONY: tidy test test-race lint build cover

tidy:
	go mod tidy

test:
	go test -v ./...

# Кроссплатформенный тест с race detector
test-race:
	go test -race ./...

lint:
	golangci-lint run ./...

build:
	go build -o bin/urlshortener ./cmd/urlshortener

cover:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out