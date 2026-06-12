.PHONY: all build clean test lint install

# Build variables
BINARY_NAME=agent-insight
BUILD_DIR=.
GO_FLAGS=-v
LDFLAGS=-ldflags "-s -w"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# all runs build
all: build

# build builds the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(GO_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# install installs the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) $(GO_FLAGS) $(LDFLAGS) -o $$GOPATH/bin/$(BINARY_NAME) .

# clean removes build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

# test runs all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# lint runs golangci-lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# deps downloads dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# help displays this help message
help:
	@echo "Available targets:"
	@echo "  all      - Build the binary (default)"
	@echo "  build    - Build the binary to build/"
	@echo "  install  - Install the binary to \$GOPATH/bin"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run all tests"
	@echo "  lint     - Run linter"
	@echo "  deps     - Download and tidy dependencies"
	@echo "  help     - Show this help message"
