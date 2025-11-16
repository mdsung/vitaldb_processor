#!/usr/bin/env python3
"""
성능 비교: JSON vs JSON(compact) vs MessagePack
"""

import subprocess
import time
import sys

def benchmark_format(format_type, compact=False, runs=5):
    """특정 포맷으로 실행 시간 측정"""
    cmd = [
        './vitaldb_processor',
        '-format', format_type,
        '-max-tracks', '0',
        '-max-samples', '0',
        '-quiet',
    ]

    if compact and format_type == 'json':
        cmd.append('-compact')

    cmd.append('./data_sample/MICUB08_240520_230000.vital')

    times = []
    for i in range(runs):
        start = time.perf_counter()
        result = subprocess.run(cmd, capture_output=True)
        elapsed = time.perf_counter() - start
        times.append(elapsed)

        if i == 0:  # 첫 실행에서 출력 크기 확인
            output_size = len(result.stdout)

    avg_time = sum(times) / len(times)
    min_time = min(times)
    max_time = max(times)

    return {
        'format': format_type,
        'compact': compact,
        'avg_time': avg_time,
        'min_time': min_time,
        'max_time': max_time,
        'output_size': output_size,
        'times': times
    }

if __name__ == '__main__':
    print("VitalDB Processor 포맷 성능 비교")
    print("=" * 60)
    print(f"테스트 파일: MICUB08_240520_230000.vital (3.12 MB)")
    print(f"실행 횟수: 5회\n")

    formats = [
        ('json', False, 'JSON (pretty)'),
        ('json', True, 'JSON (compact)'),
        ('msgpack', False, 'MessagePack'),
    ]

    results = []
    for format_type, compact, label in formats:
        print(f"테스트 중: {label}...", end=' ')
        sys.stdout.flush()

        result = benchmark_format(format_type, compact)
        results.append((label, result))

        print(f"평균: {result['avg_time']*1000:.1f}ms")

    print("\n" + "=" * 60)
    print("결과 요약")
    print("=" * 60)
    print(f"{'포맷':<20} {'평균(ms)':<12} {'최소(ms)':<12} {'최대(ms)':<12} {'크기(MB)':<12}")
    print("-" * 60)

    for label, result in results:
        print(f"{label:<20} {result['avg_time']*1000:<12.1f} "
              f"{result['min_time']*1000:<12.1f} {result['max_time']*1000:<12.1f} "
              f"{result['output_size']/1024/1024:<12.2f}")

    # 상대 성능
    print("\n" + "=" * 60)
    print("상대 성능 (JSON pretty 기준)")
    print("=" * 60)

    baseline = results[0][1]['avg_time']
    for label, result in results:
        speedup = baseline / result['avg_time']
        print(f"{label:<20} {speedup:.2f}x")
