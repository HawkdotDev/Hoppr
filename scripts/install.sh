#!/usr/bin/env bash
# Hoppr — Professional macOS / Linux Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/HawkdotDev/Hoppr/main/scripts/install.sh | bash
#
# Security: Downloads checksums.txt from the release and verifies the SHA256
#           hash of the archive before extraction. Aborts on mismatch.

set -euo pipefail

# ── Constants ──────────────────────────────────────────────────────────────────
REPO="HawkdotDev/Hoppr"
BIN_NAME="hop"

# ── Colors ─────────────────────────────────────────────────────────────────────
if [ -t 1 ] && command -v tput >/dev/null 2>&1; then
    BOLD=$(tput bold)
    CYAN=$(tput setaf 6)
    GREEN=$(tput setaf 2)
    YELLOW=$(tput setaf 3)
    RED=$(tput setaf 1)
    MAGENTA=$(tput setaf 5)
    GRAY=$(tput setaf 8)
    RESET=$(tput sgr0)
else
    BOLD="" CYAN="" GREEN="" YELLOW="" RED="" MAGENTA="" GRAY="" RESET=""
fi

step()  { echo "${CYAN}  ● ${1}${RESET}"; }
ok()    { echo "${GREEN}  ✔ ${1}${RESET}"; }
warn()  { echo "${YELLOW}  ⚠ ${1}${RESET}"; }
fail()  { echo "${RED}  ✖ ${1}${RESET}"; }
abort() { fail "$1"; cleanup; exit 1; }

# ── Cleanup Trap ───────────────────────────────────────────────────────────────
TMP_DIR=""
cleanup() {
    if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
}
trap cleanup EXIT INT TERM

# ── Banner ─────────────────────────────────────────────────────────────────────
echo ""
echo "${MAGENTA}  ╔═══════════════════════════════════════╗${RESET}"
echo "${MAGENTA}  ║     Hoppr Installer for Unix/macOS     ║${RESET}"
echo "${MAGENTA}  ║   The fast lane to your projects  🚀   ║${RESET}"
echo "${MAGENTA}  ╚═══════════════════════════════════════╝${RESET}"
echo ""

# ── 1. Detect OS & Architecture ───────────────────────────────────────────────
step "Detecting system architecture..."

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
RAW_ARCH="$(uname -m)"

case "$RAW_ARCH" in
    x86_64|amd64)   ARCH="amd64" ;;
    aarch64|arm64)   ARCH="arm64" ;;
    *)               abort "Unsupported architecture: $RAW_ARCH" ;;
esac

case "$OS" in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    *)      abort "Unsupported OS: $OS" ;;
esac

ok "Architecture: ${OS}/${ARCH}"

# ── 2. Resolve Latest Release ─────────────────────────────────────────────────
step "Fetching latest release from GitHub..."

TAG=""
if command -v curl >/dev/null 2>&1; then
    TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" \
        -H "User-Agent: Hoppr-Installer/1.0" \
        | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)
elif command -v wget >/dev/null 2>&1; then
    TAG=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" \
        --header="User-Agent: Hoppr-Installer/1.0" \
        | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)
fi

if [ -z "$TAG" ]; then
    warn "GitHub API unreachable. Falling back to v1.0.0."
    TAG="v1.0.0"
fi
ok "Release: ${TAG}"

# ── 3. Prepare Temp Directory ─────────────────────────────────────────────────
TMP_DIR="$(mktemp -d)"

ASSET="hoppr-${TAG}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${TAG}/checksums.txt"
ARCHIVE_PATH="${TMP_DIR}/${ASSET}"

# ── 4. Download Archive ───────────────────────────────────────────────────────
step "Downloading ${ASSET}..."

download() {
    local url="$1" dest="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$dest"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$dest"
    else
        abort "Neither curl nor wget found. Please install one and try again."
    fi
}

