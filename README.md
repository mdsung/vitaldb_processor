# VitalDB Processor

VitalDB 파일(.vital)을 읽고 처리하기 위한 Go 라이브러리입니다.

## 설치

```bash
go get github.com/mdsung/vitaldb_processor
```

## 사용법

### 기본 사용 예시

```go
package main

import (
    "fmt"
    "log"

    "github.com/mdsung/vitaldb_processor/vital"
)

func main() {
    // VitalDB 파일 읽기
    vf, err := vital.NewVitalFile("example.vital")
    if err != nil {
        log.Fatal(err)
    }

    // 기본 정보 출력
    fmt.Printf("Start Time: %f\n", vf.DtStart)
    fmt.Printf("End Time: %f\n", vf.DtEnd)
    fmt.Printf("Number of Tracks: %d\n", len(vf.Trks))
    fmt.Printf("Number of Devices: %d\n", len(vf.Devs))

    // 트랙 정보 출력
    for name, track := range vf.Trks {
        fmt.Printf("Track: %s, Unit: %s, Records: %d\n",
            name, track.Unit, len(track.Recs))
    }

    // 디바이스 정보 출력
    for name, device := range vf.Devs {
        fmt.Printf("Device: %s, Type: %s, Port: %s\n",
            name, device.TypeName, device.Port)
    }
}
```

## API 문서

### 주요 타입

#### VitalFile

VitalDB 파일의 전체 구조를 나타냅니다.

```go
type VitalFile struct {
    Devs    map[string]Device  // 의료 장비 정보
    Trks    map[string]Track   // 데이터 트랙 정보
    DtStart float64           // 시작 시간
    DtEnd   float64           // 종료 시간
    Dgmt    int16             // GMT 오프셋
    Order   []string          // 트랙 순서
    DevIDs  map[uint32]string // 디바이스 ID 매핑
    TrkIDs  map[uint16]string // 트랙 ID 매핑
}
```

#### Device

의료 장비 정보를 나타냅니다.

```go
type Device struct {
    Name     string // 장비 이름
    TypeName string // 장비 타입
    Port     string // 포트 정보
}
```

#### Track

데이터 트랙 정보를 나타냅니다.

```go
type Track struct {
    Name    string      // 트랙 이름
    Type    uint8       // 데이터 타입
    Fmt     uint8       // 포맷
    Unit    string      // 단위
    SRate   float32     // 샘플링 레이트
    Gain    float64     // 게인
    Offset  float64     // 오프셋
    Mindisp float32     // 최소 표시값
    Maxdisp float32     // 최대 표시값
    Col     uint32      // 색상
    Montype uint8       // 모니터 타입
    DName   string      // 디바이스 이름
    Recs    []Rec       // 데이터 레코드들
}
```

#### Rec

개별 데이터 레코드를 나타냅니다.

```go
type Rec struct {
    Dt  float64      // 시간
    Val interface{}  // 값 (데이터 타입에 따라 다름)
}
```

### 주요 함수

#### NewVitalFile

```go
func NewVitalFile(path string) (*VitalFile, error)
```

VitalDB 파일을 읽어서 VitalFile 구조체로 반환합니다.

**매개변수:**

- `path`: VitalDB 파일 경로

**반환값:**

- `*VitalFile`: 파싱된 VitalDB 파일 구조체
- `error`: 오류 정보

## 특징

- **고성능**: Go의 네이티브 성능으로 빠른 파일 처리
- **메모리 효율적**: 필요한 데이터만 메모리에 로드
- **타입 안전**: 강타입 언어의 장점을 활용한 안전한 데이터 처리
- **표준 라이브러리**: 외부 의존성 최소화

## Python에서 활용하기

개선된 Go 바이너리와 함께 Python에서 더 효과적으로 사용할 수 있습니다.

### 1. JSON 출력을 통한 완전한 데이터 분석

