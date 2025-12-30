test:
	go mod tidy
	go test -v ./...

lint:
	golangci-lint run ./...

build:
	go build -o bin/gendiff ./cmd/gendiff

cover:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out