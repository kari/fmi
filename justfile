binary := "saa"
cmd := "./cmd/saa"
version := `git describe --tags --always --dirty`
version_flag := "-X main.Version=" + version
lint := `command -v golangci-lint 2>/dev/null || echo "$HOME/go/bin/golangci-lint"`

# List available recipes
default:
    @just --list

# Build saa for the local platform
build:
    mkdir -p build
    CGO_ENABLED=0 go build -ldflags "{{version_flag}}" -o build/{{binary}} {{cmd}}

# Run tests with the race detector
test:
    go test -race -v ./...

# Run linter (includes go vet)
lint:
    {{lint}} run ./...

# Auto-format code
fmt:
    {{lint}} fmt

# Run tests and linter, same as CI
check: test lint

# Run the CLI against the live API
run place="Helsinki":
    go run {{cmd}} {{place}}

# Cross-compile distribution binaries
dist platforms="linux/amd64 darwin/arm64 windows/amd64": clean check
    #!/usr/bin/env bash
    # Override platforms: just dist "linux/arm64 darwin/arm64"
    set -euo pipefail
    mkdir -p dist
    for platform in {{platforms}}; do
        goos="${platform%%/*}"
        goarch="${platform#*/}"
        out="dist/{{binary}}-${goos}-${goarch}"
        if [ "$goos" = "windows" ]; then out="${out}.exe"; fi
        echo "--> Building for ${goos}/${goarch}..."
        CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -ldflags "{{version_flag}}" -o "$out" {{cmd}}
    done

# Install saa into $GOBIN / ~/go/bin
install:
    go install -ldflags "{{version_flag}}" {{cmd}}

# Remove build artifacts
clean:
    rm -rf build dist
