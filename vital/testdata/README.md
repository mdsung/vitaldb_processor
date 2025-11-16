# Test Data Directory

This `testdata/` directory contains test fixtures and sample data files for the VitalDB processor tests.

## Purpose

The `testdata` directory is a special name in Go's testing framework. The Go toolchain automatically ignores this directory during builds, making it the idiomatic location for test data.

## Usage in Tests

```go
// Access test data from test files
func TestVitalFileParser(t *testing.T) {
    vf, err := vital.NewVitalFile("testdata/sample.vital")
    if err != nil {
        t.Fatal(err)
    }
    // ... test assertions
}
```

When running `go test`, the current working directory is set to the package directory, so `testdata/` is always accessible with a simple relative path.

## File Organization

Planned structure:
```
testdata/
├── README.md                    # This file
├── small_sample.vital          # Small test file (< 100KB)
├── medium_sample.vital         # Medium test file (~1MB)
├── golden/                     # Expected output files
│   ├── small_sample.json       # Expected JSON output
│   └── small_sample.msgpack    # Expected MessagePack output
└── mock/                       # Mock data generators
    └── (future: programmatically generated test data)
```

## Golden File Testing

For regression testing, we use the golden file pattern:
1. Input file: `testdata/input.vital`
2. Expected output: `testdata/golden/input.json`
3. Test compares actual output with golden file

## Adding Test Data

### Small Sample Files
- Keep files under 1MB for fast tests
- Include variety of track types (WAVE, NUMERIC, STRING)
- Document the source and characteristics

### Golden Files
- Generate using verified correct implementation
- Update when intentional output format changes
- Version control to catch unintended changes

## Notes

- **Security**: Do not commit real patient data or sensitive information
- **Size**: Keep total testdata directory under 10MB
- **Format**: Prefer small, focused test files over large comprehensive ones
- **Documentation**: Add comments in test files explaining what each fixture tests

## References

- [Go Testing Best Practices](https://dave.cheney.net/2016/05/10/test-fixtures-in-go)
- [File-driven Testing in Go](https://eli.thegreenplace.net/2022/file-driven-testing-in-go/)
