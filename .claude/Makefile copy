# ONVIF GO Library Makefile

.PHONY: all build test clean install deps lint fmt vet check examples cli docker

# Configuration
BINARY_DIR := bin
GOPATH := $(shell go env GOPATH)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Binaries
CLI_BINARY := $(BINARY_DIR)/onvif-cli
QUICK_BINARY := $(BINARY_DIR)/onvif-quick

# Build all targets
all: deps check test build

# Build all binaries
build: $(CLI_BINARY) $(QUICK_BINARY)

# Build CLI tool (comprehensive)
$(CLI_BINARY):
	@echo "🔨 Building ONVIF CLI..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 go build -o $(CLI_BINARY) ./cmd/onvif-cli

# Build quick tool (simple)
$(QUICK_BINARY):
	@echo "🔨 Building ONVIF Quick Tool..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 go build -o $(QUICK_BINARY) ./cmd/onvif-quick

# Install binaries to GOPATH
install: build
	@echo "📦 Installing binaries..."
	cp $(CLI_BINARY) $(GOPATH)/bin/
	cp $(QUICK_BINARY) $(GOPATH)/bin/

# Download dependencies
deps:
	@echo "📥 Downloading dependencies..."
	go mod download
	go mod tidy

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	@echo "📊 Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run benchmarks
bench:
	@echo "⚡ Running benchmarks..."
	go test -bench=. -benchmem ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Vet code
vet:
	@echo "🔬 Vetting code..."
	go vet ./...

# Run all checks
check: fmt vet lint

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html

# Build examples
examples:
	@echo "📚 Building examples..."
	@mkdir -p $(BINARY_DIR)/examples
	go build -o $(BINARY_DIR)/examples/discovery ./examples/discovery
	go build -o $(BINARY_DIR)/examples/device_info ./examples/device_info
	go build -o $(BINARY_DIR)/examples/media ./examples/media
	go build -o $(BINARY_DIR)/examples/ptz ./examples/ptz

# Build for multiple platforms
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION)"

build-all:
	@echo "🌍 Building for multiple platforms (version: $(VERSION))..."
	@mkdir -p $(BINARY_DIR)
	
	# Linux AMD64
	@echo "Building Linux AMD64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-cli-linux-amd64 ./cmd/onvif-cli
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-quick-linux-amd64 ./cmd/onvif-quick
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-server-linux-amd64 ./cmd/onvif-server
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-diagnostics-linux-amd64 ./cmd/onvif-diagnostics
	
	# Linux ARM64
	@echo "Building Linux ARM64..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-cli-linux-arm64 ./cmd/onvif-cli
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-quick-linux-arm64 ./cmd/onvif-quick
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-server-linux-arm64 ./cmd/onvif-server
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-diagnostics-linux-arm64 ./cmd/onvif-diagnostics
	
	# Linux ARM (32-bit)
	@echo "Building Linux ARM..."
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-cli-linux-arm ./cmd/onvif-cli
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-quick-linux-arm ./cmd/onvif-quick
	
	# Windows AMD64
	@echo "Building Windows AMD64..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-cli-windows-amd64.exe ./cmd/onvif-cli
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-quick-windows-amd64.exe ./cmd/onvif-quick
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-server-windows-amd64.exe ./cmd/onvif-server
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-diagnostics-windows-amd64.exe ./cmd/onvif-diagnostics
	
	# Windows ARM64
	@echo "Building Windows ARM64..."
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-cli-windows-arm64.exe ./cmd/onvif-cli
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-quick-windows-arm64.exe ./cmd/onvif-quick
	
	# macOS AMD64 (Intel)
	@echo "Building macOS AMD64..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-cli-darwin-amd64 ./cmd/onvif-cli
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-quick-darwin-amd64 ./cmd/onvif-quick
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-server-darwin-amd64 ./cmd/onvif-server
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-diagnostics-darwin-amd64 ./cmd/onvif-diagnostics
	
	# macOS ARM64 (Apple Silicon)
	@echo "Building macOS ARM64..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-cli-darwin-arm64 ./cmd/onvif-cli
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-quick-darwin-arm64 ./cmd/onvif-quick
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-server-darwin-arm64 ./cmd/onvif-server
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_DIR)/onvif-diagnostics-darwin-arm64 ./cmd/onvif-diagnostics
	
	@echo "✅ All binaries built successfully in $(BINARY_DIR)/"
	@echo ""
	@ls -lh $(BINARY_DIR)/

# Create release archives with checksums
release: build-all
	@echo "📦 Creating release archives..."
	@mkdir -p releases
	
	# Create archives for each platform
	@cd $(BINARY_DIR) && \
	for os in linux darwin windows; do \
		for arch in amd64 arm64 arm; do \
			if [ "$$os" = "windows" ] && [ "$$arch" != "arm" ]; then \
				if [ -f onvif-cli-$$os-$$arch.exe ]; then \
					zip -j ../releases/onvif-go-$(VERSION)-$$os-$$arch.zip onvif-*-$$os-$$arch.exe ../README.md ../LICENSE 2>/dev/null || true; \
				fi; \
			elif [ "$$os" != "windows" ]; then \
				if [ -f onvif-cli-$$os-$$arch ]; then \
					tar czf ../releases/onvif-go-$(VERSION)-$$os-$$arch.tar.gz onvif-*-$$os-$$arch ../README.md ../LICENSE 2>/dev/null || true; \
				fi; \
			fi; \
		done; \
	done
	
	# Generate checksums
	@cd releases && sha256sum * > checksums.txt 2>/dev/null || shasum -a 256 * > checksums.txt
	@echo "✅ Release archives created in releases/"
	@ls -lh releases/

# Create Docker image
docker:
	@echo "🐳 Building Docker image..."
	docker build -t onvif-go:latest .

# Development setup
dev-setup:
	@echo "🛠️  Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go mod download

# Run quick tool
run-quick:
	@if [ ! -f $(QUICK_BINARY) ]; then $(MAKE) $(QUICK_BINARY); fi
	$(QUICK_BINARY)

# Run CLI tool
run-cli:
	@if [ ! -f $(CLI_BINARY) ]; then $(MAKE) $(CLI_BINARY); fi
	$(CLI_BINARY)

# Show help
help:
	@echo "📖 Available targets:"
	@echo "  all          - Build, test, and check everything"
	@echo "  build        - Build both CLI tools"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  bench        - Run benchmarks"
	@echo "  check        - Run fmt, vet, and lint"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install binaries to GOPATH"
	@echo "  examples     - Build example programs"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  docker       - Build Docker image"
	@echo "  dev-setup    - Set up development environment"
	@echo "  run-quick    - Run the quick tool"
	@echo "  run-cli      - Run the comprehensive CLI"
	@echo "  help         - Show this help"