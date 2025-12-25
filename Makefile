.PHONY: build install clean test

VERSION := 0.2.0
BUILD_DIR := build
BINARY := keel

# Build for current platform
build:
	go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY) ./cmd/keel

# Install to GOPATH/bin
install:
	go install -ldflags="-X main.version=$(VERSION)" ./cmd/keel

# Cross-compile for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/keel
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/keel
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/keel
	GOOS=linux GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/keel
	GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/keel

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run development version
dev:
	go run ./cmd/keel $(ARGS)