download "$DOWNLOAD_URL" "$ARCHIVE_PATH" || abort "Failed to download release archive. Visit https://github.com/${REPO}/releases"

FILESIZE=$(du -h "$ARCHIVE_PATH" | cut -f1)
ok "Downloaded ${FILESIZE}"

# ── 5. Verify SHA256 Checksum ─────────────────────────────────────────────────
step "Verifying SHA256 checksum..."

CHECKSUM_FILE="${TMP_DIR}/checksums.txt"
VERIFIED=0

if download "$CHECKSUM_URL" "$CHECKSUM_FILE" 2>/dev/null; then
    EXPECTED_LINE=$(grep "$ASSET" "$CHECKSUM_FILE" | head -n 1 || true)
    if [ -n "$EXPECTED_LINE" ]; then
        EXPECTED_HASH=$(echo "$EXPECTED_LINE" | awk '{print $1}' | tr '[:upper:]' '[:lower:]')

        # Use sha256sum (Linux) or shasum (macOS)
        if command -v sha256sum >/dev/null 2>&1; then
            ACTUAL_HASH=$(sha256sum "$ARCHIVE_PATH" | awk '{print $1}' | tr '[:upper:]' '[:lower:]')
        elif command -v shasum >/dev/null 2>&1; then
            ACTUAL_HASH=$(shasum -a 256 "$ARCHIVE_PATH" | awk '{print $1}' | tr '[:upper:]' '[:lower:]')
        else
            warn "No sha256sum or shasum found — skipping verification."
            EXPECTED_HASH=""
        fi

        if [ -n "$EXPECTED_HASH" ] && [ -n "$ACTUAL_HASH" ]; then
            if [ "$EXPECTED_HASH" != "$ACTUAL_HASH" ]; then
                abort "SHA256 MISMATCH! Expected: ${EXPECTED_HASH}  Got: ${ACTUAL_HASH} — The download may be corrupted or tampered with."
            fi
            ok "SHA256 verified: ${ACTUAL_HASH:0:16}..."
            VERIFIED=1
        fi
    else
        warn "No checksum entry found for ${ASSET} — skipping verification."
    fi
else
    warn "Could not download checksums.txt — skipping verification."
fi

if [ "$VERIFIED" -eq 0 ]; then
    warn "Checksum verification was skipped."
fi

# ── 6. Extract and Install ────────────────────────────────────────────────────
step "Installing Hoppr..."

