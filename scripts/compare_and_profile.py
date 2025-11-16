#!/usr/bin/env python3
"""
VitalDB Processor: Python vs Go 정확도 검증 및 성능 프로파일링

Python VitalDB (Golden Standard)와 Go 구현을 비교하여:
1. 정확도 검증 (출력 데이터 동일성)
2. 성능 프로파일링 (처리 시간, 메모리 사용량)
"""

import vitaldb
import subprocess
import json
import time
import os
import sys
import tracemalloc
import traceback
from pathlib import Path
from typing import Dict, List, Tuple, Any
from dataclasses import dataclass, asdict

# 필수: Python VitalDB 버퍼 오류 수정
vitaldb.utils.FMT_TYPE_LEN[7] = ("i", 4)
vitaldb.utils.FMT_TYPE_LEN[8] = ("I", 4)


@dataclass
class ProfileResult:
    """프로파일링 결과"""
    file_name: str
    file_size_mb: float

    # Python VitalDB
    python_time: float
    python_memory_mb: float
    python_tracks_count: int
    python_total_records: int

    # Go Processor
    go_time: float
    go_memory_mb: float
    go_tracks_count: int
    go_total_records: int

    # 비교
    accuracy_match: bool
    mismatched_tracks: List[str]
    speedup: float  # Go가 Python보다 몇 배 빠른지

    # 상세 차이점
    differences: Dict[str, Any]


