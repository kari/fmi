# Build configuration
BINARY_NAME=saa
VERSION?=0.1.0
CMD_PATH=./cmd/weather

# Output directories
DIST_DIR=dist
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Supported platforms for cross-compilation
PLATFORMS?=linux/amd64 darwin/arm64 windows/amd64

.PHONY: all clean test build dist vet dev

all: build

# Build the binary for the local platform
build: vet
	@echo "Building $(BINARY_NAME) for local platform..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(LDFLAGS) $(CMD_PATH)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run static analysis
vet:
	@echo "Vetting code..."
	$(GOVET) ./...

# Cross-compile for all platforms
dist: clean vet test
	@echo "Cross-compiling for platforms: $(PLATFORMS)"
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d"/" -f1); \
		GOARCH=$$(echo $$platform | cut -d"/" -f2); \
		output_name=$(DIST_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then output_name=$$output_name.exe; fi; \
		echo "--> Building for $$GOOS/$$GOARCH..."; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) -o $$output_name $(LDFLAGS) $(CMD_PATH); \
	done

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	$(GOCLEAN)

# Development target: build and run
dev: build
	@echo "Running development binary..."
	./$(BUILD_DIR)/$(BINARY_NAME)