# Determine install paths
if [ "$(id -u)" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
    DATA_DIR="/usr/local/share/hoppr"
else
    INSTALL_DIR="${HOME}/.local/bin"
    DATA_DIR="${XDG_DATA_HOME:-${HOME}/.local/share}/hoppr"
fi

SHELL_DIR="${DATA_DIR}/shell"
mkdir -p "$INSTALL_DIR" "$SHELL_DIR"

# Extract
EXTRACT_DIR="${TMP_DIR}/extracted"
mkdir -p "$EXTRACT_DIR"
tar -xzf "$ARCHIVE_PATH" -C "$EXTRACT_DIR"

# Find and install binary
EXTRACTED_BIN=$(find "$EXTRACT_DIR" -name "$BIN_NAME" -type f | head -n 1)
if [ -z "$EXTRACTED_BIN" ]; then
    abort "Could not find ${BIN_NAME} in the release archive. The release may be corrupt."
fi

cp "$EXTRACTED_BIN" "${INSTALL_DIR}/${BIN_NAME}"
chmod +x "${INSTALL_DIR}/${BIN_NAME}"

# Copy shell integration scripts from archive
EXTRACTED_SHELL=$(find "$EXTRACT_DIR" -type d -name "shell" | head -n 1)
if [ -n "$EXTRACTED_SHELL" ]; then
    cp "$EXTRACTED_SHELL"/* "$SHELL_DIR/" 2>/dev/null || true
else
    # Fallback: download shell wrappers directly
    warn "Shell scripts not found in archive. Downloading directly..."
    for WRAPPER in hop.bash hop.zsh hop.fish; do
        download "https://raw.githubusercontent.com/${REPO}/main/shell/${WRAPPER}" "${SHELL_DIR}/${WRAPPER}" 2>/dev/null || true
    done
fi

ok "Binary installed to ${INSTALL_DIR}/${BIN_NAME}"

# ── 7. Configure PATH ─────────────────────────────────────────────────────────
step "Checking PATH..."

if echo "$PATH" | grep -q "$INSTALL_DIR"; then
    ok "PATH already includes ${INSTALL_DIR}"
else
    warn "${INSTALL_DIR} is not in your PATH."
    echo "${GRAY}    Add this to your shell profile:${RESET}"
    echo "${GRAY}    export PATH=\"\$PATH:${INSTALL_DIR}\"${RESET}"
fi

# ── 8. Configure Shell Integration ────────────────────────────────────────────
step "Setting up shell integration..."

CONFIGURED=0

# Bash
BASHRC="${HOME}/.bashrc"
if [ -f "$BASHRC" ]; then
    if ! grep -q "hop.bash" "$BASHRC" 2>/dev/null; then
        printf '\n# Hoppr Shell Integration (added by installer)\nsource "%s/hop.bash"\n' "$SHELL_DIR" >> "$BASHRC"
        ok "Added to ~/.bashrc"
        CONFIGURED=1
    else
        ok "~/.bashrc already configured"
        CONFIGURED=1
    fi
fi

# Zsh
ZSHRC="${HOME}/.zshrc"
if [ -f "$ZSHRC" ] || command -v zsh >/dev/null 2>&1; then
    if ! grep -q "hop.zsh" "$ZSHRC" 2>/dev/null; then
        [ ! -f "$ZSHRC" ] && touch "$ZSHRC"
        printf '\n# Hoppr Shell Integration (added by installer)\nsource "%s/hop.zsh"\n' "$SHELL_DIR" >> "$ZSHRC"
        ok "Added to ~/.zshrc"
        CONFIGURED=1
    else
        ok "~/.zshrc already configured"
        CONFIGURED=1
    fi
fi

# Fish
FISH_CONFIG="${HOME}/.config/fish/config.fish"
if command -v fish >/dev/null 2>&1; then
    if [ -f "$FISH_CONFIG" ]; then
        if ! grep -q "hop.fish" "$FISH_CONFIG" 2>/dev/null; then
            printf '\n# Hoppr Shell Integration (added by installer)\nsource "%s/hop.fish"\n' "$SHELL_DIR" >> "$FISH_CONFIG"
            ok "Added to ~/.config/fish/config.fish"
            CONFIGURED=1
        else
            ok "Fish config already configured"
            CONFIGURED=1
        fi
    fi
fi

if [ "$CONFIGURED" -eq 0 ]; then
    warn "Could not detect your shell configuration file."
    echo "${GRAY}    Manually add: source \"${SHELL_DIR}/hop.bash\" to your shell profile.${RESET}"
fi

# ── Done ───────────────────────────────────────────────────────────────────────
echo ""
echo "${GREEN}  ╔═══════════════════════════════════════╗${RESET}"
echo "${GREEN}  ║     Hoppr installed successfully! 🎉   ║${RESET}"
echo "${GREEN}  ╚═══════════════════════════════════════╝${RESET}"
echo ""
echo "  Get started:"
echo "${GRAY}    hop doctor            — verify installation health${RESET}"
echo "${GRAY}    hop add .            — save current folder as a project${RESET}"
echo "${GRAY}    hop list             — view all saved projects${RESET}"
echo "${GRAY}    hop <project>        — jump to any saved project${RESET}"
echo ""
echo "${GRAY}  Restart your terminal or run 'source ~/.bashrc' to activate 'hop'.${RESET}"
echo "${GRAY}  Docs: https://github.com/${REPO}${RESET}"
echo ""
