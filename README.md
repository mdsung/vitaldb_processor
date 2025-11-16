# VitalDB Processor

VitalDB 파일(.vital)을 읽고 처리하기 위한 Go 라이브러리입니다.

## ⚡ 성능

Python VitalDB 대비 **7.29배 빠른 성능** (MessagePack 사용 시)

| 구현 | 시간 (3.12MB 파일) | Python 대비 | 크기 |
|------|-------------------|-------------|------|
| Python VitalDB | 1,123ms | 1.00x (기준) | - |
| **Go (JSON compact)** | **257ms** | **4.37x** ⚡ | 18.0MB |
| **Go (MessagePack)** | **154ms** | **7.29x** 🚀 | 12.6MB |

**주요 특징**:
- ✅ **100% 정확도**: Python VitalDB와 동일한 결과 보장
- ✅ **7.29배 빠른 처리**: MessagePack 사용 시
- ✅ **30% 작은 출력**: 효율적인 바이너리 직렬화
- ✅ **Python 통합 간편**: JSON/MessagePack 양방향 지원

자세한 최적화 내역은 [`notes/optimization_results.md`](notes/optimization_results.md)를 참조하세요.

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

### ⚠️ 중요: Python VitalDB 라이브러리 버퍼 오류 수정

Python VitalDB 라이브러리를 직접 사용할 경우, **반드시** 다음 코드를 추가해야 버퍼 오류가 발생하지 않습니다:

```python
import vitaldb

# 필수: 버퍼 오류 방지를 위한 포맷 타입 설정
vitaldb.utils.FMT_TYPE_LEN[7] = ("i", 4)
vitaldb.utils.FMT_TYPE_LEN[8] = ("I", 4)

# 이제 정상적으로 VitalDB 파일 로드 가능
vf = vitaldb.VitalFile('data.vital')
```

**주의**: 이 설정 없이 Python VitalDB를 사용하면 일부 파일에서 "buffer is too small" 오류가 발생할 수 있습니다. Go VitalDB Processor는 이러한 문제가 없습니다.

### 1. CSV/Parquet를 통한 데이터 로드 (권장)

#### 방법 A: CSV (범용, pandas 호환)

```python
import subprocess
import pandas as pd

def load_vital_csv(file_path, **kwargs):
    """VitalDB 파일을 CSV로 변환 후 pandas DataFrame으로 로드"""
    cmd = ['./vitaldb_processor', '-format', 'csv']

    # 옵션 추가
    if 'tracks' in kwargs:
        cmd.extend(['-tracks', ','.join(kwargs['tracks'])])
    if 'track_type' in kwargs:
        cmd.extend(['-track-type', kwargs['track_type']])
    if 'start_time' in kwargs:
        cmd.extend(['-start-time', str(kwargs['start_time'])])
    if 'end_time' in kwargs:
        cmd.extend(['-end-time', str(kwargs['end_time'])])
    if 'max_samples' in kwargs:
        cmd.extend(['-max-samples', str(kwargs['max_samples'])])

    cmd.append(file_path)

    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        raise Exception(f"Error processing file: {result.stderr}")

    # CSV를 pandas DataFrame으로 직접 로드
    from io import StringIO
    return pd.read_csv(StringIO(result.stdout))

# 사용 예시
df = load_vital_csv('data.vital')
print(df.head())

# 특정 트랙만 로드
ecg_df = load_vital_csv('data.vital', tracks=['ECG_II', 'HR'])

# 시간 범위 지정
df_5min = load_vital_csv('data.vital', start_time=0, end_time=300)

# pandas로 분석
print(df.groupby('track_name')['value'].describe())
```

#### 방법 B: Parquet (고성능, 압축 효율적)

