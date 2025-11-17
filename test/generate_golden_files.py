#!/usr/bin/env python3
"""
Generate golden reference files from Python VitalDB library.

This script processes VitalDB files using the Python VitalDB library
and saves the results as JSON files. These golden files serve as the
reference standard for testing the Go implementation.

CRITICAL: Python VitalDB is the golden standard. The Go implementation
must produce identical results.
"""

import json
import os
import sys
from pathlib import Path

try:
    import vitaldb
except ImportError:
    print("Error: vitaldb module not found. Install with: pip install vitaldb", file=sys.stderr)
    sys.exit(1)

# CRITICAL: Apply buffer fix for Python VitalDB library
# Without this fix, files with format types 7 and 8 will fail with "buffer is too small" errors
vitaldb.utils.FMT_TYPE_LEN[7] = ("i", 4)  # signed int, 4 bytes
vitaldb.utils.FMT_TYPE_LEN[8] = ("I", 4)  # unsigned int, 4 bytes


def serialize_value(val):
    """Convert Python VitalDB value to JSON-serializable format."""
    if val is None:
        return None

    # Handle numpy arrays and lists
    if hasattr(val, 'tolist'):
        return val.tolist()

    # Handle bytes
    if isinstance(val, bytes):
        return val.decode('utf-8', errors='replace')

    # Handle basic types
    if isinstance(val, (int, float, str, bool)):
        return val

    # Handle lists
    if isinstance(val, list):
        return [serialize_value(v) for v in val]

    # Fallback: convert to string
    return str(val)


def extract_vital_data(vital_file_path):
    """
    Extract all data from a VitalDB file using Python VitalDB library.

    Returns a dictionary with file info, devices, and tracks.
    """
    print(f"Processing: {vital_file_path}")

    try:
        # Convert Path to string for VitalFile
        vf = vitaldb.VitalFile(str(vital_file_path))
    except Exception as e:
        print(f"Error loading {vital_file_path}: {e}", file=sys.stderr)
        return None

    # Extract file metadata
    file_info = {
        "dt_start": float(vf.dtstart) if hasattr(vf, 'dtstart') else 0.0,
        "dt_end": float(vf.dtend) if hasattr(vf, 'dtend') else 0.0,
        "gmt_offset": int(vf.dgmt) if hasattr(vf, 'dgmt') else 0,
    }
    file_info["duration"] = file_info["dt_end"] - file_info["dt_start"]

    # Extract devices
    devices = {}
    for dev_name, dev in vf.devs.items():
        devices[dev_name] = {
            "name": str(dev_name),
            "type_name": str(dev.dtname) if hasattr(dev, 'dtname') else "",
            "port": "",  # Port info not available in Python VitalDB
        }

    # Extract tracks (metadata only - no records for smaller golden files)
    tracks = {}
    for trk_name, trk in vf.trks.items():
        track_info = {
            "name": str(trk_name),
            "type": int(trk.type) if hasattr(trk, 'type') else 0,
            "fmt": int(trk.fmt) if hasattr(trk, 'fmt') else 0,
            "unit": str(trk.unit) if hasattr(trk, 'unit') else "",
            "sample_rate": float(trk.srate) if hasattr(trk, 'srate') else 0.0,
            "gain": float(trk.gain) if hasattr(trk, 'gain') else 1.0,
            "offset": float(trk.offset) if hasattr(trk, 'offset') else 0.0,
            "min_display": float(trk.mindisp) if hasattr(trk, 'mindisp') else 0.0,
            "max_display": float(trk.maxdisp) if hasattr(trk, 'maxdisp') else 0.0,
            "color": int(trk.col) if hasattr(trk, 'col') else 0,
            "monitor_type": int(trk.montype) if hasattr(trk, 'montype') else 0,
        }

        # Add device name if available
        if hasattr(trk, 'dname'):
            track_info["device_name"] = str(trk.dname)

        # Extract sample records (limit to first 100 for smaller files)
        records = []
        if hasattr(trk, 'recs') and trk.recs:
            for i, rec in enumerate(trk.recs[:100]):  # Limit to 100 records
                records.append({
                    "dt": float(rec['dt']),
                    "val": serialize_value(rec['val']),
                })

        track_info["records_count"] = len(trk.recs) if hasattr(trk, 'recs') else 0
        track_info["records"] = records

        tracks[trk_name] = track_info

    return {
        "file_info": file_info,
        "devices": devices,
        "tracks": tracks,
    }


def main():
    # Determine project root and paths
    script_dir = Path(__file__).parent
    project_root = script_dir.parent
    data_sample_dir = project_root / "data_sample"
    golden_dir = project_root / "vital" / "testdata" / "golden"

    # Create golden directory if it doesn't exist
    golden_dir.mkdir(parents=True, exist_ok=True)

    # Find all .vital files in data_sample directory
    vital_files = sorted(data_sample_dir.glob("*.vital"))

    if not vital_files:
        print(f"No .vital files found in {data_sample_dir}", file=sys.stderr)
        sys.exit(1)

    print(f"Found {len(vital_files)} VitalDB files to process")
    print(f"Output directory: {golden_dir}")
    print()

    # Process each file
    success_count = 0
    for vital_file in vital_files:
        golden_data = extract_vital_data(vital_file)

        if golden_data is None:
            continue

        # Save as JSON
        output_file = golden_dir / f"{vital_file.stem}.json"
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(golden_data, f, indent=2, ensure_ascii=False)

        print(f"  ✓ Generated: {output_file.name}")
        print(f"    - Devices: {len(golden_data['devices'])}")
        print(f"    - Tracks: {len(golden_data['tracks'])}")
        print()

        success_count += 1

    print(f"Successfully generated {success_count}/{len(vital_files)} golden files")

    if success_count < len(vital_files):
        sys.exit(1)


if __name__ == "__main__":
    main()
