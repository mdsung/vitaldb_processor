#!/usr/bin/env python3
"""
Performance comparison between Python VitalDB and Go VitalDB Processor
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
    """Benchmark Python VitalDB library"""
    start_time = time.time()

    total_tracks = 0
    total_records = 0

    for file_path in file_paths:
        # Load the file
        vf = vitaldb.VitalFile(file_path)

        # Count tracks and records
        track_names = vf.get_track_names()
        total_tracks += len(track_names)

        for track_name in track_names:
            # Get track data (this actually loads the data)
            try:
                vals = vf.get_samples(track_name, interval=1.0)
                if vals is not None:
                    total_records += len(vals)
            except Exception as e:
                # Some tracks might fail to load
                pass

    elapsed = time.time() - start_time

    return {
        'elapsed': elapsed,
        'files': len(file_paths),
        'tracks': total_tracks,
        'records': total_records
    }


def main():
    # Find all .vital files
    data_dir = "../data_sample"
    if not os.path.exists(data_dir):
        print(f"ERROR: {data_dir} directory not found")
        sys.exit(1)

    file_paths = glob.glob(f"{data_dir}/*.vital")
    if not file_paths:
        print(f"ERROR: No .vital files found in {data_dir}")
        sys.exit(1)

    print("=" * 60)
    print("Python VitalDB Performance Benchmark")
    print("=" * 60)
    print(f"Files to process: {len(file_paths)}")

    # Calculate total size
    total_size = sum(os.path.getsize(f) for f in file_paths)
    print(f"Total size: {total_size / 1024 / 1024:.2f} MB")
    print()

    # Run benchmark
    print("Running benchmark...")
    result = benchmark_python_vitaldb(file_paths)

    print()
    print("Results:")
    print("-" * 60)
    print(f"Files processed:    {result['files']}")
    print(f"Total tracks:       {result['tracks']}")
    print(f"Total records:      {result['records']:,}")
    print(f"Time elapsed:       {result['elapsed']:.3f} seconds")
    print(f"Throughput:         {total_size / 1024 / 1024 / result['elapsed']:.2f} MB/s")
    print(f"Files per second:   {result['files'] / result['elapsed']:.2f}")
    print()


if __name__ == "__main__":
    main()
