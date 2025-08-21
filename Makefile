# Makefile for KSight project

.PHONY: test test-unit test-integration setup-envtest clean build dev benchmark benchmark-all benchmark-cache benchmark-informer benchmark-service benchmark-memory benchmark-cpu benchmark-report

# Variables
ENVTEST_K8S_VERSION = 1.28.3
ENVTEST_ASSETS_DIR = $(shell pwd)/bin/k8s
GINKGO = go run github.com/onsi/ginkgo/v2/ginkgo
BENCHMARK_DIR = ./pkg/test
BENCHMARK_RESULTS_DIR = benchmark-results
BENCHMARK_TIME = 30m
BENCHMARK_MEM_PROFILE = $(BENCHMARK_RESULTS_DIR)/mem.prof
BENCHMARK_CPU_PROFILE = $(BENCHMARK_RESULTS_DIR)/cpu.prof

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
	$(GINKGO) -v -p --race --randomize-all --randomize-suites \
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
	rm -rf $(BENCHMARK_RESULTS_DIR)
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

# ============================================================================
# BENCHMARK TARGETS
# ============================================================================

# Prepare benchmark environment
benchmark-setup:
	@echo "Setting up benchmark environment..."
	@mkdir -p $(BENCHMARK_RESULTS_DIR)
	@echo "Benchmark results will be stored in: $(BENCHMARK_RESULTS_DIR)"

# Run all benchmark tests
benchmark-all: benchmark-setup
	@echo "Running comprehensive benchmark suite..."
	@echo "This may take 15-30 minutes depending on your system..."
	@go test -run='^$$' -bench=. -benchmem -timeout=$(BENCHMARK_TIME) \
		-cpuprofile=$(BENCHMARK_CPU_PROFILE) \
		-memprofile=$(BENCHMARK_MEM_PROFILE) \
		$(BENCHMARK_DIR) | tee $(BENCHMARK_RESULTS_DIR)/benchmark-all.txt
	@echo "All benchmarks completed. Results saved to $(BENCHMARK_RESULTS_DIR)/"

# SQLite cache performance benchmark (main performance test)
benchmark-cache: benchmark-setup
	@echo "Running SQLite cache performance benchmark..."
	@echo "Testing: 100K pods, 10K nodes, 10K services, 1K configmaps, 1K deployments"
	@go test -run='^$$' -bench=BenchmarkSQLitePerformance -benchmem -timeout=$(BENCHMARK_TIME) \
		$(BENCHMARK_DIR) | tee $(BENCHMARK_RESULTS_DIR)/benchmark-cache.txt
	@echo "Cache benchmark completed. Results saved to $(BENCHMARK_RESULTS_DIR)/benchmark-cache.txt"

# Informer manager benchmark
benchmark-informer: benchmark-setup
	@echo "Running InformerManager benchmark..."
	@go test -run='^$$' -bench=.*Informer.* -benchmem -timeout=10m \
		$(BENCHMARK_DIR) | tee $(BENCHMARK_RESULTS_DIR)/benchmark-informer.txt
	@echo "Informer benchmark completed."

# Service layer benchmark
benchmark-service: benchmark-setup
	@echo "Running service layer benchmark..."
	@go test -run='^$$' -bench=.*Service.* -benchmem -timeout=10m \
		$(BENCHMARK_DIR) | tee $(BENCHMARK_RESULTS_DIR)/benchmark-service.txt
	@echo "Service benchmark completed."

# Memory-focused benchmark
benchmark-memory: benchmark-setup
	@echo "Running memory benchmark with profiling..."
	@go test -run='^$$' -bench=BenchmarkSQLitePerformance/LoadPerformanceTest -benchmem \
		-memprofile=$(BENCHMARK_MEM_PROFILE) -timeout=15m \
		$(BENCHMARK_DIR) | tee $(BENCHMARK_RESULTS_DIR)/benchmark-memory.txt
	@echo "Memory benchmark completed. Profile: $(BENCHMARK_MEM_PROFILE)"

# CPU-focused benchmark
benchmark-cpu: benchmark-setup
	@echo "Running CPU benchmark with profiling..."
	@go test -run='^$$' -bench=BenchmarkSQLitePerformance/SingleWrites -benchmem \
		-cpuprofile=$(BENCHMARK_CPU_PROFILE) -timeout=10m \
		$(BENCHMARK_DIR) | tee $(BENCHMARK_RESULTS_DIR)/benchmark-cpu.txt
	@echo "CPU benchmark completed. Profile: $(BENCHMARK_CPU_PROFILE)"

# Quick benchmark (just essential performance tests)
benchmark: benchmark-cache

# Benchmark for continuous integration
benchmark-ci: benchmark-setup
	@echo "Running CI benchmark (quick version)..."
	@go test -run='^$$' -bench=BenchmarkSQLitePerformance/SingleWrites -benchmem -timeout=5m \
		$(BENCHMARK_DIR) | tee $(BENCHMARK_RESULTS_DIR)/benchmark-ci.txt
	@echo "CI benchmark completed."

