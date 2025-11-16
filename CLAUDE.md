# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

VitalDB Processor is a Go library that reads and processes VitalDB files (.vital), aiming to provide better performance than the Python VitalDB SDK while maintaining **100% compatibility** with Python VitalDB results.

**Critical Design Principle**: Python VitalDB is the **golden standard**. The Go implementation must produce identical results to Python VitalDB, just faster. Any discrepancy in output is a bug in the Go implementation that needs to be fixed.

**Goals**:
1. **Accuracy**: Match Python VitalDB output exactly (golden standard)
2. **Performance**: Process files faster than Python VitalDB
3. **Reliability**: Handle all VitalDB file formats that Python VitalDB supports

## Build and Test Commands

### Building

```bash
# Build the CLI tool
cd example
go build -o vitaldb_processor main.go

# Run the built binary
./vitaldb_processor [options] <vital_file_path>
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests in vital package
go test -v ./vital

# Run integration tests (when available)
go test -tags=integration ./vital

# Run benchmarks
go test -bench=. ./vital
```

### Common CLI Usage

```bash
# Basic file info
./vitaldb_processor -info-only -quiet data.vital

# JSON output (for Python integration)
./vitaldb_processor -format json -max-tracks 0 data.vital

# Filter specific tracks
./vitaldb_processor -tracks "ECG_II,HR,PLETH" data.vital

# Filter by track type
./vitaldb_processor -track-type WAVE data.vital

# Time range extraction
./vitaldb_processor -start-time 0 -end-time 300 data.vital
```

## Architecture

### Package Structure

```
vitaldb_processor/
├── vital/                    # Core library package
│   ├── vital.go             # Main entry point (NewVitalFile)
│   ├── types.go             # Core data structures
│   ├── parser.go            # Binary packet parsing
│   ├── devinfo.go           # Device info parsing (packet type 9)
│   ├── trkinfo.go           # Track info parsing (packet type 0)
│   ├── rec.go               # Record data parsing (packet type 1)
│   ├── cmd.go               # Command parsing (packet type 6)
│   └── util.go              # Helper functions
└── example/
    └── main.go              # CLI tool with extensive options
```

### Core Data Flow

1. **File Reading** (`vital.go:NewVitalFile`):
   - Opens `.vital` file (gzip compressed)
   - Validates magic header "VITA"
   - Reads file header (version, timestamps, GMT offset)
   - Iterates through binary packets

2. **Packet Processing** (`vital.go:72-110`):
   - Packet type 9: Device information (`devinfo.go`)
   - Packet type 0: Track metadata (`trkinfo.go`)
   - Packet type 1: Actual data records (`rec.go`)
   - Packet type 6: Commands (`cmd.go`)

3. **Data Structures** (`types.go`):
   - `VitalFile`: Root container with devices, tracks, timestamps
   - `Device`: Medical equipment metadata (name, type, port)
   - `Track`: Data stream metadata (sample rate, unit, gain, offset)
   - `Rec`: Individual data points (timestamp, value)

### Key Data Types

**Track Types:**
- Type 1: WAVE data (continuous waveforms like ECG)
- Type 2: NUMERIC data (discrete values like heart rate)
- Type 5: STRING data (text annotations)

**Data Formats:**
Tracks use different binary formats (`Track.Fmt` field) that determine how `Rec.Val` is stored:
- Float32/Float64 arrays for waveforms
- Single numeric values for vitals
- Strings for annotations

### Python Compatibility (CRITICAL)

**Python VitalDB is the golden standard**. The Go implementation must match Python VitalDB behavior exactly:

- **EOF Handling**: Handles incomplete packets at EOF gracefully (`vital.go:89-95`) to match Python behavior
- **Permissive Parsing**: Matches Python's lenient approach to invalid/incomplete packets
- **Output Verification**: All outputs should be validated against Python VitalDB to ensure identical results
- **JSON Output**: Provides JSON for easy comparison with Python VitalDB results

**Testing Strategy**: Always compare Go output against Python VitalDB output on the same file. Any difference indicates a bug in the Go implementation.

## Important Implementation Details

### Binary Parsing

All data is little-endian encoded. The parser uses `encoding/binary` for safe deserialization. Packet length validation prevents memory issues (100MB limit per packet).

### Error Handling Strategy

The codebase follows Python VitalDB's lenient approach:
- EOF errors during packet reading are suppressed (incomplete final packets)
- Header length is validated before accessing fields
- Invalid packets are skipped rather than failing the entire file

### Memory Efficiency

Data is parsed on-demand:
- CLI tool supports filtering by track name, type, and time range
- `MaxTracks` and `MaxSamples` options control output size
- Only requested data is included in output

### CLI Architecture (`example/main.go`)

