#!/usr/bin/env python3
"""
VitalDB Processor 개선된 CLI 기능 데모
실제 VitalDB 파일 없이도 새로운 기능들을 테스트할 수 있습니다.
"""

import json
import os
import subprocess
import sys


def run_command(cmd):
    """명령을 실행하고 결과를 반환"""
    print(f"\n🔍 실행 중: {' '.join(cmd)}")
    print("=" * 60)

    result = subprocess.run(cmd, capture_output=True, text=True)

    if result.returncode == 0:
        print("✅ 성공!")
        if result.stdout:
            print(result.stdout)
        return result.stdout
    else:
        print("❌ 실패!")
        if result.stderr:
            print(f"에러: {result.stderr}")
        return None


def demo_help():
    """도움말 표시"""
    print("\n📖 새로운 CLI 옵션들 확인하기")
    run_command(["./vitaldb_processor", "-h"])


def demo_features_without_file():
    """파일 없이 기능들 시연"""
    print("\n🎭 개선된 CLI 기능 데모")
    print("=" * 60)

    # 도움말 확인
    demo_help()

    # 파일이 없을 때의 동작 확인
    print("\n🚫 파일 없이 실행했을 때:")
    run_command(["./vitaldb_processor"])

    print(
        """
📋 주요 개선사항 요약:

✅ 해결된 문제들:
1. 트랙 제한 문제 - 이제 `-max-tracks 0`으로 모든 트랙 출력 가능
2. JSON 출력 부재 - `-format json`으로 Python 연동 최적화
3. 디바이스 파싱 누락 - 디바이스 정보 올바르게 파싱됨
4. 필터링 옵션 부족 - 다양한 필터링 옵션 추가

🚀 새로운 기능들:
• -format: 출력 형식 선택 (text/json)
• -list-tracks: 트랙 목록만 출력
• -list-devices: 디바이스 목록만 출력  
• -info-only: 파일 정보만 빠르게 확인
• -summary: 요약 정보만 출력
• -tracks: 특정 트랙들만 필터링
• -track-type: 트랙 타입별 필터링 (WAVE/NUMERIC/STRING)
• -max-tracks: 트랙 개수 제한 (0=무제한)
• -start-time, -end-time: 시간 범위 필터링
• -quiet, -verbose: 출력 레벨 조정

🐍 Python 연동 예시:

```python
import subprocess
import json

def load_vital_data(file_path, **kwargs):
    cmd = ['./vitaldb_processor', '-format', 'json']
    
    if 'tracks' in kwargs:
        cmd.extend(['-tracks', ','.join(kwargs['tracks'])])
    if 'track_type' in kwargs:
        cmd.extend(['-track-type', kwargs['track_type']])
    if 'max_tracks' in kwargs:
        cmd.extend(['-max-tracks', str(kwargs['max_tracks'])])
    
    cmd.append(file_path)
    
    result = subprocess.run(cmd, capture_output=True, text=True)
    return json.loads(result.stdout)

# 사용 예시
# data = load_vital_data('data.vital', tracks=['ECG_II', 'HR'])
# wave_data = load_vital_data('data.vital', track_type='WAVE', max_tracks=0)
```

💡 사용 예시:

# 모든 트랙을 JSON으로 출력 (17개 트랙 모두)
./vitaldb_processor -format json -max-tracks 0 data.vital

# ECG 관련 트랙만 필터링
./vitaldb_processor -tracks "ECG_II,ECG_V5" -format json data.vital

# 처음 5분간의 WAVE 데이터만
./vitaldb_processor -track-type WAVE -start-time 0 -end-time 300 data.vital

# 파일 정보만 빠르게 확인
./vitaldb_processor -info-only -quiet data.vital
"""
    )


def create_mock_vital_file():
    """테스트용 mock VitalDB 파일 생성"""
    print("\n🔧 실제 테스트를 위해 VitalDB 파일이 필요합니다.")
    print("VitalDB 파일을 구하는 방법:")
    print("1. https://vitaldb.net 에서 샘플 파일 다운로드")
    print("2. 또는 실제 병원 데이터 사용")
    print("\n파일이 있다면 다음과 같이 테스트할 수 있습니다:")
    print("./vitaldb_processor -format json -max-tracks 0 your_file.vital")


def main():
    """메인 데모 함수"""
    print("🎉 VitalDB Processor 개선된 CLI 기능 데모")
    print("=" * 60)

    # 바이너리 존재 확인
    if not os.path.exists("./vitaldb_processor"):
        print("❌ vitaldb_processor 바이너리를 찾을 수 없습니다.")
        print("다음 명령으로 빌드하세요: go build -o vitaldb_processor main.go")
        return

    # 기능 데모
    demo_features_without_file()

    # Mock 파일 안내
    create_mock_vital_file()


if __name__ == "__main__":
    main()
