# Makefile for Git-360

# Variables
BINARY_NAME=git360
MAIN_PACKAGE=./cmd/git360
GO_FILES=$(shell find . -name "*.go" -type f)

.PHONY: all build run test clean fmt vet graph-update help

# Default target
all: build

## build: Compile the binary
build: $(BINARY_NAME)

$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

## run: Compile and run the TUI
run: build
	./$(BINARY_NAME)

## test: Run unit tests
test:
	@echo "Running tests..."
	go test -v ./...

## clean: Remove compiled binary
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)

## fmt: Format codebase source files
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

## vet: Run go vet analyzer
vet:
	@echo "Vetting Go code..."
	go vet ./...

## graph-update: Update code knowledge graph using graphify
graph-update:
	@echo "Updating codebase knowledge graph..."
	graphify update .

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
