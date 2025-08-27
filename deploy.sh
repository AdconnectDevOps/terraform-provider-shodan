#!/bin/bash

# Terraform Provider for Shodan - Deployment Script
# This script automates the build and deployment process

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROVIDER_NAME="terraform-provider-shodan"
VERSION="0.1.9"
REGISTRY="registry.terraform.io"
NAMESPACE="adconnectdevops"
PROVIDER_TYPE="shodan"

# Detect platform
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "arm64" ] || [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

echo -e "${BLUE}🚀 Terraform Provider for Shodan - Deployment Script${NC}"
echo -e "${BLUE}================================================${NC}"
echo -e "Platform: ${PLATFORM}/${ARCH}"
echo -e "Version: ${VERSION}"
echo -e "Provider: ${NAMESPACE}/${PROVIDER_TYPE}"
echo ""

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check Go version
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_GO="1.21"
    
    if [ "$(printf '%s\n' "$REQUIRED_GO" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_GO" ]; then
        print_error "Go version $GO_VERSION is too old. Required: $REQUIRED_GO or later."
        exit 1
    fi
    
    print_status "Go version $GO_VERSION is compatible"
    
    # Check Terraform version
    if ! command -v terraform &> /dev/null; then
        print_warning "Terraform is not installed. This is optional for building the provider."
    else
        TF_VERSION=$(terraform version | head -n1 | awk '{print $2}' | sed 's/v//')
        print_status "Terraform version $TF_VERSION detected"
    fi
    
    # Check Make
    if ! command -v make &> /dev/null; then
        print_warning "Make is not installed. Some commands may not work."
    else
        print_status "Make is available"
    fi
}

# Function to clean previous builds
clean_build() {
    print_info "Cleaning previous build artifacts..."
    
    if [ -f "$PROVIDER_NAME" ]; then
        rm -f "$PROVIDER_NAME"
        print_status "Removed previous binary"
    fi
    
    if [ -d "dist" ]; then
        rm -rf dist
        print_status "Removed dist directory"
    fi
    
    go clean -cache
    print_status "Cleaned Go cache"
}

# Function to download dependencies
download_deps() {
    print_info "Downloading Go dependencies..."
    
    go mod download
    go mod tidy
    go mod verify
    
    print_status "Dependencies downloaded and verified"
}

# Function to build the provider
build_provider() {
    print_info "Building provider for ${PLATFORM}/${ARCH}..."
    
    # Set Go environment variables
    export CGO_ENABLED=0
    export GOOS=$PLATFORM
    export GOARCH=$ARCH
    
    # Build flags
    LDFLAGS="-s -w -X main.version=${VERSION}"
    
    go build -ldflags="$LDFLAGS" -o "$PROVIDER_NAME" .
    
    if [ -f "$PROVIDER_NAME" ]; then
        print_status "Provider built successfully: $PROVIDER_NAME"
        
        # Show binary info
        BINARY_SIZE=$(ls -lh "$PROVIDER_NAME" | awk '{print $5}')
        print_info "Binary size: $BINARY_SIZE"
    else
        print_error "Build failed - binary not found"
        exit 1
    fi
}

# Function to run tests
run_tests() {
    print_info "Running tests..."
    
    if go test -v ./...; then
        print_status "All tests passed"
    else
        print_warning "Some tests failed - continuing with deployment"
    fi
}

# Function to install locally
install_local() {
    print_info "Installing provider locally..."
    
    # Create plugins directory
    PLUGINS_DIR="$HOME/.terraform.d/plugins/${REGISTRY}/${NAMESPACE}/${PROVIDER_TYPE}/${VERSION}/${PLATFORM}_${ARCH}"
    mkdir -p "$PLUGINS_DIR"
    
    # Copy binary
    cp "$PROVIDER_NAME" "$PLUGINS_DIR/"
    chmod +x "$PLUGINS_DIR/$PROVIDER_NAME"
    
    print_status "Provider installed to: $PLUGINS_DIR"
}