```python
import subprocess
import json
import numpy as np
import matplotlib.pyplot as plt

def load_vital_data(file_path, **kwargs):
    """VitalDB 파일을 JSON으로 로드"""
    cmd = ['./vitaldb_processor', '-format', 'json']

    # 옵션 추가
    if 'tracks' in kwargs:
        cmd.extend(['-tracks', ','.join(kwargs['tracks'])])
    if 'track_type' in kwargs:
        cmd.extend(['-track-type', kwargs['track_type']])
    if 'start_time' in kwargs:
        cmd.extend(['-start-time', str(kwargs['start_time'])])
    if 'end_time' in kwargs:
        cmd.extend(['-end-time', str(kwargs['end_time'])])
    if 'max_tracks' in kwargs:
        cmd.extend(['-max-tracks', str(kwargs['max_tracks'])])

    cmd.append(file_path)

    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        raise Exception(f"Error processing file: {result.stderr}")

    return json.loads(result.stdout)

# 사용 예시
# 전체 데이터 로드
data = load_vital_data('data.vital')

# 기본 정보 확인
file_info = data['file_info']
print(f"시작 시간: {file_info['dt_start']}")
print(f"종료 시간: {file_info['dt_end']}")
print(f"지속 시간: {file_info['duration']:.2f}초")
print(f"트랙 개수: {file_info['tracks_count']}")
print(f"디바이스 개수: {file_info['devices_count']}")

# 디바이스 정보
print("\n=== 디바이스 ===")
for name, device in data['devices'].items():
    print(f"- {name}: {device['type_name']} (포트: {device['port']})")

# 모든 트랙 정보
print("\n=== 트랙 ===")
for name, track in data['tracks'].items():
    print(f"- {name}: {track['type_name']} ({track['unit']}, {track['sample_rate']} Hz)")
```

### 2. 특정 트랙/변수 필터링

```python
# ECG와 혈압 관련 트랙만 가져오기
vital_signs = load_vital_data('data.vital', tracks=['ECG_II', 'ART', 'HR'])

# WAVE 타입 트랙들만 가져오기 (모든 트랙, 제한 없음)
wave_data = load_vital_data('data.vital', track_type='WAVE', max_tracks=0)

# 수치형 데이터만 가져오기
numeric_data = load_vital_data('data.vital', track_type='NUMERIC', max_tracks=0)
```

### 3. 시간 범위 기반 데이터 추출

```python
# 처음 5분간의 ECG 데이터
ecg_5min = load_vital_data('data.vital',
                          tracks=['ECG_II'],
                          start_time=0,
                          end_time=300)

# 수술 중 특정 구간 (30분-60분)
surgery_data = load_vital_data('data.vital',
                              start_time=1800,
                              end_time=3600)
```

### 4. 파일 정보 빠른 확인

```python
def get_file_info(file_path):
    """파일 정보만 빠르게 확인"""
    cmd = ['./vitaldb_processor', '-info-only', '-format', 'json', '-quiet', file_path]
    result = subprocess.run(cmd, capture_output=True, text=True)
    return json.loads(result.stdout)

def list_available_tracks(file_path):
    """사용 가능한 트랙 목록 확인"""
    cmd = ['./vitaldb_processor', '-list-tracks', '-format', 'json', '-quiet', file_path]
    result = subprocess.run(cmd, capture_output=True, text=True)
    return json.loads(result.stdout)

# 사용 예시
file_info = get_file_info('data.vital')
tracks_info = list_available_tracks('data.vital')

print(f"파일 지속시간: {file_info['file_info']['duration']:.2f}초")
print("사용 가능한 트랙들:")
for name, track in tracks_info['tracks'].items():
    print(f"  • {name}: {track['type_name']} ({track['unit']}, {track['sample_rate']} Hz)")
```

### 5. 실시간 스트리밍 처리