class VitalDBComparator:
    """Python VitalDB vs Go VitalDB Processor 비교 도구"""

    def __init__(self, go_binary_path: str = "./example/vitaldb_processor"):
        self.go_binary = go_binary_path
        self._check_go_binary()

    def _check_go_binary(self):
        """Go 바이너리 존재 확인"""
        if not os.path.exists(self.go_binary):
            raise FileNotFoundError(
                f"Go binary not found at {self.go_binary}\n"
                f"Please build it first: cd example && go build -o vitaldb_processor main.go"
            )

    def load_with_python(self, vital_path: str) -> Tuple[Dict, float, float]:
        """
        Python VitalDB로 파일 로드 (시간 및 메모리 측정)

        Returns:
            (data_dict, elapsed_time, memory_mb)
        """
        tracemalloc.start()
        start_time = time.perf_counter()

        try:
            vf = vitaldb.VitalFile(vital_path)

            # 데이터를 딕셔너리로 변환
            data = {
                'file_info': {
                    'dt_start': vf.dtstart,
                    'dt_end': vf.dtend,
                    'duration': vf.dtend - vf.dtstart if vf.dtend and vf.dtstart else 0,
                    'dgmt': vf.dgmt if hasattr(vf, 'dgmt') else 0,
                },
                'devices': {},
                'tracks': {}
            }

            # 디바이스 정보
            if hasattr(vf, 'devs') and vf.devs:
                for dev_name, dev in vf.devs.items():
                    data['devices'][dev_name] = {
                        'type_name': dev.type if hasattr(dev, 'type') else '',
                        'port': dev.port if hasattr(dev, 'port') else ''
                    }

            # 트랙 정보
            total_records = 0
            for trk_name, trk in vf.trks.items():
                # VitalDB Track 객체는 recs 속성으로 접근
                records_count = len(trk.recs) if hasattr(trk, 'recs') and trk.recs else 0
                total_records += records_count

                data['tracks'][trk_name] = {
                    'type': trk.type if hasattr(trk, 'type') else 0,
                    'fmt': trk.fmt if hasattr(trk, 'fmt') else 0,
                    'unit': trk.unit if hasattr(trk, 'unit') else '',
                    'sample_rate': trk.srate if hasattr(trk, 'srate') else 0,
                    'records_count': records_count,
                    # 실제 데이터는 메모리 절약을 위해 샘플만 저장
                    'first_record': self._get_first_record(trk) if records_count > 0 else None,
                    'last_record': self._get_last_record(trk) if records_count > 0 else None,
                }

            data['file_info']['tracks_count'] = len(vf.trks)
            data['file_info']['devices_count'] = len(vf.devs) if hasattr(vf, 'devs') else 0
            data['file_info']['total_records'] = total_records

        finally:
            elapsed_time = time.perf_counter() - start_time
            current, peak = tracemalloc.get_traced_memory()
            tracemalloc.stop()
            memory_mb = peak / 1024 / 1024

        return data, elapsed_time, memory_mb

    def _get_first_record(self, track) -> Dict:
        """트랙의 첫 번째 레코드 추출"""
        try:
            if hasattr(track, 'recs') and track.recs and len(track.recs) > 0:
                rec = track.recs[0]
                val = rec.get('val')
                # numpy array나 list인 경우 첫 몇 개만
                if hasattr(val, '__len__') and not isinstance(val, str):
                    val = list(val[:5]) if len(val) > 5 else list(val)
                    # numpy int16 등을 Python int로 변환
                    val = [int(x) if hasattr(x, 'item') else x for x in val]
                return {'time': rec.get('dt'), 'value': val}
        except Exception as e:
            pass
        return None

    def _get_last_record(self, track) -> Dict:
        """트랙의 마지막 레코드 추출"""
        try:
            if hasattr(track, 'recs') and track.recs and len(track.recs) > 0:
                rec = track.recs[-1]
                val = rec.get('val')
                if hasattr(val, '__len__') and not isinstance(val, str):
                    val = list(val[:5]) if len(val) > 5 else list(val)
                    val = [int(x) if hasattr(x, 'item') else x for x in val]
                return {'time': rec.get('dt'), 'value': val}
        except Exception as e:
            pass
        return None

    def load_with_go(self, vital_path: str) -> Tuple[Dict, float]:
        """
        Go VitalDB Processor로 파일 로드 (시간 측정)

        Note: Go의 메모리 사용량은 별도 프로파일링 필요

        Returns:
            (data_dict, elapsed_time)
        """
        start_time = time.perf_counter()

        cmd = [
            self.go_binary,
            '-format', 'json',
            '-max-tracks', '0',      # 모든 트랙
            '-max-samples', '0',    # 모든 샘플 (무제한)
            '-quiet',
            vital_path
        ]

        result = subprocess.run(cmd, capture_output=True, text=True)
        elapsed_time = time.perf_counter() - start_time

        if result.returncode != 0:
            raise RuntimeError(f"Go binary failed: {result.stderr}")

        data = json.loads(result.stdout)

        # total_records 계산
        total_records = sum(
            track.get('records_count', len(track.get('records', [])))
            for track in data.get('tracks', {}).values()
        )
        data['file_info']['total_records'] = total_records

        return data, elapsed_time

    def compare_outputs(self, python_data: Dict, go_data: Dict) -> Tuple[bool, List[str], Dict]:
        """
        Python과 Go 출력 비교

        Returns:
            (is_match, mismatched_tracks, detailed_differences)
        """
        differences = {}
        mismatched_tracks = []

        # 1. 파일 정보 비교
        py_info = python_data['file_info']
        go_info = go_data['file_info']

        for key in ['dt_start', 'dt_end', 'tracks_count', 'devices_count']:
            if abs(py_info.get(key, 0) - go_info.get(key, 0)) > 1e-6:
                differences[f'file_info.{key}'] = {
                    'python': py_info.get(key),
                    'go': go_info.get(key)
                }

        # 2. 트랙 개수 비교
        py_tracks = set(python_data['tracks'].keys())
        go_tracks = set(go_data['tracks'].keys())

        only_python = py_tracks - go_tracks
        only_go = go_tracks - py_tracks

        if only_python:
            differences['tracks_only_in_python'] = list(only_python)
            mismatched_tracks.extend(only_python)

        if only_go:
            differences['tracks_only_in_go'] = list(only_go)
            mismatched_tracks.extend(only_go)

        # 3. 공통 트랙 비교
        common_tracks = py_tracks & go_tracks
        for track_name in common_tracks:
            py_track = python_data['tracks'][track_name]
            go_track = go_data['tracks'][track_name]

            track_diff = {}

            # 메타데이터 비교
            for key in ['type', 'fmt', 'unit', 'sample_rate', 'records_count']:
                py_val = py_track.get(key)
                go_val = go_track.get(key)

                if isinstance(py_val, float) and isinstance(go_val, float):
                    if abs(py_val - go_val) > 1e-6:
                        track_diff[key] = {'python': py_val, 'go': go_val}
                elif py_val != go_val:
                    track_diff[key] = {'python': py_val, 'go': go_val}

            if track_diff:
                differences[f'track.{track_name}'] = track_diff
                mismatched_tracks.append(track_name)

        is_match = len(differences) == 0
        return is_match, mismatched_tracks, differences

    def profile_file(self, vital_path: str) -> ProfileResult:
        """
        단일 파일에 대한 전체 프로파일링
        """
        file_name = os.path.basename(vital_path)
        file_size_mb = os.path.getsize(vital_path) / 1024 / 1024

        print(f"\n{'='*60}")
        print(f"Profiling: {file_name} ({file_size_mb:.2f} MB)")
        print(f"{'='*60}")

        # Python VitalDB
        print("Loading with Python VitalDB...")
        py_data, py_time, py_memory = self.load_with_python(vital_path)
        py_tracks = len(py_data['tracks'])
        py_records = py_data['file_info']['total_records']
        print(f"  ✓ Time: {py_time:.4f}s, Memory: {py_memory:.2f} MB")
        print(f"  ✓ Tracks: {py_tracks}, Records: {py_records}")

        # Go VitalDB Processor
        print("Loading with Go VitalDB Processor...")
        go_data, go_time = self.load_with_go(vital_path)
        go_tracks = len(go_data['tracks'])
        go_records = go_data['file_info']['total_records']
        print(f"  ✓ Time: {go_time:.4f}s")
        print(f"  ✓ Tracks: {go_tracks}, Records: {go_records}")

        # 정확도 비교
        print("Comparing outputs...")
        is_match, mismatched, differences = self.compare_outputs(py_data, go_data)

        if is_match:
            print("  ✓ MATCH: Python and Go outputs are identical!")
        else:
            print(f"  ✗ MISMATCH: Found {len(mismatched)} differences")
            if mismatched:
                print(f"  Mismatched tracks: {mismatched[:5]}...")

        # 성능 비교
        speedup = py_time / go_time if go_time > 0 else 0
        if speedup > 1:
            print(f"  🚀 Go is {speedup:.2f}x faster than Python")
        elif speedup < 1:
            print(f"  🐌 Go is {1/speedup:.2f}x slower than Python")

        return ProfileResult(
            file_name=file_name,
            file_size_mb=file_size_mb,
            python_time=py_time,
            python_memory_mb=py_memory,
            python_tracks_count=py_tracks,
            python_total_records=py_records,
            go_time=go_time,
            go_memory_mb=0,  # Go 메모리는 별도 측정 필요
            go_tracks_count=go_tracks,
            go_total_records=go_records,
            accuracy_match=is_match,
            mismatched_tracks=mismatched,
            speedup=speedup,
            differences=differences
        )

    def profile_directory(self, data_dir: str) -> List[ProfileResult]:
        """
        디렉토리 내 모든 .vital 파일 프로파일링
        """
        vital_files = sorted(Path(data_dir).glob("*.vital"))

        if not vital_files:
            raise FileNotFoundError(f"No .vital files found in {data_dir}")

        print(f"\nFound {len(vital_files)} .vital files in {data_dir}")

        results = []
        for vital_path in vital_files:
            try:
                result = self.profile_file(str(vital_path))
                results.append(result)
            except Exception as e:
                print(f"  ✗ ERROR: {e}")
                traceback.print_exc()

        return results

    def generate_report(self, results: List[ProfileResult], output_path: str = None):
        """
        프로파일링 결과 리포트 생성
        """
        if not results:
            print("No results to report")
            return

        print("\n" + "="*80)
        print("PROFILING SUMMARY")
        print("="*80)

        # 전체 통계
        total_files = len(results)
        total_matches = sum(1 for r in results if r.accuracy_match)
        avg_speedup = sum(r.speedup for r in results) / total_files
        total_python_time = sum(r.python_time for r in results)
        total_go_time = sum(r.go_time for r in results)

        print(f"\nTotal Files: {total_files}")
        print(f"Accuracy Matches: {total_matches}/{total_files} ({total_matches/total_files*100:.1f}%)")
        print(f"Average Speedup: {avg_speedup:.2f}x")
        print(f"Total Python Time: {total_python_time:.4f}s")
        print(f"Total Go Time: {total_go_time:.4f}s")
        print(f"Overall Speedup: {total_python_time/total_go_time:.2f}x")

        # 개별 파일 결과
        print("\n" + "-"*80)
        print("INDIVIDUAL RESULTS")
        print("-"*80)
        print(f"{'File':<30} {'Size(MB)':<10} {'Py(s)':<10} {'Go(s)':<10} {'Speedup':<10} {'Match':<10}")
        print("-"*80)

        for r in results:
            match_str = "✓" if r.accuracy_match else "✗"
            print(f"{r.file_name:<30} {r.file_size_mb:<10.2f} {r.python_time:<10.4f} "
                  f"{r.go_time:<10.4f} {r.speedup:<10.2f} {match_str:<10}")

        # 불일치 상세
        mismatches = [r for r in results if not r.accuracy_match]
        if mismatches:
            print("\n" + "-"*80)
            print("MISMATCHES DETAIL")
            print("-"*80)
            for r in mismatches:
                print(f"\n{r.file_name}:")
                print(f"  Mismatched tracks: {len(r.mismatched_tracks)}")
                if r.differences:
                    print(f"  Differences:")
                    for key, val in list(r.differences.items())[:5]:
                        print(f"    - {key}: {val}")

        # JSON 리포트 저장
        if output_path:
            report_data = {
                'summary': {
                    'total_files': total_files,
                    'accuracy_matches': total_matches,
                    'average_speedup': avg_speedup,
                    'total_python_time': total_python_time,
                    'total_go_time': total_go_time,
                    'overall_speedup': total_python_time / total_go_time
                },
                'results': [asdict(r) for r in results]
            }

            with open(output_path, 'w') as f:
                json.dump(report_data, f, indent=2)

            print(f"\n✓ Report saved to: {output_path}")


def main():
    """메인 실행"""
    import argparse

    parser = argparse.ArgumentParser(description='Compare and profile Python VitalDB vs Go implementation')
    parser.add_argument('--data-dir', default='./data_sample', help='Directory containing .vital files')
    parser.add_argument('--go-binary', default='./example/vitaldb_processor', help='Path to Go binary')
    parser.add_argument('--output', default='./benchmark/profile_results.json', help='Output report path')
    parser.add_argument('--file', help='Profile single file instead of directory')

    args = parser.parse_args()

    try:
        comparator = VitalDBComparator(go_binary_path=args.go_binary)

        if args.file:
            # 단일 파일 프로파일링
            result = comparator.profile_file(args.file)
            results = [result]
        else:
            # 디렉토리 프로파일링
            results = comparator.profile_directory(args.data_dir)

        # 리포트 생성
        comparator.generate_report(results, output_path=args.output)

        # 종료 코드
        all_match = all(r.accuracy_match for r in results)
        sys.exit(0 if all_match else 1)

    except Exception as e:
        print(f"\n✗ ERROR: {e}", file=sys.stderr)
        traceback.print_exc()
        sys.exit(2)


if __name__ == '__main__':
    main()
