# Build configuration
BINARY_NAME=fmi
VERSION?=0.1.0
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

# Supported platforms
PLATFORMS=linux/amd64 darwin/arm64

.PHONY: all clean test build dist vet

all: clean test build

clean:
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	$(GOCLEAN)

test:
	$(GOTEST) -v ./...

vet:
	$(GOVET) ./...

build: clean
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(LDFLAGS)

dist: clean test
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d"/" -f1) \
		GOARCH=$$(echo $$platform | cut -d"/" -f2) \
		output_name=$(DIST_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then output_name=$$output_name.exe; fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) -o $$output_name $(LDFLAGS) ./cmd/sää; \
	done

# Development targets
dev: build
	./$(BUILD_DIR)/$(BINARY_NAME)