```python
def stream_vital_data(file_path, window_size=10):
    """윈도우 단위로 데이터를 스트리밍 처리"""
    # 전체 파일 정보 먼저 확인
    file_info = get_file_info(file_path)['file_info']

    total_duration = file_info['duration']
    current_time = file_info['dt_start']

    while current_time < file_info['dt_end']:
        end_time = min(current_time + window_size, file_info['dt_end'])

        # 현재 윈도우 데이터 가져오기
        window_data = load_vital_data(file_path,
                                    start_time=current_time,
                                    end_time=end_time)

        # 데이터 처리 (예: 이상 감지, 알람 등)
        process_window(window_data)

        current_time = end_time
        time.sleep(0.1)  # 실시간 시뮬레이션

def process_window(data):
    """윈도우 데이터 처리 로직"""
    if 'HR' in data['tracks']:
        hr_records = data['tracks']['HR']['records']
        if hr_records:
            avg_hr = sum(r['val'] for r in hr_records) / len(hr_records)
            if avg_hr > 100:
                print(f"⚠️  빈맥 감지: {avg_hr:.1f} bpm")
            elif avg_hr < 60:
                print(f"⚠️  서맥 감지: {avg_hr:.1f} bpm")
```

### 6. 배치 처리 및 분석

```python
import os
import glob
import pandas as pd
from concurrent.futures import ProcessPoolExecutor

def process_vital_file(file_path):
    """단일 VitalDB 파일 처리"""
    try:
        data = load_vital_data(file_path, summary=True)  # summary 모드 사용
        file_info = data['file_info']

        return {
            'file': os.path.basename(file_path),
            'duration': file_info['duration'],
            'tracks_count': file_info['tracks_count'],
            'devices_count': file_info['devices_count'],
            'has_ecg': 'ECG_II' in data['tracks'],
            'has_bp': any('ART' in track for track in data['tracks']),
            'avg_hr': get_average_hr(data)
        }
    except Exception as e:
        return {'file': file_path, 'error': str(e)}

def get_average_hr(data):
    """평균 심박수 계산"""
    if 'HR' in data['tracks'] and data['tracks']['HR']['records']:
        hr_values = [r['val'] for r in data['tracks']['HR']['records']]
        return sum(hr_values) / len(hr_values)
    return None

# 여러 파일 배치 처리
vital_files = glob.glob('data/*.vital')

with ProcessPoolExecutor(max_workers=4) as executor:
    results = list(executor.map(process_vital_file, vital_files))

# 결과를 DataFrame으로 정리
df = pd.DataFrame(results)
print(df.describe())
```

## 새로운 기능 요약

### 해결된 문제점

1. **✅ 트랙 제한 해제**: 이제 모든 트랙을 출력할 수 있습니다 (`-max-tracks 0`)
2. **✅ JSON 출력 지원**: Python 연동에 최적화된 JSON 형식 지원
3. **✅ 디바이스 파싱**: 디바이스 정보가 올바르게 파싱됩니다
4. **✅ 필터링 옵션**: 트랙 타입, 이름, 시간 범위별 필터링 가능
5. **✅ 다양한 출력 모드**: 요약, 목록, 상세 모드 등 지원

### 성능 향상

- **빠른 정보 조회**: `-info-only`, `-quiet` 옵션으로 빠른 파일 확인
- **효율적인 메모리 사용**: 필요한 데이터만 로드
- **병렬 처리 지원**: Python에서 멀티프로세싱으로 배치 처리 가능

## 예제 실행

```bash
# 개선된 바이너리 빌드
cd example
go build -o vitaldb_processor main.go

# 기본 사용법
./vitaldb_processor /path/to/your/file.vital

# JSON 형태로 모든 트랙 출력
./vitaldb_processor -format json -max-tracks 0 /path/to/your/file.vital

# 특정 트랙만 확인
./vitaldb_processor -tracks "ECG_II,HR" /path/to/your/file.vital

# 파일 정보만 빠르게 확인
./vitaldb_processor -info-only -quiet /path/to/your/file.vital

# 새로운 기능들 데모 (VitalDB 파일 없이도 가능)
python3 demo.py
```

