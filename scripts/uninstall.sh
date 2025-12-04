#!/bin/bash
# Gemstone Uninstall Script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Gemstone Process Manager - Uninstaller${NC}"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root (sudo)${NC}"
    exit 1
fi

# Stop and disable service
echo "Stopping gemstone service..."
systemctl stop gemstone 2>/dev/null || true
systemctl disable gemstone 2>/dev/null || true

# Remove files
echo "Removing binaries..."
rm -f /usr/local/bin/gem
rm -f /usr/local/bin/gemstoned

echo "Removing systemd service..."
rm -f /etc/systemd/system/gemstone.service
systemctl daemon-reload

echo ""
echo -e "${GREEN}Gemstone has been uninstalled.${NC}"
echo ""
echo "The following directories were NOT removed (they may contain your data):"
echo "  - /etc/gemstone (configuration)"
echo "  - /var/lib/gemstone (process data)"
echo "  - /var/log/gemstone (logs)"
echo ""
echo "To remove all data, run:"
echo "  sudo rm -rf /etc/gemstone /var/lib/gemstone /var/log/gemstone"
