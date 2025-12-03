#!/bin/bash

set -e  # Exit on any error

echo "Setting up frontend dependencies..."

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check and install Deno if missing
if ! command_exists deno; then
    # Check if running in CI environment
    if [ -n "${CI:-}" ] || [ -n "${GITHUB_ACTIONS:-}" ]; then
        echo "Warning: Running in CI environment. For GitHub Actions, use the official 'denoland/setup-deno' action:" >&2
        echo "  - name: Set up Deno" >&2
        echo "    uses: denoland/setup-deno@v2" >&2
        echo "    with:" >&2
        echo "      deno-version: v1.x" >&2
        echo "" >&2
        echo "Proceeding with manual installation for this session..." >&2
    fi
    
    echo "Deno not found. Installing Deno securely..."
    
    # Create temporary directory for installer
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    INSTALLER_SCRIPT="$TEMP_DIR/install.sh"
    CHECKSUM_FILE="$TEMP_DIR/SHA256SUM"
    
    # Download installer script
    echo "Downloading Deno installer..."
    if ! curl -fsSL -o "$INSTALLER_SCRIPT" https://deno.land/install.sh; then
        echo "Error: Failed to download Deno installer" >&2
        exit 1
    fi
    
    # Optional: Add checksum verification here if needed
    
    # Execute verified installer
    echo "Installing Deno..."
    if ! sh "$INSTALLER_SCRIPT"; then
        echo "Error: Failed to install Deno" >&2
        exit 1
    fi
    
    echo "Deno installed successfully."
    
    # Export DENO_INSTALL and add to PATH for current session
    export DENO_INSTALL="$HOME/.deno"
    export PATH="$DENO_INSTALL/bin:$PATH"
    
    # Detect shell and append to appropriate profile file for persistence
    SHELL_PROFILE=""
    if [ -n "${ZSH_VERSION:-}" ]; then
        SHELL_PROFILE="$HOME/.zshrc"
    elif [ -n "${BASH_VERSION:-}" ]; then
        SHELL_PROFILE="$HOME/.bashrc"
        # Fallback to .profile if .bashrc doesn't exist
        if [ ! -f "$SHELL_PROFILE" ]; then
            SHELL_PROFILE="$HOME/.profile"
        fi
    else
    SHELL_PROFILE=""
    if [ "$SHELL" = "/bin/zsh" ] || [ "$SHELL" = "/usr/bin/zsh" ]; then
        SHELL_PROFILE="$HOME/.zshrc"
    elif [ "$SHELL" = "/bin/bash" ] || [ "$SHELL" = "/usr/bin/bash" ]; then
        SHELL_PROFILE="$HOME/.bashrc"
        # Fallback to .profile if .bashrc doesn't exist
        if [ ! -f "$SHELL_PROFILE" ]; then
            SHELL_PROFILE="$HOME/.profile"
        fi
    else
        # Default to .profile for other shells
        SHELL_PROFILE="$HOME/.profile"
    fi
                echo ""
                echo "# Deno"
                echo "$PROFILE_LINE"
                echo "$PATH_LINE"
            } >> "$SHELL_PROFILE" 2>/dev/null; then
                echo "Added Deno to PATH in $SHELL_PROFILE"
                echo "Note: Please restart your shell or run 'source $SHELL_PROFILE' for the changes to take effect."
            else
                echo "Warning: Could not write to $SHELL_PROFILE. Please manually add:" >&2
                echo "  export DENO_INSTALL=\"\$HOME/.deno\"" >&2
                echo "  export PATH=\"\$DENO_INSTALL/bin:\$PATH\"" >&2
            fi
        else
            echo "Deno PATH already configured in $SHELL_PROFILE"
        fi
    fi
else
    echo "Deno is already installed: $(deno --version)"
fi

# Install/cache dependencies
if command_exists deno; then
    echo "Caching frontend dependencies with Deno..."
    # Try running cache task if defined, otherwise skip
    if ! deno task cache 2>/dev/null; then
        echo "Warning: No cache task defined, skipping dependency caching" >&2
    fi
fi

# Install Playwright browsers
if command_exists deno; then
    echo "Installing Playwright browsers with Deno..."
    if ! deno run npm:playwright install; then
        echo "Error: Failed to install Playwright browsers with Deno" >&2
        exit 1
    fi
    echo "Playwright browsers installed successfully."
elif command_exists bun; then
    echo "Installing Playwright browsers with bun..."
    if ! bunx playwright install; then
        echo "Error: Failed to install Playwright browsers with bun" >&2
        exit 1
    fi
    echo "Playwright browsers installed successfully."
elif command_exists npm; then
    echo "Installing Playwright browsers with npm..."
    if ! npx playwright install; then
        echo "Error: Failed to install Playwright browsers with npm" >&2
        exit 1
    fi
    echo "Playwright browsers installed successfully."
else
    echo "Error: Neither Deno, bun, nor npm available for Playwright browser installation" >&2
    exit 1
fi

echo "Frontend setup complete."