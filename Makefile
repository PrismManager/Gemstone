.PHONY: build build-cli build-daemon dev stop-dev runcli test-dev install uninstall clean test fmt lint

# Variables
VERSION ?= 0.1.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || printf "unknown")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Directories
DIST_DIR := dist
CONFIG_DIR := /etc/gemstone
DATA_DIR := /var/lib/gemstone
LOG_DIR := /var/log/gemstone
RUN_DIR := /run/gemstone

# Colors
RESET := \033[0m
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
MAGENTA := \033[0;35m
CYAN := \033[0;36m
WHITE := \033[0;37m

# Build targets
all: build

build: build-cli build-daemon

build-cli:
	@printf "$(GREEN)Building gem CLI...$(RESET)\n"
	@mkdir -p $(DIST_DIR)
	go build $(LDFLAGS) -o $(DIST_DIR)/gem ./cmd/gem
	@printf "$(GREEN)Gem CLI built successfully.$(RESET)\n"

build-daemon:
	@printf "$(GREEN)Building gemstoned daemon...$(RESET)\n"
	@mkdir -p $(DIST_DIR)
	go build $(LDFLAGS) -o $(DIST_DIR)/gemstoned ./cmd/gemstoned
	@printf "$(GREEN)Gemstoned daemon built successfully.$(RESET)\n"

# Development builds
dev: build-cli build-daemon
	@printf "$(CYAN)Starting gemstone daemon in dev mode...$(RESET)\n"
	@mkdir -p development development/data development/logs development/run
	@GEMSTONE_CONFIG=./configs/config.yaml GEMSTONE_DATA=./development/data GEMSTONE_LOG=./development/logs GEMSTONE_SOCKET=./development/run/gemstone.sock ./dist/gemstoned &
	@printf "$(GREEN)Gemstone daemon started in dev mode.$(RESET)\n"

stop-dev:
	@printf "$(YELLOW)Stopping gemstone daemon...$(RESET)\n"
	@pkill -f "./dist/gemstoned" || printf "No daemon process found\n"

test-dev: dev
	@printf "$(CYAN)Starting test app with auto-restart...$(RESET)\n"
	@rm -rf ./development/
	@GEMSTONE_CONFIG=./configs/config.yaml GEMSTONE_DATA=./development/data GEMSTONE_LOG=./development/logs GEMSTONE_SOCKET=./development/run/gemstone.sock ./dist/gem start './test.py' --name testapp --auto-restart --max-restarts 5
	@printf "$(GREEN)Test app started. Check logs with: make runcli CMD=logs$(RESET)\n"
	@printf ""
	@printf "$(GREEN)Monitor with: make runcli CMD=status$(RESET)\n"
	@printf ""
	@printf "$(GREEN)Stop with: make runcli CMD='stop testapp'$(RESET)\n"

# Install (requires root)
install: build
	@printf "$(YELLOW)Installing gemstone...$(RESET)"
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
	@printf "$(GREEN)Gemstone installed successfully!$(RESET)"
	@printf ""
	@printf "$(CYAN)To start the daemon:$(RESET)"
	@printf "  sudo systemctl start gemstone"
	@printf ""
	@printf "$(CYAN)To enable auto-start on boot:$(RESET)"
	@printf "  sudo systemctl enable gemstone"

# Uninstall (requires root)
uninstall:
	@printf "$(YELLOW)Uninstalling gemstone...$(RESET)"
	@systemctl stop gemstone 2>/dev/null || true
	@systemctl disable gemstone 2>/dev/null || true
	@rm -f /etc/systemd/system/gemstone.service
	@systemctl daemon-reload
	@rm -f /usr/local/bin/gem
	@rm -f /usr/local/bin/gemstoned
	@printf "$(GREEN)Gemstone uninstalled.$(RESET)"
	@printf ""
	@printf "$(CYAN)Note: Configuration and data directories were not removed:$(RESET)"
	@printf "  $(CONFIG_DIR)"
	@printf "  $(DATA_DIR)"
	@printf "  $(LOG_DIR)"

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
		printf "golangci-lint not installed, skipping..."; \
	fi

# Download dependencies
deps:
	go mod download
	go mod tidy

# Help
help:
	@printf "$(BLUE)Gemstone Process Manager - Build System$(RESET)"
	@printf ""
	@printf "$(BLUE)Usage: make [target]$(RESET)"
	@printf ""
	@printf "$(BLUE)Targets:$(RESET)"
	@printf "  build        Build both CLI and daemon"
	@printf "  build-cli    Build only the CLI (gem)"
	@printf "  build-daemon Build only the daemon (gemstoned)"
	@printf "  dev          Development build (no optimizations)"
	@printf "  install      Install gemstone (requires root)"
	@printf "  uninstall    Uninstall gemstone (requires root)"
	@printf "  clean        Clean build artifacts"
	@printf "  test         Run tests"
	@printf "  fmt          Format code"
	@printf "  lint         Run linter"
	@printf "  deps         Download dependencies"
	@printf "  help         Show this help message"
