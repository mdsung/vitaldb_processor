.PHONY: test test-unit test-integration test-all bench verify-linecount help

# Default target
help:
	@echo "VitalDB Processor - Test Commands"
	@echo "=================================="
	@echo "make test           - Run unit tests only (fast, no external dependencies)"
	@echo "make test-integration - Run integration tests (requires .vital files)"
	@echo "make test-all       - Run both unit and integration tests"
	@echo "make bench          - Run benchmarks"
	@echo "make verify-linecount - Check that test files are under line limits"
	@echo ""

# Run unit tests only (no external file dependencies)
test-unit:
	@echo "Running unit tests..."
	go test ./vital -v

# Run integration tests (requires build tag and external files)
test-integration:
	@echo "Running integration tests..."
	go test -tags=integration ./vital -v

# Run both unit and integration tests
test-all: test-unit test-integration

# Alias: 'test' runs unit tests by default
test: test-unit

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. ./vital

# Verify that test files are under line count limits
verify-linecount:
	@echo "Checking line counts for test files..."
	@echo "Target: < 200 lines per file"
	@echo ""
	@wc -l vital/helper_test.go vital/unit_test.go vital/integration_test.go vital/benchmark_test.go | tail -1
	@echo ""
	@echo "Individual file counts:"
	@wc -l vital/helper_test.go vital/unit_test.go vital/integration_test.go vital/benchmark_test.go | grep -v total
	@echo ""
	@if [ $$(wc -l < vital/unit_test.go) -gt 200 ]; then \
		echo "WARNING: unit_test.go exceeds 200 lines"; \
	fi
	@if [ $$(wc -l < vital/integration_test.go) -gt 200 ]; then \
		echo "NOTE: integration_test.go is $$(wc -l < vital/integration_test.go) lines (integration tests may be longer)"; \
	fi
	@if [ $$(wc -l < vital/benchmark_test.go) -gt 200 ]; then \
		echo "WARNING: benchmark_test.go exceeds 200 lines"; \
	fi
	@if [ $$(wc -l < vital/helper_test.go) -gt 200 ]; then \
		echo "WARNING: helper_test.go exceeds 200 lines"; \
	fi
