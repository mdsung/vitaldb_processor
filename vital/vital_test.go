package vital

// This file has been split into separate test files for better organization:
// - helper_test.go: Common helper functions
// - unit_test.go: Unit tests (no external dependencies, will be populated in Task #2)
// - integration_test.go: Integration tests (requires -tags=integration flag and real .vital files)
// - benchmark_test.go: Performance benchmarks
//
// To run tests:
//   go test ./vital              # Runs unit tests only
//   go test -tags=integration ./vital   # Runs integration tests
//   go test -bench=. ./vital     # Runs benchmarks
//
// Note: This file is kept for backward compatibility but will be removed in a future version.
// All test functionality has been moved to the appropriate test files.
