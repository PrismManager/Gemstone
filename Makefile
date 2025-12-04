.PHONY: build build-cli build-daemon install uninstall clean test fmt lint

# Variables
VERSION ?= 0.1.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Directories
DIST_DIR := dist
CONFIG_DIR := /etc/gemstone
DATA_DIR := /var/lib/gemstone
LOG_DIR := /var/log/gemstone
RUN_DIR := /run/gemstone

# Build targets
all: build

build: build-cli build-daemon

build-cli:
	@echo "Building gem CLI..."
	@mkdir -p $(DIST_DIR)
	go build $(LDFLAGS) -o $(DIST_DIR)/gem ./cmd/gem

build-daemon:
	@echo "Building gemstoned daemon..."
	@mkdir -p $(DIST_DIR)
	go build $(LDFLAGS) -o $(DIST_DIR)/gemstoned ./cmd/gemstoned

# Development builds
dev: 
	go build -o $(DIST_DIR)/gem ./cmd/gem
	go build -o $(DIST_DIR)/gemstoned ./cmd/gemstoned

# Install (requires root)
install: build
	@echo "Installing gemstone..."
	@install -d $(CONFIG_DIR)
	@install -d $(DATA_DIR)
	@install -d $(LOG_DIR)
	@install -m 755 $(DIST_DIR)/gem /usr/local/bin/gem
	@install -m 755 $(DIST_DIR)/gemstoned /usr/local/bin/gemstoned
	@if [ ! -f $(CONFIG_DIR)/config.yaml ]; then \
		install -m 644 configs/config.yaml $(CONFIG_DIR)/config.yaml; \
	fi
	@install -m 644 init/gemstone.service /etc/systemd/system/gemstone.service
	@systemctl daemon-reload
	@echo "Gemstone installed successfully!"
	@echo ""
	@echo "To start the daemon:"
	@echo "  sudo systemctl start gemstone"
	@echo ""
	@echo "To enable auto-start on boot:"
	@echo "  sudo systemctl enable gemstone"

# Uninstall (requires root)
uninstall:
	@echo "Uninstalling gemstone..."
	@systemctl stop gemstone 2>/dev/null || true
	@systemctl disable gemstone 2>/dev/null || true
	@rm -f /etc/systemd/system/gemstone.service
	@systemctl daemon-reload
	@rm -f /usr/local/bin/gem
	@rm -f /usr/local/bin/gemstoned
	@echo "Gemstone uninstalled."
	@echo "Note: Configuration and data directories were not removed:"
	@echo "  $(CONFIG_DIR)"
	@echo "  $(DATA_DIR)"
	@echo "  $(LOG_DIR)"

# Clean build artifacts
clean:
	@rm -rf $(DIST_DIR)
	@go clean

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# Download dependencies
deps:
	go mod download
	go mod tidy

# Help
help:
	@echo "Gemstone Process Manager - Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build both CLI and daemon"
	@echo "  build-cli    Build only the CLI (gem)"
	@echo "  build-daemon Build only the daemon (gemstoned)"
	@echo "  dev          Development build (no optimizations)"
	@echo "  install      Install gemstone (requires root)"
	@echo "  uninstall    Uninstall gemstone (requires root)"
	@echo "  clean        Clean build artifacts"
	@echo "  test         Run tests"
	@echo "  fmt          Format code"
	@echo "  lint         Run linter"
	@echo "  deps         Download dependencies"
	@echo "  help         Show this help message"
