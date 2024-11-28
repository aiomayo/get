.PHONY: build clean lint

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags="-X 'main.version=$(VERSION)'"

build:
	go build $(LDFLAGS) -o bin/get ./cmd/get


clean:
	rm -rf bin/
	go clean -modcache

lint:
	golangci-lint run

run:
	go run ./cmd/get