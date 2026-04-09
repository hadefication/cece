#!/usr/bin/env bash
#
# Install cece (cc) — Claude Code session manager
# Usage: curl -sSL https://raw.githubusercontent.com/inggo/cece/main/install.sh | bash
#

set -euo pipefail

REPO="inggo/cece"
BINARY="cc"
INSTALL_DIR="${HOME}/.local/bin"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

echo "Fetching latest release..."
LATEST=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Could not determine latest version."
  exit 1
fi

echo "Installing cece ${LATEST} (${OS}/${ARCH})..."

URL="https://github.com/${REPO}/releases/download/${LATEST}/cece_${LATEST#v}_${OS}_${ARCH}.tar.gz"
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

curl -sSL "$URL" -o "${TMPDIR}/cece.tar.gz"
tar -xzf "${TMPDIR}/cece.tar.gz" -C "$TMPDIR"

mkdir -p "$INSTALL_DIR"
cp "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"

echo "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"

if ! echo "$PATH" | tr ':' '\n' | grep -q "^${INSTALL_DIR}$"; then
  echo ""
  echo "Add to your PATH:"
  echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi

echo ""
echo "Run 'cc init' to get started."
