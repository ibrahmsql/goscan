#!/bin/bash
set -e

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [[ "$ARCH" == "x86_64" ]]; then
  ARCH="amd64"
elif [[ "$ARCH" == "arm64" || "$ARCH" == "aarch64" ]]; then
  ARCH="arm64"
fi

ASSET_NAME="goscan-${OS}-${ARCH}"

echo "Detected system: $OS-$ARCH"
echo "Looking for asset: $ASSET_NAME"

latest_json=$(curl -s "https://api.github.com/repos/isa-programmer/goscan/releases/latest")

asset_url=$(echo "$latest_json" | jq -r --arg NAME "$ASSET_NAME" \
  '.assets[] | select(.name == $NAME) | .browser_download_url')

if [[ -z "$asset_url" || "$asset_url" == "null" ]]; then
  echo "Error: No asset found for $ASSET_NAME in the latest release."
  exit 1
fi

echo "Downloading $ASSET_NAME ..."
curl -L -o "$ASSET_NAME" "$asset_url"
chmod +x "$ASSET_NAME"

echo "Moving $ASSET_NAME to /usr/local/bin/ (requires sudo)..."
sudo mv "$ASSET_NAME" /usr/local/bin/goscan

echo "Done! You can now run 'goscan --help'"
