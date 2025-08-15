.PHONY: full clean lint lint-go fix fix-go test test-go build build-go dev-go watch watch-go docker-build docker-test docs-go

SHELL=/bin/bash -o pipefail
$(shell git config core.hooksPath ops/git-hooks)
GO_PATH := $(shell go env GOPATH 2> /dev/null)
PATH := /usr/local/bin:$(GO_PATH)/bin:$(PATH)

full: clean lint test build

## Clean the project of temporary files
clean:
	git clean -Xdff --exclude="!.env.local" --exclude="!.env.*.local"

## Lint the project
lint: lint-go

lint-go:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.5
	go mod tidy
	golangci-lint run ./...

## Fix the project
fix: fix-go

fix-go:
	go mod tidy
	gofmt -s -w .

## Test the project
test: test-go

test-go:
	@go install github.com/boumenot/gocover-cobertura@latest
	@mkdir -p tmp/coverage/go/
	go test -cover -coverprofile tmp/coverage/go/profile.txt ./...
	@go tool cover -func tmp/coverage/go/profile.txt | awk '/^total/{print $$1 " " $$3}'
	@go tool cover -html tmp/coverage/go/profile.txt -o tmp/coverage/go/coverage.html
	@gocover-cobertura < tmp/coverage/go/profile.txt > tmp/coverage/go/cobertura-coverage.xml

## Build the project
build: build-go

build-go:
	go generate
	go build -ldflags='-s -w' -o tmp/build/poseidon .
	go install .

dev-go:
	go run . | jq

## Watch the project
watch:
	hera

watch-go:
	hera backend

docker-build:
	docker build -t 'ghcr.io/lunagic/poseidon:latest' .

docker-test: docker-build
	docker run --rm -it -p 3000:3000/tcp ghcr.io/lunagic/poseidon:latest

## Run the docs server for the project
docs-go:
	@go install golang.org/x/tools/cmd/godoc@latest
	@echo "listening on http://127.0.0.1:6060/pkg/github.com/lunagic/poseidon"
	@godoc -http=127.0.0.1:6060
