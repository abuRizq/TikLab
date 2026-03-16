# TikLab Sandbox Beta - Build and development targets
# See specs/001-tiklab-sandbox-beta/plan.md for project structure

.PHONY: build build-engine build-image test lint cross-build dist
.PHONY: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-arm64

# Version injected at build time via -ldflags
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

LDFLAGS = -ldflags "-X main.Version=$(VERSION)"

# Output directory for cross-platform binaries
DIST_DIR = dist

# Build targets
build:
	go build $(LDFLAGS) -o tiklab ./cmd/tiklab

build-engine:
	go build $(LDFLAGS) -o tiklab-engine ./cmd/tiklab-engine

build-image:
	docker build --build-arg VERSION=$(VERSION) -t tiklab/sandbox:$(VERSION) -t tiklab/sandbox:latest -f build/Dockerfile .

# Test and lint
test:
	go test ./...

# Integration tests (requires Docker, tiklab/sandbox image)
test-integration:
	go test -tags=integration -count=1 -v ./tests/integration/... -timeout 10m

lint:
	golangci-lint run

# Cross-compilation targets — produce tiklab-{os}-{arch} in dist/
$(DIST_DIR):
	mkdir -p $(DIST_DIR)

build-linux-amd64: $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/tiklab-linux-amd64 ./cmd/tiklab
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/tiklab-engine-linux-amd64 ./cmd/tiklab-engine

build-linux-arm64: $(DIST_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/tiklab-linux-arm64 ./cmd/tiklab
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/tiklab-engine-linux-arm64 ./cmd/tiklab-engine

build-darwin-amd64: $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/tiklab-darwin-amd64 ./cmd/tiklab
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/tiklab-engine-darwin-amd64 ./cmd/tiklab-engine

build-darwin-arm64: $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/tiklab-darwin-arm64 ./cmd/tiklab
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/tiklab-engine-darwin-arm64 ./cmd/tiklab-engine

build-windows-amd64: $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/tiklab-windows-amd64.exe ./cmd/tiklab
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/tiklab-engine-windows-amd64.exe ./cmd/tiklab-engine

build-windows-arm64: $(DIST_DIR)
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/tiklab-windows-arm64.exe ./cmd/tiklab
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/tiklab-engine-windows-arm64.exe ./cmd/tiklab-engine

# Build all cross-compiled binaries to dist/
dist: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-arm64

# Legacy cross-build (outputs to current directory)
cross-build: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-arm64
