#!/usr/bin/env bash
# Hoppr macOS / Linux 1-Line Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/HawkdotDev/Hoppr/main/scripts/install.sh | bash

set -e

REPO="HawkdotDev/Hoppr"
echo -e "\033[36m>>> Installing Hoppr...\033[0m"

# 1. Detect OS & Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  linux) OS="linux" ;;
  darwin) OS="darwin" ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# 2. Determine Latest Release
echo -e "\033[33m>>> Fetching latest release from GitHub...\033[0m"
TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "v1.0.0")
[ -z "$TAG" ] && TAG="v1.0.0"

ASSET="hoppr-${TAG}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

# 3. Setup Install Target
INSTALL_DIR="${HOME}/.local/bin"
SHELL_DIR="${HOME}/.local/share/hoppr/shell"
mkdir -p "$INSTALL_DIR"
mkdir -p "$SHELL_DIR"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

# 4. Download and Extract
echo -e "\033[33m>>> Downloading ${ASSET}...\033[0m"
curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/hoppr.tar.gz"
tar -xzf "$TMP_DIR/hoppr.tar.gz" -C "$TMP_DIR"

# Move binary
EXTRACTED_BIN=$(find "$TMP_DIR" -name "hoppr" -type f | head -n 1)
cp "$EXTRACTED_BIN" "$INSTALL_DIR/hoppr"
chmod +x "$INSTALL_DIR/hoppr"

# Download shell wrapper
curl -fsSL "https://raw.githubusercontent.com/${REPO}/main/shell/hop.bash" -o "$SHELL_DIR/hop.bash"
curl -fsSL "https://raw.githubusercontent.com/${REPO}/main/shell/hop.zsh" -o "$SHELL_DIR/hop.zsh"
curl -fsSL "https://raw.githubusercontent.com/${REPO}/main/shell/hop.fish" -o "$SHELL_DIR/hop.fish"

# 5. Configure Shell Hook
CONFIGURED=0
if [ -n "$BASH_VERSION" ] && [ -f "$HOME/.bashrc" ]; then
    if ! grep -q "hop.bash" "$HOME/.bashrc"; then
        echo -e "\n# Hoppr Shell Integration\nsource \"$SHELL_DIR/hop.bash\"" >> "$HOME/.bashrc"
        CONFIGURED=1
    fi
fi

if [ -f "$HOME/.zshrc" ]; then
    if ! grep -q "hop.zsh" "$HOME/.zshrc"; then
        echo -e "\n# Hoppr Shell Integration\nsource \"$SHELL_DIR/hop.zsh\"" >> "$HOME/.zshrc"
        CONFIGURED=1
    fi
fi

echo ""
echo -e "\033[32m>>> Hoppr installed successfully to ${INSTALL_DIR}/hoppr! 🚀\033[0m"
if [ "$CONFIGURED" -eq 1 ]; then
    echo -e "\033[36m>>> Shell integration added. Run 'source ~/.bashrc' (or ~/.zshrc) to activate 'hop'.\033[0m"
fi
echo -e "\033[32m>>> Run 'hoppr doctor' to verify your installation.\033[0m"
