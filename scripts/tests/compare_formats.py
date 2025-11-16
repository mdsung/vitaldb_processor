#!/usr/bin/env python3
"""
Go VitalDB Processor 포맷 비교 및 검증 스크립트
- Python VitalDB 기준 데이터 생성
- Go의 JSON, JSON(compact), MessagePack 출력 검증
- 성능 및 정확도 비교
"""

import subprocess
import time
import json
import sys
import os

# MessagePack import (옵션)
try:
    import msgpack
    MSGPACK_AVAILABLE = True
except ImportError:
    MSGPACK_AVAILABLE = False
    print("⚠️  msgpack이 설치되지 않았습니다. MessagePack 테스트를 건너뜁니다.")
    print("   설치: pip install msgpack")

import vitaldb

# VitalDB 버퍼 오류 수정 (필수)
vitaldb.utils.FMT_TYPE_LEN[7] = ("i", 4)
vitaldb.utils.FMT_TYPE_LEN[8] = ("I", 4)

def load_python_vitaldb(filepath):
    """Python VitalDB로 파일 로드 (기준 데이터)"""
    print(f"📖 Python VitalDB로 파일 로드: {filepath}")
    start = time.perf_counter()
    vf = vitaldb.VitalFile(filepath)
    elapsed = time.perf_counter() - start

    # 기본 정보 추출
    result = {
        'file_info': {
            'dt_start': vf.dtstart,
            'dt_end': vf.dtend,
            'duration': vf.dtend - vf.dtstart if vf.dtend and vf.dtstart else 0,
            'tracks_count': len(vf.trks),
        },
        'tracks': {},
        'load_time': elapsed
    }

    # 트랙 정보
    for trk_name, trk in vf.trks.items():
        records = []
        if hasattr(trk, 'times') and hasattr(trk, 'vals'):
            if trk.times is not None and trk.vals is not None:
                for t, v in zip(trk.times, trk.vals):
                    records.append({'dt': t, 'val': v})

        result['tracks'][trk_name] = {
            'name': trk_name,
            'type': trk.type,
            'unit': trk.unit,
            'sample_rate': trk.srate,
            'records_count': len(records),
            'records': records[:3]  # 샘플 데이터만
        }

    return result

def load_go_json(filepath, compact=False):
    """Go VitalDB Processor (JSON 포맷)"""
    format_label = "JSON (compact)" if compact else "JSON (pretty)"
    print(f"🔧 Go VitalDB Processor ({format_label}): {filepath}")

    cmd = [
        './vitaldb_processor',
        '-format', 'json',
        '-max-tracks', '0',
        '-max-samples', '0',
        '-quiet',
    ]

    if compact:
        cmd.append('-compact')

    cmd.append(filepath)

    start = time.perf_counter()
    result = subprocess.run(cmd, capture_output=True, text=True, check=True)
    elapsed = time.perf_counter() - start

    data = json.loads(result.stdout)
    data['load_time'] = elapsed
    data['output_size'] = len(result.stdout)

    return data

def load_go_msgpack(filepath):
    """Go VitalDB Processor (MessagePack 포맷)"""
    if not MSGPACK_AVAILABLE:
        return None

    print(f"📦 Go VitalDB Processor (MessagePack): {filepath}")

    cmd = [
        './vitaldb_processor',
        '-format', 'msgpack',
        '-max-tracks', '0',
        '-max-samples', '0',
        '-quiet',
        filepath
    ]

    start = time.perf_counter()
    result = subprocess.run(cmd, capture_output=True, check=True)
    elapsed = time.perf_counter() - start

    data = msgpack.unpackb(result.stdout)
    data['load_time'] = elapsed
    data['output_size'] = len(result.stdout)

    return data

def compare_file_info(python_data, go_data, label):
    """파일 정보 비교"""
    print(f"\n  {'='*60}")
    print(f"  {label} - 파일 정보 비교")
    print(f"  {'='*60}")

    py_info = python_data['file_info']
    go_info = go_data.get('file_info', {})

    fields = ['dt_start', 'dt_end', 'duration', 'tracks_count']
    all_match = True

    for field in fields:
        py_val = py_info.get(field)
        go_val = go_info.get(field)

        match = py_val == go_val
        all_match = all_match and match

        status = "✅" if match else "❌"
        print(f"  {status} {field:15} Python: {py_val:>10} | Go: {go_val:>10}")

    return all_match

