#!/bin/bash

# Build script for Terraform Provider for Shodan
# This script builds the provider for multiple platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Version from git tag or default
VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
echo -e "${GREEN}Building Terraform Provider for Shodan version: ${VERSION}${NC}"

# Create build directory
BUILD_DIR="build"
mkdir -p $BUILD_DIR

# Build targets
TARGETS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# Build for each target
for target in "${TARGETS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$target"
    
    echo -e "${YELLOW}Building for ${GOOS}/${GOARCH}...${NC}"
    
    # Set environment variables
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    export CGO_ENABLED=0
    
    # Build binary
    go build -ldflags="-s -w -X main.version=${VERSION}" -o terraform-provider-shodan
    
    # Create output filename
    if [ "$GOOS" = "windows" ]; then
        BINARY_NAME="terraform-provider-shodan_${VERSION}_${GOOS}_${GOARCH}.exe"
        ARCHIVE_NAME="terraform-provider-shodan_${VERSION}_${GOOS}_${GOARCH}.zip"
    else
        BINARY_NAME="terraform-provider-shodan_${VERSION}_${GOOS}_${GOARCH}"
        ARCHIVE_NAME="terraform-provider-shodan_${VERSION}_${GOOS}_${GOARCH}.tar.gz"
    fi
    
    # Move binary to build directory
    mv terraform-provider-shodan $BUILD_DIR/$BINARY_NAME
    
    # Create archive
    cd $BUILD_DIR
    if [ "$GOOS" = "windows" ]; then
        zip $ARCHIVE_NAME $BINARY_NAME
    else
        tar -czf $ARCHIVE_NAME $BINARY_NAME
    fi
    cd ..
    
    echo -e "${GREEN}✓ Built ${BINARY_NAME}${NC}"
done

# Create checksums
echo -e "${YELLOW}Creating checksums...${NC}"
cd $BUILD_DIR
sha256sum *.tar.gz *.zip > checksums.txt
cd ..

echo -e "${GREEN}✓ Build complete! Binaries are in the ${BUILD_DIR} directory${NC}"
echo -e "${GREEN}✓ Checksums file created: ${BUILD_DIR}/checksums.txt${NC}"

# List built files
echo -e "\n${YELLOW}Built files:${NC}"
ls -la $BUILD_DIR/
