#!/bin/bash

# Usage: ./build.sh [--swagger true|false] [--arch amd64|arm64] [--tag myimage:latest]
# Defaults: swagger=true, arch=amd64, tag=auroraboot:latest

SWAGGER=true
ARCH=amd64
TAG=quay.io/kairos/auroraboot:local

if [[ "$1" == "-h" || "$1" == "--help" ]]; then
  echo "Usage: $0 [--swagger true|false] [--arch amd64|arm64] [--tag myimage:latest]"
  echo "  --swagger   Enable or disable Swagger docs (default: true)"
  echo "  --arch      Set target architecture (default: amd64)"
  echo "  --tag       Set image tag (default: auroraboot:latest)"
  echo "  -h, --help  Show this help message"
  exit 0
fi

while [[ $# -gt 0 ]]; do
  case $1 in
    --swagger)
      SWAGGER="$2"
      shift 2
      ;;
    --arch)
      ARCH="$2"
      shift 2
      ;;
    --tag)
      TAG="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

if [[ "$SWAGGER" == "true" ]]; then
  SWAGGER_STAGE=with-swagger
else
  SWAGGER_STAGE=without-swagger
fi

echo "Building Docker image with SWAGGER_STAGE=$SWAGGER_STAGE, TARGETARCH=$ARCH, TAG=$TAG"
docker build --no-cache \
  --build-arg SWAGGER_STAGE=$SWAGGER_STAGE \
  --build-arg TARGETARCH=$ARCH \
  -t $TAG . | tee build.log

echo "\n--- Docker build log (last 40 lines) ---"
tail -n 40 build.log

rm -f build.log
