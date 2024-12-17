#!/bin/bash

version=$(git describe --tags --abbrev=0)
checksums_file="kubectl-crashwatch_${version#v}_checksums.txt"

platforms=(
  "linux_amd64"
  "linux_386" 
  "linux_arm64"
  "darwin_amd64"
  "darwin_arm64"
  "windows_amd64"
  "windows_386"
)

for platform in "${platforms[@]}"; do
  sha256=$(grep "${platform}" "$checksums_file" | cut -d ' ' -f 1)
  sed -i "s/\($platform.*sha256: \).*/\1\"$sha256\"/" krew.yaml
done