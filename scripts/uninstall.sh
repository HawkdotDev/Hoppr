#!/usr/bin/env bash
# Hoppr Uninstaller for macOS / Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/HawkdotDev/Hoppr/main/scripts/uninstall.sh | bash

set -u

echo ""
echo -e "\033[33m=========================================\033[0m"
echo -e "\033[33m     Hoppr Uninstaller for Unix/macOS    \033[0m"
echo -e "\033[33m=========================================\033[0m"
echo ""

# 1. Determine Binary and Share Locations
if [ "$(id -u)" -eq 0 ]; then
    BIN_PATH="/usr/local/bin/hop"
    DATA_DIR="/usr/local/share/hoppr"
else
    BIN_PATH="${HOME}/.local/bin/hop"
    DATA_DIR="${XDG_DATA_HOME:-${HOME}/.local/share}/hoppr"
fi
CONFIG_DIR="${HOME}/.hoppr"

# 2. Remove Binary
if [ -f "$BIN_PATH" ]; then
    echo -e "\033[36m[*] Removing binary $BIN_PATH...\033[0m"
    rm -f "$BIN_PATH"
    echo -e "\033[32m[+] Binary removed.\033[0m"
fi

# 3. Remove Data / Shell directory
if [ -d "$DATA_DIR" ]; then
    echo -e "\033[36m[*] Removing application files from $DATA_DIR...\033[0m"
    rm -rf "$DATA_DIR"
    echo -e "\033[32m[+] Application files removed.\033[0m"
fi

# 4. Clean Shell Profiles
clean_file() {
    local file="$1"
    local pattern="$2"
    if [ -f "$file" ] && grep -q "$pattern" "$file" 2>/dev/null; then
        echo -e "\033[36m[*] Removing shell integration from $file...\033[0m"
        sed -i.hoppr-bak "/$pattern/d" "$file" 2>/dev/null || sed -i "" "/$pattern/d" "$file" 2>/dev/null
        # Remove empty comment lines
        sed -i.hoppr-bak '/# Hoppr Shell Integration/d' "$file" 2>/dev/null || sed -i "" '/# Hoppr Shell Integration/d' "$file" 2>/dev/null
        rm -f "${file}.hoppr-bak" 2>/dev/null || true
        echo -e "\033[32m[+] Cleaned $file.\033[0m"
    fi
}

clean_file "$HOME/.bashrc" "hop.bash"
clean_file "$HOME/.zshrc" "hop.zsh"
clean_file "$HOME/.config/fish/config.fish" "hop.fish"

# 5. Prompt to Remove Config Directory
echo ""
read -p "Do you also want to delete your saved projects and config (~/.hoppr)? (y/N): " choice
case "$choice" in
    y|Y)
        if [ -d "$CONFIG_DIR" ]; then
            rm -rf "$CONFIG_DIR"
            echo -e "\033[32m[+] Config and saved projects removed.\033[0m"
        fi
        ;;
    *)
        echo -e "\033[90m[*] Preserved ~/.hoppr configuration.\033[0m"
        ;;
esac

echo ""
echo -e "\033[32m=========================================\033[0m"
echo -e "\033[32m     Hoppr uninstalled successfully!     \033[0m"
echo -e "\033[32m=========================================\033[0m"
echo ""
