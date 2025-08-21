# Makefile for KSight project

.PHONY: test test-unit test-integration setup-envtest clean build dev

# Variables
ENVTEST_K8S_VERSION = 1.28.3
ENVTEST_ASSETS_DIR = $(shell pwd)/bin/k8s
GINKGO = go run github.com/onsi/ginkgo/v2/ginkgo

# Default target
all: build

# Build the application
build:
	wails build

# Run in development mode
dev:
	wails dev

# Setup test environment binaries
setup-envtest:
	@echo "Setting up test environment..."
	@mkdir -p $(ENVTEST_ASSETS_DIR)
	@test -f $(ENVTEST_ASSETS_DIR)/etcd && test -f $(ENVTEST_ASSETS_DIR)/kube-apiserver && test -f $(ENVTEST_ASSETS_DIR)/kubectl || \
	(echo "Downloading Kubernetes test binaries..." && \
	 KUBEBUILDER_ASSETS=$(ENVTEST_ASSETS_DIR) go run sigs.k8s.io/controller-runtime/tools/setup-envtest@latest use $(ENVTEST_K8S_VERSION) --bin-dir $(ENVTEST_ASSETS_DIR))

# Run all tests
test: setup-envtest test-unit test-integration

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v ./pkg/informer ./pkg/service -race -coverprofile=coverage-unit.out

# Run integration tests with Ginkgo
test-integration: setup-envtest
	@echo "Running integration tests..."
	@export KUBEBUILDER_ASSETS=$(ENVTEST_ASSETS_DIR) && \
	$(GINKGO) -v --race --randomize-all --randomize-suites \
		--cover --coverprofile=coverage-integration.out \
		--output-dir=test-results \
		--json-report=test-results/integration-report.json \
		--junit-report=test-results/integration-junit.xml \
		./pkg/test

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage-integration.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run specific test suite
test-informer: setup-envtest
	@export KUBEBUILDER_ASSETS=$(ENVTEST_ASSETS_DIR) && \
	$(GINKGO) -v --focus="InformerManager" ./pkg/test

test-service: setup-envtest
	@export KUBEBUILDER_ASSETS=$(ENVTEST_ASSETS_DIR) && \
	$(GINKGO) -v --focus="ClusterService" ./pkg/test

test-persistence: setup-envtest
	@export KUBEBUILDER_ASSETS=$(ENVTEST_ASSETS_DIR) && \
	$(GINKGO) -v --focus="ResourceVersion Persistence" ./pkg/test

# Clean up test artifacts and build files
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf build/bin/
	rm -rf test-results/
	rm -f coverage*.out coverage.html
	go clean -testcache

# Install test dependencies
deps-test:
	go mod download
	go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...
	goimports -w .

# Generate mocks (if needed in the future)
generate:
	go generate ./...

# Run tests in watch mode for development
test-watch: setup-envtest
	@export KUBEBUILDER_ASSETS=$(ENVTEST_ASSETS_DIR) && \
	$(GINKGO) watch -v ./pkg/test

# Quick test run (skip setup if binaries exist)
test-quick:
	@if [ -f $(ENVTEST_ASSETS_DIR)/etcd ]; then \
		$(MAKE) test-integration; \
	else \
		$(MAKE) test; \
	fi

# Help target
help:
	@echo "Available targets:"
	@echo "  build           - Build the Wails application"
	@echo "  dev             - Run in development mode"
	@echo "  test            - Run all tests (unit + integration)"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration - Run integration tests with real K8s API"
	@echo "  test-coverage   - Run tests and generate coverage report"
	@echo "  test-informer   - Run InformerManager tests only"
	@echo "  test-service    - Run ClusterService tests only"
	@echo "  test-persistence - Run persistence tests only"
	@echo "  test-watch      - Run tests in watch mode"
	@echo "  test-quick      - Quick test run (skip setup if possible)"
	@echo "  setup-envtest   - Download K8s test binaries"
	@echo "  clean           - Clean up build and test artifacts"
	@echo "  deps-test       - Install test dependencies"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  help            - Show this help message"
