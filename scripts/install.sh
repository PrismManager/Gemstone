#!/bin/bash
# Gemstone Quick Install Script
# Usage: curl -fsSL https://raw.githubusercontent.com/PrismManager/gemstone/main/scripts/install.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Gemstone Process Manager - Installer${NC}"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root (sudo)${NC}"
    exit 1
fi

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    echo -e "${RED}Cannot detect OS${NC}"
    exit 1
fi

echo "Detected OS: $OS"

# Check for Go
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}Go is not installed. Installing...${NC}"
    
    case $OS in
        ubuntu|debian)
            apt-get update
            apt-get install -y golang-go
            ;;
        centos|rhel|fedora)
            dnf install -y golang
            ;;
        arch)
            pacman -S --noconfirm go
            ;;
        *)
            echo -e "${RED}Please install Go manually and re-run this script${NC}"
            exit 1
            ;;
    esac
fi

# Create temp directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

echo "Downloading Gemstone..."
git clone https://github.com/PrismManager/gemstone.git
cd gemstone

echo "Building..."
make build

echo "Installing..."
make install

# Cleanup
cd /
rm -rf "$TEMP_DIR"

echo ""
echo -e "${GREEN}Gemstone installed successfully!${NC}"
echo ""
echo "Quick Start:"
echo "  1. Start the daemon:  sudo systemctl start gemstone"
echo "  2. Enable on boot:    sudo systemctl enable gemstone"
echo "  3. Start a process:   gem start 'node app.js' --name myapp"
echo "  4. List processes:    gem list"
echo "  5. View logs:         gem logs myapp"
echo ""
echo "For more information, visit: https://github.com/PrismManager/gemstone"
