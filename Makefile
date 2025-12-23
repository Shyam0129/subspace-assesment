.PHONY: help build run clean test deps install

# Default target
help:
	@echo "LinkedIn Automation - Makefile Commands"
	@echo ""
	@echo "Available targets:"
	@echo "  make deps      - Download Go dependencies"
	@echo "  make build     - Build the application"
	@echo "  make run       - Run the application"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make test      - Run tests"
	@echo "  make install   - Install the binary to GOPATH/bin"
	@echo ""

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

# Build the application
build: deps
	@echo "Building application..."
	go build -o linkedin-automation ./cmd/main.go

# Run the application
run: build
	@echo "Running application..."
	./linkedin-automation

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f linkedin-automation
	rm -rf data/
	rm -rf logs/

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Install binary
install: build
	@echo "Installing..."
	go install ./cmd/main.go

# Initialize project (first time setup)
init:
	@echo "Initializing project..."
	@if [ ! -f .env ]; then cp .env.example .env; echo "Created .env file - please configure it"; fi
	@mkdir -p data logs
	@echo "Project initialized!"
