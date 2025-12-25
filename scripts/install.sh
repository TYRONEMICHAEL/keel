#!/bin/bash
set -e

# Keel installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/TYRONEMICHAEL/keel/main/scripts/install.sh | bash

REPO="TYRONEMICHAEL/keel"
BINARY="keel"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    darwin|linux) ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get latest release
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    echo "Could not determine latest version. Falling back to building from source..."
    if command -v go &> /dev/null; then
        go install "github.com/$REPO/cmd/keel@latest"
        echo "Installed keel via 'go install'"
        exit 0
    else
        echo "Go is not installed. Please install Go or download a binary from:"
        echo "https://github.com/$REPO/releases"
        exit 1
    fi
fi

# Download binary
URL="https://github.com/$REPO/releases/download/$LATEST/${BINARY}-${OS}-${ARCH}"
echo "Downloading keel $LATEST for $OS/$ARCH..."

mkdir -p "$INSTALL_DIR"
curl -fsSL "$URL" -o "$INSTALL_DIR/$BINARY"
chmod +x "$INSTALL_DIR/$BINARY"

# Create 'ke' alias
ln -sf "$INSTALL_DIR/$BINARY" "$INSTALL_DIR/ke"

echo ""
echo "Installed keel to $INSTALL_DIR/$BINARY"
echo "Created alias: ke -> keel"
echo ""

# Check if in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "Add to your PATH:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    echo ""
fi

echo "Run 'keel --version' to verify installation."