The CLI tool separates concerns:
- `parseFlags()`: Command-line argument parsing
- `processVitalFile()`: Business logic for filtering/transforming data
- `processTracks()`: Track filtering by name, type, time range
- `printTextOutput()` vs JSON output: Different output formats

## Development Workflow

### Adding New Features

1. Update core types in `vital/types.go` if needed
2. Implement parsing logic in appropriate parser file
3. Update CLI in `example/main.go` if exposing to users
4. Add tests in `vital/vital_test.go`

### Testing Strategy (From TODO.md)

Current test file is too large (440 lines). Planned improvements:
- Split into `unit_test.go`, `integration_test.go`, `benchmark_test.go`
- Create `testdata/` directory with embedded test files
- Use table-driven tests for parser validation
- Add mock data generation functions

### Performance Considerations

- Avoid unnecessary allocations in hot paths
- Use binary.Read sparingly (manual byte slicing is faster)
- Consider memory pooling for large waveform arrays
- Profile before optimizing (`go test -bench=. -cpuprofile=cpu.prof`)

## Known Issues and Quirks

### Python VitalDB Buffer Error (CRITICAL)

**⚠️ CRITICAL**: When using Python VitalDB library directly, you MUST apply this fix to avoid "buffer is too small" errors:

```python
import vitaldb

# REQUIRED: Fix buffer errors for format types 7 and 8
vitaldb.utils.FMT_TYPE_LEN[7] = ("i", 4)  # signed int, 4 bytes
vitaldb.utils.FMT_TYPE_LEN[8] = ("I", 4)  # unsigned int, 4 bytes

# Now safe to use
vf = vitaldb.VitalFile('data.vital')
```

**Root Cause**: Python VitalDB library has incorrect/missing buffer size definitions for format types 7 (signed integer) and 8 (unsigned integer).

**Impact**: Without this fix, files containing tracks with these format types will fail to load with buffer errors.

**Go Implementation**: The Go VitalDB Processor handles all format types correctly without any patches needed. However, when validating Go output, always use Python VitalDB with the buffer fix applied, as Python VitalDB (with fix) is the golden standard.

### EOF Handling

Some VitalDB files have incomplete packets at EOF. The library matches Python VitalDB by ignoring these (`vital.go:91-92`), which fixed compatibility with real-world files.

### File Path Dependencies

Tests currently hardcode paths like `../../data/sample_vitalfiles/`. This is a known issue tracked in TODO.md - tests should use embedded testdata instead.

### Type Safety vs Flexibility

`Rec.Val` is `any` because VitalDB tracks can contain various types (float32, float64, arrays, strings). Users must type-assert based on `Track.Type` and `Track.Fmt`.

## Python Integration Pattern

### Option 1: Using Go Binary (Recommended)

The recommended approach leverages Go's performance and reliability:

```python
import subprocess
import json

def load_vital_data(file_path, **kwargs):
    cmd = ['./vitaldb_processor', '-format', 'json']
    if 'tracks' in kwargs:
        cmd.extend(['-tracks', ','.join(kwargs['tracks'])])
    cmd.append(file_path)

    result = subprocess.run(cmd, capture_output=True, text=True)
    return json.loads(result.stdout)
```

This leverages Go's performance while maintaining Python's ease of use for analysis.

### Option 2: Using Python VitalDB Library Directly

**CRITICAL**: If using Python VitalDB library directly, you **MUST** apply this buffer fix first:

```python
import vitaldb

# REQUIRED: Fix buffer errors in Python VitalDB library
vitaldb.utils.FMT_TYPE_LEN[7] = ("i", 4)
vitaldb.utils.FMT_TYPE_LEN[8] = ("I", 4)

# Now safe to load VitalDB files
vf = vitaldb.VitalFile('data.vital')
```

**Why this is needed**: The Python VitalDB library has a known buffer size issue with certain track formats (types 7 and 8). Without this fix, you'll encounter "buffer is too small" errors. The Go implementation handles these formats correctly without any patches.

**For Production Use**: Go binary approach (Option 1) is faster and more convenient.

**For Validation/Testing**: Always use Python VitalDB (with buffer fix) as the golden standard to verify Go implementation correctness.

## File Format Notes

VitalDB files are gzip-compressed binary files with structure:
```
[GZIP Header]
  "VITA" magic (4 bytes)
  version (uint32)
  header_length (uint16)
  header_data (variable)
  [packets until EOF]
    packet_type (uint8)
    packet_length (uint32)
    packet_data (variable)
```

Understanding this format is crucial when debugging parsing issues.

## Task Master AI Instructions
**Import Task Master's development workflow commands and guidelines, treat as if import is in the main CLAUDE.md file.**
@./.taskmaster/CLAUDE.md
