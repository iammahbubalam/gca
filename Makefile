.PHONY: all build test lint clean install proto help deps verify

# Variables
BINARY_NAME=ghost-agent
CLI_NAME=agentctl
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_DIR=build
LDFLAGS=-ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## deps: Download dependencies
deps:
	go mod download
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

## proto: Generate protobuf code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/agentpb/agent.proto

## Build targets
## build: Build ghost-agent binary
build: proto
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/ghost-agent/main.go
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-cli: Build agentctl (CLI) binary
build-cli: proto
	@echo "Building $(CLI_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_NAME) cmd/agentctl/main.go
	@echo "✅ Build complete: $(BUILD_DIR)/$(CLI_NAME)"

## build-all-local: Build all binaries for the current platform
build-all-local: build build-cli
	@echo "✅ All local binaries built"

## build-all: Build for multiple platforms
build-all: proto
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 cmd/ghost-agent/main.go
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 cmd/ghost-agent/main.go

## test: Run all tests
test:
	go test -v -race -coverprofile=coverage.out ./...

## test-unit: Run unit tests only
test-unit:
	go test -v -short -race ./...

## test-integration: Run integration tests
test-integration:
	go test -v -run Integration ./test/integration/...

## coverage: Generate coverage report
coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linters
lint:
	golangci-lint run --timeout 5m

## fmt: Format code
fmt:
	go fmt ./...
	goimports -w .

## vet: Run go vet
vet:
	go vet ./...

## install: Install binaries
install: build
	sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	sudo install -m 755 $(BUILD_DIR)/$(CLI_NAME) /usr/local/bin/$(CLI_NAME)

## clean: Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## verify: Run all checks (lint, test)
verify: lint test
	@echo "All checks passed!"

all: clean verify build
