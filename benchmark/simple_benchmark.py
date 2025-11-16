#!/usr/bin/env python3
"""
Simple performance comparison - just file opening and basic parsing
"""
import time
import glob
import sys
import os

try:
    import vitaldb
except ImportError:
    print("ERROR: vitaldb package not installed.")
    print("Install with: pip install vitaldb")
    sys.exit(1)


def benchmark_python_vitaldb(file_paths):
    """Benchmark Python VitalDB library - simple version"""
    start_time = time.time()

    total_tracks = 0
    total_devices = 0
    files_processed = 0

    for file_path in file_paths:
        try:
            # Load the file
            vf = vitaldb.VitalFile(file_path)

            # Count tracks
            track_names = vf.get_track_names()
            total_tracks += len(track_names)

            files_processed += 1
        except Exception as e:
            print(f"Error loading {file_path}: {e}")
            continue

    elapsed = time.time() - start_time

    return {
        'elapsed': elapsed,
        'files': files_processed,
        'tracks': total_tracks,
    }


def main():
    # Find all .vital files
    data_dir = "../data_sample"
    if not os.path.exists(data_dir):
        print(f"ERROR: {data_dir} directory not found")
        sys.exit(1)

    file_paths = sorted(glob.glob(f"{data_dir}/*.vital"))
    if not file_paths:
        print(f"ERROR: No .vital files found in {data_dir}")
        sys.exit(1)

    print("=" * 60)
    print("Python VitalDB Performance Benchmark (Simple)")
    print("=" * 60)
    print(f"Files to process: {len(file_paths)}")

    # Calculate total size
    total_size = sum(os.path.getsize(f) for f in file_paths)
    print(f"Total size: {total_size / 1024 / 1024:.2f} MB")
    print()

    # Warm-up run
    print("Warm-up run...")
    _ = benchmark_python_vitaldb(file_paths[:1])

    # Actual benchmark (run 3 times and take average)
    print("Running benchmark (3 iterations)...")
    times = []
    for i in range(3):
        result = benchmark_python_vitaldb(file_paths)
        times.append(result['elapsed'])
        print(f"  Iteration {i+1}: {result['elapsed']:.3f} seconds")

    avg_time = sum(times) / len(times)

    print()
    print("Results:")
    print("-" * 60)
    print(f"Files processed:    {result['files']}")
    print(f"Total tracks:       {result['tracks']}")
    print(f"Average time:       {avg_time:.3f} seconds")
    print(f"Throughput:         {total_size / 1024 / 1024 / avg_time:.2f} MB/s")
    print(f"Files per second:   {result['files'] / avg_time:.2f}")
    print()


if __name__ == "__main__":
    main()
