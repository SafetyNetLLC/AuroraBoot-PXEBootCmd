#!/bin/bash
# build-go.sh: Build the Go auroraboot binary for the current platform
# Usage: ./build-go.sh [--arch amd64|arm64] [--output auroraboot]

ARCH=amd64
OUTPUT=auroraboot

if [[ "$1" == "-h" || "$1" == "--help" ]]; then
  echo "Usage: $0 [--arch amd64|arm64] [--output filename]"
  echo "  --arch     Set target architecture (default: amd64)"
  echo "  -h, --help Show this help message"
  exit 0
fi

while [[ $# -gt 0 ]]; do
  case $1 in
    --arch)
      ARCH="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

export GOARCH=$ARCH

echo "Building Go binary for $GOARCH..."
go build -o auroraboot

if [[ $? -eq 0 ]]; then
  echo "Build successful: $OUTPUT ($GOARCH)"
else
  echo "Build failed"
  exit 1
fi
