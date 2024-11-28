#!/bin/bash

set -euo pipefail

__print_info() {
    printf "\033[0;34m[INFO]\033[0m %s\n" "$1"
}

__print_success() {
    printf "\033[0;32m[SUCCESS]\033[0m %s\n" "$1"
}

__print_error() {
    printf "\033[0;31m[ERROR]\033[0m %s\n" "$1" >&2
}

REPO_OWNER="AIO-Develope"
REPO_NAME="get"
CLI_NAME="get"

__check_dependencies() {
    for cmd in curl grep tr uname chmod; do
        if ! command -v "$cmd" &> /dev/null; then
            __print_error "Required dependency '$cmd' is not installed."
            exit 1
        fi
    done
}

__get_os() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        darwin|linux|windows)
            echo "$os"
            ;;
        *)
            __print_error "Unsupported operating system: $os"
            exit 1
            ;;
    esac
}

__get_arch() {
    local arch=$(uname -m)
    case "$arch" in
        x86_64)  echo "amd64" ;;
        aarch64) echo "arm64" ;;
        arm64)   echo "arm64" ;;
        armv7l)  echo "arm" ;;
        *)
            __print_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
}

__update_path() {
    local install_dir="$1"
    local shell_configs=(
        "$HOME/.bashrc"
        "$HOME/.bash_profile"
        "$HOME/.zshrc"
        "$HOME/.config/fish/config.fish"
    )

    mkdir -p "$install_dir"

    for config in "${shell_configs[@]}"; do
        if [ -f "$config" ]; then
            if ! grep -q "export PATH=.*$install_dir" "$config"; then
                case "$config" in
                    *fish*)
                        echo "fish_add_path $install_dir" >> "$config"
                        ;;
                    *)
                        echo "export PATH=\"$install_dir:\$PATH\"" >> "$config"
                        ;;
                esac
                __print_info "Updated $config with $install_dir"
            fi
        fi
    done

    export PATH="$install_dir:$PATH"
}

__get_latest_release() {
    local url="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
    curl -s -L "$url"
}

__find_download_url() {
    local os="$1"
    local arch="$2"
    local release="$3"

    echo "$release" | grep -i "browser_download_url" | grep -i "${os}" | grep -i "${arch}" | cut -d '"' -f 4 | head -n 1
}

install_cli() {
    __check_dependencies

    local os=$(__get_os)
    local arch=$(__get_arch)

    __print_info "Detected OS: $os, Architecture: $arch"

    local install_dir
    if [[ $EUID -eq 0 ]]; then
        install_dir="/usr/local/bin"
    else
        install_dir="$HOME/.local/bin"
    fi

    local release
    release=$(__get_latest_release)

    local download_url
    download_url=$(__find_download_url "$os" "$arch" "$release")

    if [ -z "$download_url" ]; then
        __print_error "No compatible binary found for ${os} ${arch}"
        exit 1
    fi

    local tmp_dir
    tmp_dir=$(mktemp -d)

    __print_info "Downloading ${CLI_NAME} for ${os} ${arch}..."
    if ! curl -L -o "$tmp_dir/$CLI_NAME" "$download_url"; then
        __print_error "Failed to download binary"
        exit 1
    fi

    chmod +x "$tmp_dir/$CLI_NAME"

    if ! mv "$tmp_dir/$CLI_NAME" "$install_dir/$CLI_NAME"; then
        __print_error "Failed to move binary to $install_dir"
        exit 1
    fi

    __update_path "$install_dir"

    if command -v "$CLI_NAME" &> /dev/null; then
        __print_success "${CLI_NAME} installed successfully!"
        __print_info "Installed at: $(command -v "$CLI_NAME")"
        "$CLI_NAME" --version
    else
        __print_error "Installation failed. ${CLI_NAME} not found in PATH."
        exit 1
    fi

    rm -rf "$tmp_dir"
}

install_cli
