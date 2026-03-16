# TikLab Sandbox Beta - Build and development targets
# See specs/001-tiklab-sandbox-beta/plan.md for project structure

.PHONY: build build-engine build-image test lint cross-build
.PHONY: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-arm64

# Version injected at build time via -ldflags
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

LDFLAGS = -ldflags "-X main.Version=$(VERSION)"

# Build targets
build:
	go build $(LDFLAGS) -o tiklab ./cmd/tiklab

build-engine:
	go build $(LDFLAGS) -o tiklab-engine ./cmd/tiklab-engine

build-image:
	docker build -t tiklab/sandbox:$(VERSION) -f build/Dockerfile .

# Test and lint
test:
	go test ./...

lint:
	golangci-lint run

# Cross-compilation targets (GOOS=linux,darwin,windows GOARCH=amd64,arm64)
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o tiklab-linux-amd64 ./cmd/tiklab
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o tiklab-engine-linux-amd64 ./cmd/tiklab-engine

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o tiklab-linux-arm64 ./cmd/tiklab
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o tiklab-engine-linux-arm64 ./cmd/tiklab-engine

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o tiklab-darwin-amd64 ./cmd/tiklab
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o tiklab-engine-darwin-amd64 ./cmd/tiklab-engine

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o tiklab-darwin-arm64 ./cmd/tiklab
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o tiklab-engine-darwin-arm64 ./cmd/tiklab-engine

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o tiklab-windows-amd64.exe ./cmd/tiklab
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o tiklab-engine-windows-amd64.exe ./cmd/tiklab-engine

build-windows-arm64:
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o tiklab-windows-arm64.exe ./cmd/tiklab
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o tiklab-engine-windows-arm64.exe ./cmd/tiklab-engine

# Build all cross-compiled binaries
cross-build: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-arm64