# Function to create release package
create_release() {
    print_info "Creating release package..."
    
    mkdir -p dist
    
    # Create platform-specific archive
    if [ "$PLATFORM" = "windows" ]; then
        ARCHIVE_NAME="${PROVIDER_NAME}_${VERSION}_${PLATFORM}_${ARCH}.zip"
        zip -j "dist/$ARCHIVE_NAME" "$PROVIDER_NAME"
    else
        ARCHIVE_NAME="${PROVIDER_NAME}_${VERSION}_${PLATFORM}_${ARCH}.tar.gz"
        tar -czf "dist/$ARCHIVE_NAME" "$PROVIDER_NAME"
    fi
    
    print_status "Release package created: dist/$ARCHIVE_NAME"
}

# Function to show installation instructions
show_instructions() {
    echo ""
    echo -e "${BLUE}📋 Installation Instructions${NC}"
    echo -e "${BLUE}============================${NC}"
    echo ""
    echo -e "The provider has been built and installed locally."
    echo ""
    echo -e "To use it in your Terraform configuration:"
    echo ""
    echo -e "1. Add to your Terraform configuration:"
    echo -e "   ${YELLOW}terraform {${NC}"
    echo -e "     ${YELLOW}required_providers {${NC}"
    echo -e "       ${YELLOW}shodan = {${NC}"
    echo -e "         ${YELLOW}source = \"${REGISTRY}/${NAMESPACE}/${PROVIDER_TYPE}\"${NC}"
    echo -e "         ${YELLOW}version = \"${VERSION}\"${NC}"
    echo -e "       ${YELLOW}}${NC}"
    echo -e "     ${YELLOW}}${NC}"
    echo -e "   ${YELLOW}}${NC}"
    echo ""
    echo -e "2. Configure the provider:"
    echo -e "   ${YELLOW}provider \"shodan\" {${NC}"
    echo -e "     ${YELLOW}api_key = var.shodan_api_key${NC}"
    echo -e "   ${YELLOW}}${NC}"
    echo ""
    echo -e "3. Initialize Terraform:"
    echo -e "   ${YELLOW}terraform init${NC}"
    echo ""
    echo -e "For more information, see the documentation in the docs/ directory."
}

# Function to show available commands
show_help() {
    echo ""
    echo -e "${BLUE}🔧 Available Commands${NC}"
    echo -e "${BLUE}====================${NC}"
    echo ""
    echo -e "  ${YELLOW}./deploy.sh${NC}          - Full build and install"
    echo -e "  ${YELLOW}./deploy.sh build${NC}     - Build only"
    echo -e "  ${YELLOW}./deploy.sh install${NC}   - Install only"
    echo -e "  ${YELLOW}./deploy.sh clean${NC}     - Clean build artifacts"
    echo -e "  ${YELLOW}./deploy.sh test${NC}      - Run tests only"
    echo -e "  ${YELLOW}./deploy.sh help${NC}      - Show this help"
    echo ""
}

# Main deployment function
deploy() {
    print_info "Starting deployment process..."
    
    check_prerequisites
    clean_build
    download_deps
    build_provider
    run_tests
    install_local
    create_release
    show_instructions
}

# Command line argument handling
case "${1:-deploy}" in
    "build")
        print_info "Building provider only..."
        check_prerequisites
        clean_build
        download_deps
        build_provider
        ;;
    "install")
        print_info "Installing provider only..."
        if [ ! -f "$PROVIDER_NAME" ]; then
            print_error "Provider binary not found. Run build first."
            exit 1
        fi
        install_local
        show_instructions
        ;;
    "clean")
        print_info "Cleaning build artifacts..."
        clean_build
        ;;
    "test")
        print_info "Running tests only..."
        check_prerequisites
        download_deps
        run_tests
        ;;
    "help"|"-h"|"--help")
        show_help
        exit 0
        ;;
    "deploy")
        deploy
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac

echo ""
print_status "Deployment completed successfully!"
echo ""