## CLI 사용법

이제 다양한 CLI 옵션을 지원합니다:

### 기본 사용법

```bash
./vitaldb_processor [options] <vital_file_path>
```

### 사용 가능한 옵션

```
-format string
    출력 형식 (text, json) (기본값: "text")
-info-only
    파일 정보만 출력
-list-devices
    디바이스 목록만 출력
-list-tracks
    트랙 목록만 출력
-max-samples int
    샘플 데이터 최대 개수 (기본값: 3)
-max-tracks int
    최대 트랙 개수 제한 (0 = 무제한)
-quiet
    조용한 모드 (에러만 출력)
-start-time float
    시작 시간
-end-time float
    종료 시간 (0 = 파일 끝까지)
-summary
    요약 정보만 출력
-track-type string
    트랙 타입 필터 (WAVE, NUMERIC, STRING)
-tracks string
    특정 트랙들만 출력 (쉼표로 구분)
-verbose
    상세 모드
```

### 출력 형식 옵션

```bash
# JSON 형태로 출력 (Python 연동에 최적화)
./vitaldb_processor -format json data.vital

# 기본 텍스트 형태로 출력
./vitaldb_processor data.vital

# 요약 정보만 출력
./vitaldb_processor -summary data.vital
```

### 트랙 필터링 옵션

```bash
# 특정 트랙들만 추출
./vitaldb_processor -tracks "ECG_II,HR,PLETH" data.vital

# 트랙 타입별 필터링
./vitaldb_processor -track-type WAVE data.vital
./vitaldb_processor -track-type NUMERIC data.vital
./vitaldb_processor -track-type STRING data.vital

# 모든 트랙 출력 (제한 없음)
./vitaldb_processor -max-tracks 0 data.vital

# 처음 5개 트랙만 출력
./vitaldb_processor -max-tracks 5 data.vital
```

### 시간 범위 옵션

```bash
# 특정 시간 범위 추출 (초 단위)
./vitaldb_processor -start-time 0 -end-time 300 data.vital

# 처음 5분간의 데이터
./vitaldb_processor -start-time 0 -end-time 300 data.vital
```

### 정보 조회 옵션

```bash
# 트랙 목록만 출력
./vitaldb_processor -list-tracks data.vital

# 파일 정보만 출력
./vitaldb_processor -info-only data.vital

# 디바이스 정보만 출력
./vitaldb_processor -list-devices data.vital
```

### 출력 제어 옵션

```bash
# 샘플 개수 제한
./vitaldb_processor -max-samples 10 data.vital

# 조용한 모드 (에러만 출력)
./vitaldb_processor -quiet data.vital

# 상세 모드 (샘플 데이터까지 표시)
./vitaldb_processor -verbose data.vital
```

### 사용 예시

```bash
# ECG 데이터만 처음 5분간 JSON으로 추출
./vitaldb_processor -tracks "ECG_II" -start-time 0 -end-time 300 -format json data.vital

# 모든 수치형 데이터를 JSON으로 저장
./vitaldb_processor -track-type NUMERIC -format json data.vital > vitals.json

# 파일 정보 빠르게 확인
./vitaldb_processor -info-only -quiet data.vital

# 모든 트랙을 JSON으로 출력 (Python 연동용)
./vitaldb_processor -format json -max-tracks 0 data.vital
```

## 성능

이 라이브러리는 Python VitalDB SDK와 비교하여 약 20% 성능 향상을 제공합니다:

- Python SDK: 3.487초 (20개 파일, 67.2MB)
- Go Library: 2.905초 (20개 파일, 67.2MB) - **1.20x 빠름**

## 라이센스

MIT License

## 기여

이슈나 풀 리퀘스트는 언제나 환영합니다!

## 관련 프로젝트

- [VitalDB](https://vitaldb.net/) - 의료 데이터베이스
- [VitalDB Python SDK](https://github.com/vitaldb/vitaldb-python) - 공식 Python SDK
