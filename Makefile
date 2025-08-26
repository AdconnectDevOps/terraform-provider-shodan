.PHONY: help build clean test install

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Terraform provider binary
	export GOROOT=/opt/homebrew/opt/go/libexec && GOOS=darwin GOARCH=arm64 go build -o terraform-provider-shodan

clean: ## Clean build artifacts
	rm -f terraform-provider-shodan
	go clean -cache

test: ## Run tests
	export GOROOT=/opt/homebrew/opt/go/libexec && go test -v ./...

install: build ## Build and install the provider locally
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/adconnectdevops/shodan/0.1.0/darwin_arm64/
	cp terraform-provider-shodan ~/.terraform.d/plugins/registry.terraform.io/adconnectdevops/shodan/0.1.0/darwin_arm64/

fmt: ## Format Go code
	export GOROOT=/opt/homebrew/opt/go/libexec && go fmt ./...

vet: ## Run go vet
	export GOROOT=/opt/homebrew/opt/go/libexec && go vet ./...

deps: ## Download dependencies
	export GOROOT=/opt/homebrew/opt/go/libexec && go mod tidy
	export GOROOT=/opt/homebrew/opt/go/libexec && go mod download

dev: ## Run in development mode with debug
	export GOROOT=/opt/homebrew/opt/go/libexec && go run . -debug
