#!/bin/bash

set -e  # Exit on any error

echo "Setting up frontend dependencies..."

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check and install Deno if missing
if ! command_exists deno; then
    echo "Deno not found. Installing Deno..."
    if ! curl -fsSL https://deno.land/install.sh | sh; then
        echo "Error: Failed to install Deno" >&2
        exit 1
    fi
    echo "Deno installed successfully."
    # Add Deno to PATH for current session
    export PATH="$HOME/.deno/bin:$PATH"
else
    echo "Deno is already installed: $(deno --version)"
fi

# Install/cache dependencies
if command_exists deno; then
    echo "Caching frontend dependencies with Deno..."
    if ! (deno task --quiet || deno cache deno.json); then
        echo "Error: Failed to cache dependencies with Deno" >&2
        exit 1
    fi
    echo "Frontend dependencies cached successfully."
else
    echo "Error: Deno not available after installation attempt" >&2
    exit 1
fi

# Install Playwright browsers
if command_exists deno; then
    echo "Installing Playwright browsers with Deno..."
    if ! deno run npm:playwright install; then
        echo "Error: Failed to install Playwright browsers with Deno" >&2
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
    echo "Error: Neither Deno nor npm available for Playwright browser installation" >&2
    exit 1
fi

echo "Frontend setup complete."