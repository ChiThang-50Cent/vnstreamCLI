#!/usr/bin/env bash

set -euo pipefail

REPO="ChiThang-50Cent/vnstreamCLI"
APP_NAME="vnstream"
DEST_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
DEST_BIN="${DEST_DIR}/${APP_NAME}"
VERSION="${VERSION:-latest}"

TMP_DIR=""

info() {
    printf '[INFO] %s\n' "$1"
}

warn() {
    printf '[WARN] %s\n' "$1"
}

error() {
    printf '[ERROR] %s\n' "$1" >&2
    exit 1
}

cleanup() {
    if [[ -n "$TMP_DIR" && -d "$TMP_DIR" ]]; then
        rm -rf "$TMP_DIR"
    fi
}

require_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        error "Missing required command: $1"
    fi
}

detect_platform() {
    local os_raw arch_raw

    os_raw="$(uname -s)"
    arch_raw="$(uname -m)"

    case "$os_raw" in
        Linux) PLATFORM_OS="linux" ;;
        Darwin) PLATFORM_OS="darwin" ;;
        *) PLATFORM_OS="unsupported" ;;
    esac

    case "$arch_raw" in
        x86_64|amd64) PLATFORM_ARCH="amd64" ;;
        aarch64|arm64) PLATFORM_ARCH="arm64" ;;
        *) PLATFORM_ARCH="unsupported" ;;
    esac
}

assert_supported_release() {
    if [[ "$PLATFORM_OS" == "linux" && "$PLATFORM_ARCH" == "amd64" ]]; then
        return 0
    fi

    error "No prebuilt release for this platform (${PLATFORM_OS}/${PLATFORM_ARCH}). Please build from source on your machine: https://github.com/${REPO}"
}

build_asset_url() {
    local asset_name

    asset_name="${APP_NAME}_${PLATFORM_OS}_${PLATFORM_ARCH}.tar.gz"
    if [[ "$VERSION" == "latest" ]]; then
        ASSET_URL="https://github.com/${REPO}/releases/latest/download/${asset_name}"
    else
        ASSET_URL="https://github.com/${REPO}/releases/download/${VERSION}/${asset_name}"
    fi
}

check_runtime_dependencies() {
    if ! command -v vlc >/dev/null 2>&1 && ! command -v qvlc >/dev/null 2>&1; then
        warn "VLC not found (vlc/qvlc)."
        warn "You can still install VNStream, but playback will not work until VLC is installed."
        warn "Ubuntu/Debian: sudo apt install vlc"
        warn "Fedora: sudo dnf install vlc"
        warn "Arch: sudo pacman -S vlc"
        warn "macOS (Homebrew): brew install vlc"
    fi
}

download_and_extract_binary() {
    local archive_path extracted_path

    TMP_DIR="$(mktemp -d)"
    archive_path="${TMP_DIR}/${APP_NAME}.tar.gz"

    info "Downloading release asset: ${ASSET_URL}"
    if ! curl -fsSL "$ASSET_URL" -o "$archive_path"; then
        error "Failed to download release asset. Check VERSION=${VERSION} and release assets."
    fi

    tar -xzf "$archive_path" -C "$TMP_DIR"
    extracted_path="${TMP_DIR}/${APP_NAME}"
    if [[ ! -f "$extracted_path" ]]; then
        error "Downloaded archive does not contain '${APP_NAME}' binary."
    fi
    chmod +x "$extracted_path"
    SRC_BIN="$extracted_path"
}

install_binary() {
    mkdir -p "$DEST_DIR"
    install -m 0755 "$SRC_BIN" "$DEST_BIN"
}

check_path_hint() {
    case ":$PATH:" in
        *":$DEST_DIR:"*) return 0 ;;
    esac

    warn "$DEST_DIR is not in PATH."
    warn "Add this to your shell rc file (e.g. ~/.bashrc, ~/.zshrc):"
    printf '  export PATH="%s:$PATH"\n' "$DEST_DIR"
}

main() {
    trap cleanup EXIT

    require_cmd curl
    require_cmd tar
    require_cmd install

    detect_platform
    assert_supported_release
    build_asset_url
    check_runtime_dependencies
    download_and_extract_binary

    info "Installing ${APP_NAME} to: $DEST_BIN"
    install_binary

    info "Installation completed successfully."
    info "Run the app with: ${APP_NAME}"
    info "Or quick search: ${APP_NAME} \"movie name\""

    check_path_hint
}

main "$@"
