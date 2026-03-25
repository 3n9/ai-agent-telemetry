#!/usr/bin/env sh
# Install ai-log and ai-log-report from GitHub Releases.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/3n9/ai-agent-telemetry/main/scripts/install.sh | sh
#
# Environment variables:
#   VERSION     — specific release tag to install (e.g. VERSION=v0.2.0)
#   INSTALL_DIR — destination directory (default: ~/.local/bin)

set -e

REPO="3n9/ai-agent-telemetry"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-}"

# ── Detect OS and architecture ──────────────────────────────────────────────

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  darwin) PLAT="darwin" ;;
  linux)  PLAT="linux"  ;;
  *)
    echo "❌ Unsupported OS: $OS"
    echo "   Download manually: https://github.com/$REPO/releases"
    exit 1
    ;;
esac

case "$ARCH" in
  arm64|aarch64) PLAT="${PLAT}-arm64" ;;
  x86_64|amd64)  PLAT="${PLAT}-amd64" ;;
  *)
    echo "❌ Unsupported architecture: $ARCH"
    echo "   Download manually: https://github.com/$REPO/releases"
    exit 1
    ;;
esac

# Only linux-amd64 is supported for Linux at this time
if [ "$OS" = "linux" ] && [ "$PLAT" != "linux-amd64" ]; then
  echo "❌ Linux/$ARCH is not yet supported."
  echo "   Download manually: https://github.com/$REPO/releases"
  exit 1
fi

# ── Resolve latest version if not specified ──────────────────────────────────

if [ -z "$VERSION" ]; then
  echo "🔍 Fetching latest release..."
  VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
fi

if [ -z "$VERSION" ]; then
  echo "❌ Could not determine latest version."
  echo "   Set VERSION= to specify manually, e.g.: VERSION=v0.1.0 sh install.sh"
  exit 1
fi

# ── Download and extract ─────────────────────────────────────────────────────

ARCHIVE="ai-agent-telemetry-${PLAT}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"

echo "📦 Downloading $VERSION for $PLAT..."

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -fsSL "$URL" -o "$TMP_DIR/$ARCHIVE"
tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

mkdir -p "$INSTALL_DIR"
mv "$TMP_DIR/ai-log" "$INSTALL_DIR/ai-log"
mv "$TMP_DIR/ai-log-report" "$INSTALL_DIR/ai-log-report"
chmod +x "$INSTALL_DIR/ai-log" "$INSTALL_DIR/ai-log-report"

# ── Done ─────────────────────────────────────────────────────────────────────

echo "✅ Installed ai-log and ai-log-report to $INSTALL_DIR"

if ! echo ":$PATH:" | grep -q ":$INSTALL_DIR:"; then
  echo ""
  echo "⚠️  $INSTALL_DIR is not in your PATH. Add this to your shell profile:"
  echo "   export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo ""
echo "🏁 Run 'ai-log init' to initialise the database."
echo "💡 To inject agent prompts globally:"
echo "   curl -fsSL https://raw.githubusercontent.com/$REPO/main/scripts/install-global.sh | sh"