```python
import subprocess
import pandas as pd

def load_vital_parquet(file_path, **kwargs):
    """VitalDB 파일을 Parquet로 변환 후 pandas DataFrame으로 로드"""
    cmd = ['./vitaldb_processor', '-format', 'parquet']

    # 옵션 추가 (CSV와 동일)
    if 'tracks' in kwargs:
        cmd.extend(['-tracks', ','.join(kwargs['tracks'])])
    if 'track_type' in kwargs:
        cmd.extend(['-track-type', kwargs['track_type']])
    if 'start_time' in kwargs:
        cmd.extend(['-start-time', str(kwargs['start_time'])])
    if 'end_time' in kwargs:
        cmd.extend(['-end-time', str(kwargs['end_time'])])

    cmd.append(file_path)

    result = subprocess.run(cmd, capture_output=True)
    if result.returncode != 0:
        raise Exception(f"Error processing file: {result.stderr}")

    # Parquet를 pandas DataFrame으로 직접 로드
    from io import BytesIO
    return pd.read_parquet(BytesIO(result.stdout))

# 사용 예시 (CSV보다 약 30% 빠름)
df = load_vital_parquet('data.vital')
print(df.head())
```

### 2. 데이터 로드 (JSON / MessagePack)

#### 방법 A: JSON (범용, 디버깅 용이)

```python
import subprocess
import json

def load_vital_data_json(file_path, **kwargs):
    """VitalDB 파일을 JSON으로 로드 (4.37배 빠름)"""
    cmd = ['./vitaldb_processor', '-format', 'json', '-compact']

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
```

#### 방법 B: MessagePack (최고 성능, 7.29배 빠름)

```python
import subprocess
import msgpack  # pip install msgpack

def load_vital_data_msgpack(file_path, **kwargs):
    """VitalDB 파일을 MessagePack으로 로드 (7.29배 빠름, 30% 작은 크기)"""
    cmd = ['./vitaldb_processor', '-format', 'msgpack']

    # 옵션 추가 (JSON과 동일)
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

    result = subprocess.run(cmd, capture_output=True)
    if result.returncode != 0:
        raise Exception(f"Error processing file: {result.stderr}")

    return msgpack.unpackb(result.stdout)

# 추천: 성능을 위해 MessagePack 사용, 필요시 JSON fallback
def load_vital_data(file_path, **kwargs):
    """VitalDB 파일 로드 (MessagePack 우선, JSON fallback)"""
    try:
        import msgpack
        return load_vital_data_msgpack(file_path, **kwargs)
    except ImportError:
        return load_vital_data_json(file_path, **kwargs)

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
6. **✅ 파일 호환성 개선**: 파일 끝 불완전 패킷 처리로 더 많은 VitalDB 파일 지원

### 성능 향상

- **빠른 정보 조회**: `-info-only`, `-quiet` 옵션으로 빠른 파일 확인
- **효율적인 메모리 사용**: 필요한 데이터만 로드
- **병렬 처리 지원**: Python에서 멀티프로세싱으로 배치 처리 가능

### 최근 버그 수정 (2025-06-17)

**문제**: 일부 VitalDB 파일에서 `unexpected EOF` 에러가 발생하여 파일을 읽을 수 없었습니다.

**원인**: 파일 끝에서 불완전한 패킷이 있을 때, Go 코드가 엄격하게 에러를 발생시켰으나 Python VitalDB는 이를 무시하고 진행했습니다.

**해결**: Python VitalDB와 동일한 방식으로 파일 끝의 불완전한 패킷을 무시하도록 수정하여 호환성을 개선했습니다.

**결과**: 모든 `data_sample` 파일들이 성공적으로 처리되며, Python VitalDB와 동일한 결과를 얻을 수 있습니다.

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
    출력 형식 (csv, parquet, text, json, msgpack) (기본값: "csv")
-compact
    Compact JSON 출력 (들여쓰기 없음, 성능 향상)
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
# CSV 형태로 출력 (기본값, pandas 호환)
./vitaldb_processor data.vital > output.csv
./vitaldb_processor -format csv data.vital > output.csv

# Parquet 형태로 출력 (압축 효율적, 고성능)
./vitaldb_processor -format parquet data.vital > output.parquet

# MessagePack 형태로 출력 (최고 성능, 7.29배 빠름)
./vitaldb_processor -format msgpack data.vital > output.msgpack

# JSON Compact 형태로 출력 (4.37배 빠름)
./vitaldb_processor -format json -compact data.vital > output.json

# JSON 형태로 출력 (가독성 우선, Pretty-print)
./vitaldb_processor -format json data.vital

# 텍스트 형태로 출력
./vitaldb_processor -format text data.vital

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
# ECG 데이터만 처음 5분간 MessagePack으로 추출 (최고 성능)
./vitaldb_processor -tracks "ECG_II" -start-time 0 -end-time 300 -format msgpack data.vital > ecg.msgpack

# 모든 수치형 데이터를 JSON Compact로 저장
./vitaldb_processor -track-type NUMERIC -format json -compact data.vital > vitals.json

# 파일 정보 빠르게 확인
./vitaldb_processor -info-only -quiet data.vital

# 모든 트랙을 MessagePack으로 출력 (Python 연동용, 최고 성능)
./vitaldb_processor -format msgpack -max-tracks 0 -max-samples 0 data.vital > output.msgpack

# 모든 트랙을 JSON Compact으로 출력 (Python 연동용, 범용)
./vitaldb_processor -format json -compact -max-tracks 0 -max-samples 0 data.vital > output.json
```