# Generate benchmark report
benchmark-report: benchmark-setup
	@echo "Generating benchmark report..."
	@if [ -f $(BENCHMARK_RESULTS_DIR)/benchmark-cache.txt ]; then \
		echo "=== CACHE PERFORMANCE SUMMARY ===" > $(BENCHMARK_RESULTS_DIR)/summary.txt; \
		grep -E "(Pods|Nodes|Services|ConfigMaps|Deployments|Average|resources/sec)" $(BENCHMARK_RESULTS_DIR)/benchmark-cache.txt >> $(BENCHMARK_RESULTS_DIR)/summary.txt || true; \
		echo "" >> $(BENCHMARK_RESULTS_DIR)/summary.txt; \
		echo "=== DETAILED RESULTS ===" >> $(BENCHMARK_RESULTS_DIR)/summary.txt; \
		cat $(BENCHMARK_RESULTS_DIR)/benchmark-cache.txt >> $(BENCHMARK_RESULTS_DIR)/summary.txt; \
	else \
		echo "No benchmark results found. Run 'make benchmark' first."; \
	fi
	@echo "Report generated: $(BENCHMARK_RESULTS_DIR)/summary.txt"

# Analyze memory profile (requires go tool pprof)
benchmark-analyze-memory:
	@if [ -f $(BENCHMARK_MEM_PROFILE) ]; then \
		echo "Opening memory profile analysis..."; \
		go tool pprof -http=:8080 $(BENCHMARK_MEM_PROFILE); \
	else \
		echo "Memory profile not found. Run 'make benchmark-memory' first."; \
	fi

# Analyze CPU profile (requires go tool pprof)
benchmark-analyze-cpu:
	@if [ -f $(BENCHMARK_CPU_PROFILE) ]; then \
		echo "Opening CPU profile analysis..."; \
		go tool pprof -http=:8081 $(BENCHMARK_CPU_PROFILE); \
	else \
		echo "CPU profile not found. Run 'make benchmark-cpu' first."; \
	fi

# Compare benchmarks (for performance regression testing)
benchmark-compare: benchmark-setup
	@echo "Running baseline benchmark for comparison..."
	@if [ -f $(BENCHMARK_RESULTS_DIR)/benchmark-baseline.txt ]; then \
		go test -run='^$$' -bench=BenchmarkSQLitePerformance -benchmem \
			$(BENCHMARK_DIR) > $(BENCHMARK_RESULTS_DIR)/benchmark-current.txt; \
		echo "Comparing with baseline..."; \
		benchcmp $(BENCHMARK_RESULTS_DIR)/benchmark-baseline.txt $(BENCHMARK_RESULTS_DIR)/benchmark-current.txt || \
		echo "benchcmp not available. Install with: go install golang.org/x/tools/cmd/benchcmp@latest"; \
	else \
		echo "No baseline found. Creating baseline..."; \
		go test -run='^$$' -bench=BenchmarkSQLitePerformance -benchmem \
			$(BENCHMARK_DIR) | tee $(BENCHMARK_RESULTS_DIR)/benchmark-baseline.txt; \
		echo "Baseline created at $(BENCHMARK_RESULTS_DIR)/benchmark-baseline.txt"; \
	fi

# Clean benchmark results
benchmark-clean:
	@echo "Cleaning benchmark results..."
	@rm -rf $(BENCHMARK_RESULTS_DIR)
	@echo "Benchmark results cleaned."

# ============================================================================

# Help target
help:
	@echo "Available targets:"
	@echo ""
	@echo "BUILD & DEVELOPMENT:"
	@echo "  build           - Build the Wails application"
	@echo "  dev             - Run in development mode"
	@echo "  clean           - Clean up build and test artifacts"
	@echo "  deps-test       - Install test dependencies"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo ""
	@echo "TESTING:"
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
	@echo ""
	@echo "BENCHMARKING:"
	@echo "  benchmark       - Run SQLite cache benchmark (100K pods, 10K nodes, etc.)"
	@echo "  benchmark-all   - Run comprehensive benchmark suite (15-30 min)"
	@echo "  benchmark-cache - Run SQLite cache performance benchmark"
	@echo "  benchmark-informer - Run InformerManager benchmark"
	@echo "  benchmark-service - Run service layer benchmark"
	@echo "  benchmark-memory - Run memory benchmark with profiling"
	@echo "  benchmark-cpu   - Run CPU benchmark with profiling"
	@echo "  benchmark-ci    - Quick benchmark for continuous integration"
	@echo "  benchmark-report - Generate benchmark report summary"
	@echo "  benchmark-compare - Compare benchmark results (regression testing)"
	@echo "  benchmark-analyze-memory - Open memory profile in pprof (port 8080)"
	@echo "  benchmark-analyze-cpu - Open CPU profile in pprof (port 8081)"
	@echo "  benchmark-clean - Clean benchmark results"
	@echo ""
	@echo "EXAMPLES:"
	@echo "  make benchmark        # Quick cache performance test"
	@echo "  make benchmark-all    # Full benchmark suite"
	@echo "  make benchmark-memory # Memory profiling"
	@echo "  make benchmark-report # Generate summary report"
	@echo ""
	@echo "  help            - Show this help message"
