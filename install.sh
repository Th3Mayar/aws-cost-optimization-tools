#!/bin/bash
# Installation script for coaws
# Supports Linux and macOS

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Repository information
REPO="Th3Mayar/aws-cost-optimization-tools"
BINARY_NAME="coaws"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

echo -e "${CYAN}"
echo "    ██████╗ ██████╗  █████╗ ██╗    ██╗███████╗"
echo "   ██╔════╝██╔═══██╗██╔══██╗██║    ██║██╔════╝"
echo "   ██║     ██║   ██║███████║██║ █╗ ██║███████╗"
echo "   ██║     ██║   ██║██╔══██║██║███╗██║╚════██║"
echo "   ╚██████╗╚██████╔╝██║  ██║╚███╔███╔╝███████║"
echo "    ╚═════╝ ╚═════╝ ╚═╝  ╚═╝ ╚══╝╚══╝ ╚══════╝"
echo -e "${NC}"
echo -e "${BLUE}AWS Cost Optimization & Savings Tool${NC}"
echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    i386|i686)
        ARCH="386"
        ;;
    armv7l)
        ARCH="arm"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "${GREEN}Detected OS: ${NC}$OS"
echo -e "${GREEN}Detected Architecture: ${NC}$ARCH"
echo ""

# Get latest release version
echo -e "${BLUE}Fetching latest release...${NC}"
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo -e "${RED}Failed to fetch latest version. Please check your internet connection.${NC}"
    exit 1
fi

echo -e "${GREEN}Latest version: ${NC}$LATEST_VERSION"

# Construct download URL
ARCHIVE_NAME="coaws_${LATEST_VERSION#v}_${OS}_${ARCH}.tar.gz"
if [ "$OS" = "windows" ]; then
    ARCHIVE_NAME="coaws_${LATEST_VERSION#v}_${OS}_${ARCH}.zip"
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/$ARCHIVE_NAME"

echo -e "${BLUE}Downloading $ARCHIVE_NAME...${NC}"

# Create temporary directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Download archive
if ! curl -sL "$DOWNLOAD_URL" -o "$ARCHIVE_NAME"; then
    echo -e "${RED}Failed to download $ARCHIVE_NAME${NC}"
    echo -e "${RED}URL: $DOWNLOAD_URL${NC}"
    rm -rf "$TMP_DIR"
    exit 1
fi

echo -e "${GREEN}Download complete!${NC}"

# Extract archive
echo -e "${BLUE}Extracting...${NC}"
if [ "$OS" = "windows" ]; then
    unzip -q "$ARCHIVE_NAME"
else
    tar -xzf "$ARCHIVE_NAME"
fi

# Install binary
echo -e "${BLUE}Installing to $INSTALL_DIR...${NC}"

if [ ! -d "$INSTALL_DIR" ]; then
    echo -e "${BLUE}Creating $INSTALL_DIR...${NC}"
    sudo mkdir -p "$INSTALL_DIR"
fi

if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Cleanup
cd - > /dev/null
rm -rf "$TMP_DIR"

echo ""
echo -e "${GREEN}✓ Installation complete!${NC}"
echo ""
echo -e "${CYAN}To get started, run:${NC}"
echo -e "  ${BINARY_NAME} --help"
echo -e "  ${BINARY_NAME} start"
echo ""
echo -e "${CYAN}For more information, visit:${NC}"
echo -e "  https://github.com/$REPO"
echo ""
