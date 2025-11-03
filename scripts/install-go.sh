#!/bin/bash

# Install Go for SecureChat development
# This script installs Go 1.21 or later

set -e

GO_VERSION="1.21.5"
GO_OS="linux"
GO_ARCH="amd64"

# Detect architecture
case $(uname -m) in
    x86_64) GO_ARCH="amd64" ;;
    aarch64|arm64) GO_ARCH="arm64" ;;
    armv7l) GO_ARCH="armv6l" ;;
    *) echo "Unsupported architecture: $(uname -m)"; exit 1 ;;
esac

# Detect OS
case $(uname -s) in
    Linux) GO_OS="linux" ;;
    Darwin) GO_OS="darwin" ;;
    *) echo "Unsupported OS: $(uname -s)"; exit 1 ;;
esac

GO_TARBALL="go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz"
GO_URL="https://golang.org/dl/${GO_TARBALL}"

echo "Installing Go ${GO_VERSION} for ${GO_OS}/${GO_ARCH}..."

# Check if Go is already installed
if command -v go >/dev/null 2>&1; then
    CURRENT_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo "Go ${CURRENT_VERSION} is already installed."
    
    # Check if version is sufficient
    if [ "$(printf '%s\n' "$GO_VERSION" "$CURRENT_VERSION" | sort -V | head -n1)" = "$GO_VERSION" ]; then
        echo "Current Go version is sufficient."
        exit 0
    else
        echo "Current Go version is too old. Installing newer version..."
    fi
fi

# Create temporary directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

# Download Go
echo "Downloading ${GO_URL}..."
curl -LO "$GO_URL"

# Extract Go
echo "Extracting Go..."
tar -xzf "$GO_TARBALL"

# Install Go
INSTALL_DIR="$HOME/.local"
mkdir -p "$INSTALL_DIR"

if [ -d "$INSTALL_DIR/go" ]; then
    echo "Removing existing Go installation..."
    rm -rf "$INSTALL_DIR/go"
fi

mv go "$INSTALL_DIR/"

# Update PATH
SHELL_RC=""
if [ -n "$BASH_VERSION" ]; then
    SHELL_RC="$HOME/.bashrc"
elif [ -n "$ZSH_VERSION" ]; then
    SHELL_RC="$HOME/.zshrc"
else
    SHELL_RC="$HOME/.profile"
fi

# Add Go to PATH if not already present
if ! grep -q "/.local/go/bin" "$SHELL_RC" 2>/dev/null; then
    echo "" >> "$SHELL_RC"
    echo "# Go installation" >> "$SHELL_RC"
    echo "export PATH=\$HOME/.local/go/bin:\$PATH" >> "$SHELL_RC"
    echo "export GOPATH=\$HOME/go" >> "$SHELL_RC"
    echo "export PATH=\$GOPATH/bin:\$PATH" >> "$SHELL_RC"
fi

# Clean up
cd /
rm -rf "$TEMP_DIR"

echo "Go ${GO_VERSION} installed successfully!"
echo "Please run 'source ${SHELL_RC}' or restart your shell to use Go."
echo ""
echo "Verify installation with: go version"