def compare_tracks(python_data, go_data, label):
    """트랙 정보 비교 (샘플 데이터)"""
    print(f"\n  {'='*60}")
    print(f"  {label} - 트랙 비교 (샘플)")
    print(f"  {'='*60}")

    py_tracks = python_data['tracks']
    go_tracks = go_data.get('tracks', {})

    # 트랙 개수 확인
    py_count = len(py_tracks)
    go_count = len(go_tracks)
    count_match = py_count == go_count

    status = "✅" if count_match else "❌"
    print(f"  {status} 트랙 개수: Python {py_count} | Go {go_count}")

    # 샘플 트랙 3개만 비교
    sample_tracks = list(py_tracks.keys())[:3]
    track_match = True

    for trk_name in sample_tracks:
        py_trk = py_tracks.get(trk_name, {})
        go_trk = go_tracks.get(trk_name, {})

        if not go_trk:
            print(f"  ❌ 트랙 '{trk_name}' Go에서 누락")
            track_match = False
            continue

        # 레코드 개수 비교
        py_rec_count = py_trk.get('records_count', 0)
        go_rec_count = go_trk.get('records_count', 0)
        rec_match = py_rec_count == go_rec_count

        status = "✅" if rec_match else "❌"
        print(f"  {status} {trk_name[:20]:20} records: Python {py_rec_count:>6} | Go {go_rec_count:>6}")

        track_match = track_match and rec_match

    return count_match and track_match

def compare_performance(results):
    """성능 비교"""
    print(f"\n{'='*70}")
    print(f"성능 비교")
    print(f"{'='*70}")

    baseline = results['Python VitalDB']['load_time']

    print(f"{'포맷':<25} {'시간(ms)':>12} {'크기(MB)':>12} {'Python 대비':>15}")
    print(f"{'-'*70}")

    for label, data in results.items():
        load_time = data['load_time'] * 1000  # ms로 변환
        size_mb = data.get('output_size', 0) / 1024 / 1024
        speedup = baseline / data['load_time'] if data['load_time'] > 0 else 0

        size_str = f"{size_mb:.2f}" if size_mb > 0 else "N/A"

        print(f"{label:<25} {load_time:>12.1f} {size_str:>12} {speedup:>14.2f}x")

def main():
    if len(sys.argv) < 2:
        print("Usage: python compare_formats.py <vital_file_path>")
        print("\nExample:")
        print("  python compare_formats.py ./data_sample/MICUB08_240520_230000.vital")
        sys.exit(1)

    filepath = sys.argv[1]

    if not os.path.exists(filepath):
        print(f"❌ 파일을 찾을 수 없습니다: {filepath}")
        sys.exit(1)

    print("="*70)
    print("VitalDB Processor 포맷 비교 및 검증")
    print("="*70)
    print(f"파일: {filepath}\n")

    # 1. Python VitalDB (기준)
    try:
        python_result = load_python_vitaldb(filepath)
        print(f"   로딩 시간: {python_result['load_time']*1000:.1f}ms")
        print(f"   트랙 개수: {len(python_result['tracks'])}")
    except Exception as e:
        print(f"❌ Python VitalDB 로드 실패: {e}")
        sys.exit(1)

    results = {'Python VitalDB': python_result}

    # 2. Go JSON (pretty)
    print()
    try:
        go_json_pretty = load_go_json(filepath, compact=False)
        print(f"   로딩 시간: {go_json_pretty['load_time']*1000:.1f}ms")
        print(f"   출력 크기: {go_json_pretty['output_size']/1024/1024:.2f}MB")
        results['Go JSON (pretty)'] = go_json_pretty
    except Exception as e:
        print(f"❌ Go JSON (pretty) 실패: {e}")

    # 3. Go JSON (compact)
    print()
    try:
        go_json_compact = load_go_json(filepath, compact=True)
        print(f"   로딩 시간: {go_json_compact['load_time']*1000:.1f}ms")
        print(f"   출력 크기: {go_json_compact['output_size']/1024/1024:.2f}MB")
        results['Go JSON (compact)'] = go_json_compact
    except Exception as e:
        print(f"❌ Go JSON (compact) 실패: {e}")

    # 4. Go MessagePack
    print()
    if MSGPACK_AVAILABLE:
        try:
            go_msgpack = load_go_msgpack(filepath)
            if go_msgpack:
                print(f"   로딩 시간: {go_msgpack['load_time']*1000:.1f}ms")
                print(f"   출력 크기: {go_msgpack['output_size']/1024/1024:.2f}MB")
                results['Go MessagePack'] = go_msgpack
        except Exception as e:
            print(f"❌ Go MessagePack 실패: {e}")
    else:
        print("   ⏭️  MessagePack 테스트 건너뜀 (msgpack 미설치)")

    # 정확도 검증
    print(f"\n{'='*70}")
    print("정확도 검증")
    print(f"{'='*70}")

    for label, data in results.items():
        if label == 'Python VitalDB':
            continue

        file_match = compare_file_info(python_result, data, label)
        track_match = compare_tracks(python_result, data, label)

        overall = "✅ 정확도 100%" if (file_match and track_match) else "❌ 불일치 발견"
        print(f"\n  {overall}")

    # 성능 비교
    compare_performance(results)

    print(f"\n{'='*70}")
    print("테스트 완료")
    print(f"{'='*70}")

if __name__ == '__main__':
    main()
