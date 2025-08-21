# Benchmarking Guide for KSight

This guide explains how to use the comprehensive benchmarking system for KSight's performance testing.

## Quick Start

```bash
# Run the main benchmark (SQLite cache with 122K resources)
make benchmark

# Generate a summary report
make benchmark-report

# View all available benchmark commands
make help
```

## Available Benchmark Commands

### 🚀 **Core Benchmarks**

| Command | Description | Duration | Use Case |
|---------|-------------|----------|----------|
| `make benchmark` | Main SQLite cache benchmark | ~1-2 min | Quick performance validation |
| `make benchmark-all` | Comprehensive benchmark suite | 15-30 min | Full system performance analysis |
| `make benchmark-ci` | Fast benchmark for CI/CD | ~2-5 min | Automated testing pipelines |

### 📊 **Specific Component Benchmarks**

| Command | Component | Focus |
|---------|-----------|-------|
| `make benchmark-cache` | SQLite Database Cache | 122K resources load & write performance |
| `make benchmark-informer` | InformerManager | Kubernetes informer operations |
| `make benchmark-service` | Service Layer | API service layer performance |

### 🔍 **Profiling & Analysis**

| Command | Type | Output |
|---------|------|--------|
| `make benchmark-memory` | Memory Profiling | Memory usage patterns & allocation |
| `make benchmark-cpu` | CPU Profiling | CPU usage & bottlenecks |
| `make benchmark-analyze-memory` | Memory Analysis | Opens pprof web UI (port 8080) |
| `make benchmark-analyze-cpu` | CPU Analysis | Opens pprof web UI (port 8081) |

### 📈 **Reporting & Management**

| Command | Purpose |
|---------|---------|
| `make benchmark-report` | Generate performance summary |
| `make benchmark-compare` | Compare with baseline (regression testing) |
| `make benchmark-clean` | Clean all benchmark results |

## Benchmark Test Dataset

The main benchmark tests KSight's performance with a realistic large Kubernetes cluster:

- **100,000 Pods** - Typical large cluster workload
- **10,000 Nodes** - Enterprise cluster size  
- **10,000 Services** - Service mesh scenarios
- **1,000 ConfigMaps** - Configuration management (sensitive)
- **1,000 Deployments** - Application deployments
- **Total: 122,000 resources**

## Understanding Results

### Performance Metrics

```bash
# Example benchmark output interpretation:
Pods: 31.19s total, 311.97 μs/resource, 3205 resources/sec
```

- **Total Time**: Time to process all resources of this type
- **μs/resource**: Microseconds per individual resource (lower is better)
- **Resources/sec**: Throughput rate (higher is better)

### Key Performance Indicators

| Metric | Good | Acceptable | Poor |
|--------|------|------------|------|
| **Load Time (122K resources)** | <30s | 30-60s | >60s |
| **Single Write Performance** | <500μs | 500μs-1ms | >1ms |
| **Memory per Resource** | <8KB | 8-15KB | >15KB |
| **Sensitive Resource Overhead** | <10% | 10-20% | >20% |

### Typical Results (Apple M4 Pro)

```
=== OVERALL PERFORMANCE ===
Total resources: 122,000
Total time: ~37 seconds
Average per resource: ~304 μs
Resources per second: 3,291
Memory usage: ~1GB allocated

=== BY RESOURCE TYPE ===
Pods (100K):      311.97 μs/resource (3,205/sec)
Nodes (10K):      269.89 μs/resource (3,705/sec) 
Services (10K):   259.74 μs/resource (3,850/sec)
ConfigMaps (1K):  289.98 μs/resource (3,448/sec) - Sensitive
Deployments (1K): ~317 μs/resource (~3,155/sec)
```

## Using Benchmark Results

### Performance Validation
```bash
# Run benchmark and check results
make benchmark
make benchmark-report

# Check summary for performance regression
cat benchmark-results/summary.txt
```

### Memory Analysis
```bash
# Generate memory profile
make benchmark-memory

# Analyze memory usage (opens web UI)
make benchmark-analyze-memory
# Visit http://localhost:8080 to view profile
```

### CPU Analysis  
```bash
# Generate CPU profile
make benchmark-cpu

# Analyze CPU usage (opens web UI)
make benchmark-analyze-cpu
# Visit http://localhost:8081 to view profile
```

### Regression Testing
```bash
# Create baseline
make benchmark-compare  # Creates baseline on first run

# Later, compare against baseline
make benchmark-compare  # Compares current vs baseline
```

## Integration with CI/CD

### GitHub Actions Example
```yaml
name: Performance Tests
on: [push, pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.24
    - name: Run CI Benchmark
      run: make benchmark-ci
    - name: Upload Results
      uses: actions/upload-artifact@v3
      with:
        name: benchmark-results
        path: benchmark-results/
```

### Performance Gates
```bash
# In CI pipeline, fail if performance degrades
make benchmark-ci
if grep -q "μs/resource.*[0-9][0-9][0-9][0-9]" benchmark-results/benchmark-ci.txt; then
  echo "Performance regression detected!"
  exit 1
fi
```

## Troubleshooting

### Common Issues

**Issue: Benchmark timeout**
```bash
# Increase timeout for large systems
BENCHMARK_TIME=60m make benchmark
```

**Issue: Database locked errors**
```bash
# Clean and retry
make benchmark-clean
make benchmark
```

**Issue: High memory usage**
```bash
# Run memory analysis
make benchmark-memory
make benchmark-analyze-memory
```

### Performance Tips

1. **SSD Storage**: Use SSD for better SQLite performance
2. **Memory**: Ensure sufficient RAM (4GB+ recommended)
3. **Clean Environment**: Run `make clean` before benchmarks
4. **Background Tasks**: Close resource-intensive applications

## Benchmark File Structure

```
benchmark-results/
├── benchmark-cache.txt       # Main cache benchmark results
├── benchmark-memory.txt      # Memory benchmark results  
├── benchmark-cpu.txt         # CPU benchmark results
├── summary.txt              # Generated performance summary
├── mem.prof                 # Memory profile data
├── cpu.prof                 # CPU profile data
├── benchmark-baseline.txt   # Baseline for comparisons
└── benchmark-current.txt    # Current results for comparison
```

## Advanced Usage

### Custom Resource Counts
Edit `cache_performance_test.go` to modify test dataset sizes:
```go
const (
    POD_COUNT        = 50000   // Reduce for faster testing
    NODE_COUNT       = 5000
    SERVICE_COUNT    = 5000
    CONFIGMAP_COUNT  = 500
    DEPLOYMENT_COUNT = 500
)
```

### Custom Benchmarks
Add new benchmark functions following the pattern:
```go
func BenchmarkCustomComponent(b *testing.B) {
    // Your custom benchmark code
}
```

## Best Practices

1. **Consistent Environment**: Run benchmarks on the same hardware
2. **Clean State**: Use `make benchmark-clean` between runs
3. **Multiple Runs**: Average results from multiple benchmark runs
4. **Baseline Tracking**: Maintain baseline results for comparison
5. **CI Integration**: Include benchmarks in automated testing
6. **Performance Budgets**: Set acceptable performance thresholds

This benchmarking system provides comprehensive performance validation for KSight's SQLite cache system and overall architecture.