## 테스트

프로젝트는 세 가지 유형의 테스트를 지원합니다:

### 테스트 실행 방법

```bash
# 유닛 테스트만 실행 (빠름, 외부 파일 불필요)
make test
# 또는
go test ./vital

# 통합 테스트 실행 (실제 .vital 파일 필요)
make test-integration
# 또는
go test -tags=integration ./vital

# 모든 테스트 실행
make test-all

# 벤치마크 실행
make bench

# 테스트 파일 줄 수 검증
make verify-linecount

# 코드 커버리지 생성 (로컬)
go test ./... -covermode=atomic -coverprofile=coverage.out
go tool cover -html=coverage.out  # 브라우저에서 확인
```

### 테스트 파일 구조

- `vital/unit_test.go` - 유닛 테스트 (외부 파일 의존성 없음)
- `vital/integration_test.go` - 통합 테스트 (`//go:build integration` 태그 필요)
- `vital/benchmark_test.go` - 성능 벤치마크
- `vital/helper_test.go` - 공통 테스트 헬퍼 함수

통합 테스트는 `//go:build integration` 빌드 태그를 사용하여 실제 .vital 파일이 있을 때만 실행됩니다.

### CI/CD 파이프라인

이 프로젝트는 GitHub Actions를 통한 자동화된 CI/CD 파이프라인을 제공합니다:

**자동 검사 항목**:
- ✅ **Multi-OS 테스트**: Ubuntu, macOS, Windows에서 자동 빌드 및 테스트
- ✅ **코드 품질**: golangci-lint를 통한 정적 분석
- ✅ **코드 커버리지**: Codecov.io를 통한 커버리지 추적 및 시각화
- ✅ **의존성 캐싱**: Go 모듈 및 빌드 캐시 자동 관리

**CI 워크플로우** (`.github/workflows/ci.yml`):
```yaml
# 모든 푸시 및 PR에서 자동 실행:
- Test (ubuntu-latest, macos-latest, windows-latest)
- Lint (golangci-lint)
- Coverage (Codecov 업로드)
```

**필요한 설정**:

오픈소스로 공개 시, [Codecov](https://codecov.io)에서 토큰을 발급받고 GitHub 저장소의 Secrets에 추가:
1. https://codecov.io에 접속하여 GitHub 계정으로 로그인
2. 저장소 추가 및 `CODECOV_TOKEN` 발급
3. GitHub 저장소 Settings → Secrets and variables → Actions
4. `CODECOV_TOKEN` 시크릿 추가

**로컬에서 CI와 동일하게 검증**:
```bash
# 모든 OS에서 실행되는 테스트 로컬 실행
go test ./vital -v
go test -tags=integration ./vital -v

# 린트 실행 (golangci-lint 설치 필요)
golangci-lint run --timeout=5m

# 커버리지 생성
go test ./... -covermode=atomic -coverprofile=coverage.out
```

**배지 추가** (README 상단):
```markdown
![CI](https://github.com/mdsung/vitaldb_processor/workflows/CI/badge.svg)
![Coverage](https://codecov.io/gh/mdsung/vitaldb_processor/badge.svg)
```

## 프로젝트 목표 및 설계 원칙

### 설계 원칙

**🎯 Python VitalDB = Golden Standard**

이 프로젝트의 핵심 원칙:
1. **정확도**: Python VitalDB (버퍼 오류 수정 적용)와 **100% 동일한 결과** 산출
2. **성능**: Python VitalDB보다 빠른 처리 속도
3. **호환성**: Python VitalDB가 지원하는 모든 파일 형식 지원

**중요**: Python VitalDB와 결과가 다르다면, 이는 Go 구현의 **버그**입니다. Python VitalDB의 출력이 정답입니다.

### 성능 목표

Go 구현은 다음을 목표로 합니다:
- ✅ Python VitalDB와 **동일한 데이터** 추출
- ✅ Python VitalDB보다 **빠른 처리 속도**
- ✅ Python VitalDB보다 **낮은 메모리 사용**

### 검증 방법

Go 구현의 정확성을 검증하려면:

```bash
# 1. Python VitalDB로 데이터 추출 (버퍼 오류 수정 적용)
python3 -c "
import vitaldb
vitaldb.utils.FMT_TYPE_LEN[7] = ('i', 4)
vitaldb.utils.FMT_TYPE_LEN[8] = ('I', 4)
vf = vitaldb.VitalFile('data.vital')
# ... 결과 저장
"

# 2. Go로 동일한 파일 처리
./vitaldb_processor -format json data.vital > go_output.json

# 3. 결과 비교 - 동일해야 함!
```

### 사용 사례

**Go VitalDB Processor 권장**:
- ✅ 프로덕션 시스템 (빠른 처리 속도 필요)
- ✅ 대용량 배치 처리
- ✅ 서버 환경에서 Python 설치 불가능한 경우
- ✅ 컨테이너/도커 환경 (단일 바이너리)

**Python VitalDB 권장**:
- ✅ 데이터 분석 (Pandas, NumPy 등과 함께 사용)
- ✅ 프로토타이핑 및 탐색적 분석
- ✅ Python 생태계 통합이 중요한 경우

**하이브리드 접근** (최적):
```bash
# 방법 1: MessagePack 사용 (최고 성능, 7.29배 빠름)
./vitaldb_processor -format msgpack -max-tracks 0 data.vital > output.msgpack
python analyze.py output.msgpack  # msgpack.unpackb() 사용

# 방법 2: JSON Compact 사용 (범용, 4.37배 빠름)
./vitaldb_processor -format json -compact -max-tracks 0 data.vital > output.json
python analyze.py output.json  # json.loads() 사용
```

**성능 비교**:
- Python VitalDB 직접 사용: 1,123ms
- Go + JSON Compact: 257ms (4.37배 빠름)
- Go + MessagePack: 154ms (7.29배 빠름) ⚡

자세한 벤치마크 결과는 [`notes/optimization_results.md`](notes/optimization_results.md)를 참조하세요.

## 라이센스

MIT License

## 기여

이슈나 풀 리퀘스트는 언제나 환영합니다!

## 관련 프로젝트

- [VitalDB](https://vitaldb.net/) - 의료 데이터베이스
- [VitalDB Python SDK](https://github.com/vitaldb/vitaldb-python) - 공식 Python SDK
