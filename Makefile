.PHONY: full clean lint lint-go fix fix-go test test-go build build-go watch docker-build docker-test docs-go

SHELL=/bin/bash -o pipefail
$(shell git config core.hooksPath ops/git-hooks)
PROJECT_NAME := $(shell basename $(CURDIR))
GO_MODULE := $(shell grep "^module " go.mod | awk '{print $$2}')
GO_PATH := $(shell go env GOPATH 2> /dev/null)
PATH := $(GO_PATH)/bin:$(PATH)

full: clean lint test build

## Clean the project of temporary files
clean:
	git clean -Xdff --exclude="!.env*local"

## Lint the project
lint: lint-go

lint-go:
	go get ./...
	go mod tidy
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55

## Fix the project
fix: fix-go

fix-go:
	go mod tidy
	gofmt -s -w .

## Test the project
test: test-go

test-go:
	@mkdir -p .stencil/tmp/coverage/go/
	@go install github.com/boumenot/gocover-cobertura@latest
	go test -p 1 -count=1 -cover -coverprofile .stencil/tmp/coverage/go/profile.txt ./...
	@go tool cover -func .stencil/tmp/coverage/go/profile.txt | awk '/^total/{print $$1 " " $$3}'
	@go tool cover -html .stencil/tmp/coverage/go/profile.txt -o .stencil/tmp/coverage/go/coverage.html
	@gocover-cobertura < .stencil/tmp/coverage/go/profile.txt > .stencil/tmp/coverage/go/cobertura-coverage.xml

## Build the project
build: build-go

build-go:
	go generate
	go build -ldflags='-s -w' -o .stencil/tmp/build .
	go install .

## Watch the project
watch:

docker-build:
	docker build --pull --rm --platform linux/amd64,linux/arm64 -t local/poseidon:latest "."

docker-test: docker-build
	docker run -p 3000:3000/tcp -v ./poseidon/test_data/:/var/www/html/ local/poseidon:latest

docker-publish: docker-build
	docker tag local/poseidon:latest ghcr.io/aaronellington/poseidon:latest
	docker push ghcr.io/aaronellington/poseidon:latest

## Run the docs server for the project
docs-go:
	@go install golang.org/x/tools/cmd/godoc@latest
	@echo "listening on http://127.0.0.1:6060/pkg/${GO_MODULE}"
	@godoc -http=127.0.0.1:6060
