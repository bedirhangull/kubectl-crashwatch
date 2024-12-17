#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/.."

# Get the latest tag
VERSION=$(git describe --tags --abbrev=0)

KREW_MANIFEST="deploy/krew/plugin.yaml"

# Ensure the manifest exists
if [[ ! -f "$KREW_MANIFEST" ]]; then
    echo "Krew manifest not found at $KREW_MANIFEST"
    exit 1
fi

CHECKSUMS_FILE="dist/kubectl-crashwatch_${VERSION#v}_checksums.txt"

if [[ ! -f "$CHECKSUMS_FILE" ]]; then
    echo "Checksums file not found: $CHECKSUMS_FILE"
    exit 1
fi

# Platforms to update
PLATFORMS=(
  "linux_amd64"
  "linux_386"
  "linux_arm64"
  "darwin_amd64"
  "darwin_arm64"
  "windows_amd64"
  "windows_386"
)

for platform in "${PLATFORMS[@]}"; do
    SHA256=$(grep "${platform}" "$CHECKSUMS_FILE" | cut -d ' ' -f 1)
    
    URI="https://github.com/bedirhangull/kubectl-crashwatch/releases/download/${VERSION}/kubectl-crashwatch_${platform}.tar.gz"
    
    # Update the manifest using yq for YAML manipulation
    yq eval -i "
      .spec.version = \"${VERSION}\" |
      .spec.platforms[] | select(.selector.matchLabels.os == \"$(echo $platform | cut -d'_' -f1)\" and .selector.matchLabels.arch == \"$(echo $platform | cut -d'_' -f2)\") |
      .uri = \"${URI}\" |
      .sha256 = \"${SHA256}\"
    " "$KREW_MANIFEST"
done

echo "Krew manifest updated successfully for ${VERSION}"