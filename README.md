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

Go의 성능 이점을 활용하면서 Python에서 편리하게 사용할 수 있는 여러 방법을 제공합니다.

### 1. CLI를 통한 기본 활용

먼저 Go 프로그램을 빌드하고 Python에서 subprocess로 호출:

```bash
# Go 프로그램 빌드
go build -o vitaldb_processor example/main.go
```

```python
import subprocess
import json

# VitalDB 파일 처리
result = subprocess.run(['./vitaldb_processor', 'data.vital'],
                       capture_output=True, text=True)
print(result.stdout)
```

### 2. JSON 출력을 통한 데이터 분석

Go 프로그램에 JSON 출력 기능을 추가하여 구조화된 데이터로 받기:

```python
import subprocess
import json
import numpy as np
import matplotlib.pyplot as plt

# JSON 형태로 데이터 받기
result = subprocess.run(['./vitaldb_processor', '--json', 'data.vital'],
                       capture_output=True, text=True)
data = json.loads(result.stdout)

# 기본 정보 확인
print(f"시작 시간: {data['dt_start']}")
print(f"종료 시간: {data['dt_end']}")
print(f"트랙 개수: {len(data['tracks'])}")

# 특정 트랙 데이터 분석
if 'ECG_II' in data['tracks']:
    ecg_track = data['tracks']['ECG_II']
    print(f"ECG 샘플링 레이트: {ecg_track['sample_rate']} Hz")

    # 첫 번째 레코드의 파형 데이터 시각화
    if ecg_track['records']:
        first_record = ecg_track['records'][0]
        timestamps = first_record['dt']
        values = first_record['val']

        plt.figure(figsize=(12, 4))
        plt.plot(np.linspace(timestamps, timestamps + len(values)/ecg_track['sample_rate'], len(values)), values)
        plt.title('ECG II 파형')
        plt.xlabel('시간 (초)')
        plt.ylabel('진폭')
        plt.show()
```

### 3. 특정 트랙/변수 필터링

```python
import subprocess
import json

# 특정 트랙들만 추출
def get_tracks(file_path, track_names=None, track_type=None):
    cmd = ['./vitaldb_processor', '--format', 'json']

    if track_names:
        cmd.extend(['--tracks', ','.join(track_names)])

    if track_type:
        cmd.extend(['--track-type', track_type])

    cmd.append(file_path)

    result = subprocess.run(cmd, capture_output=True, text=True)
    return json.loads(result.stdout)

# ECG와 혈압 관련 트랙만 가져오기
vital_signs = get_tracks('data.vital', track_names=['ECG_II', 'ART', 'HR'])

# WAVE 타입 트랙들만 가져오기
wave_data = get_tracks('data.vital', track_type='WAVE')

# 수치형 데이터만 가져오기
numeric_data = get_tracks('data.vital', track_type='NUMERIC')
```

### 4. 시간 범위 기반 데이터 추출

```python
def get_time_range_data(file_path, start_time=0, end_time=None, tracks=None):
    """특정 시간 범위의 데이터만 추출"""
    cmd = ['./vitaldb_processor']
    cmd.extend(['--start-time', str(start_time)])

    if end_time:
        cmd.extend(['--end-time', str(end_time)])

    if tracks:
        cmd.extend(['--tracks', ','.join(tracks)])

    cmd.extend(['--format', 'json', file_path])

    result = subprocess.run(cmd, capture_output=True, text=True)
    return json.loads(result.stdout)

# 처음 5분간의 ECG 데이터
ecg_5min = get_time_range_data('data.vital',
                               start_time=0,
                               end_time=300,
                               tracks=['ECG_II'])

# 수술 중 특정 구간 (30분-60분)
surgery_data = get_time_range_data('data.vital',
                                   start_time=1800,
                                   end_time=3600)
```

### 5. 실시간 스트리밍 처리

```python
import subprocess
import json
import time

def stream_vital_data(file_path, window_size=10):
    """윈도우 단위로 데이터를 스트리밍 처리"""
    # 전체 파일 정보 먼저 확인
    info_result = subprocess.run(['./vitaldb_processor', '--info-only', file_path],
                                capture_output=True, text=True)
    file_info = json.loads(info_result.stdout)

    total_duration = file_info['dt_end'] - file_info['dt_start']
    current_time = file_info['dt_start']

    while current_time < file_info['dt_end']:
        end_time = min(current_time + window_size, file_info['dt_end'])

        # 현재 윈도우 데이터 가져오기
        window_data = get_time_range_data(file_path, current_time, end_time)

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
        result = subprocess.run(['./vitaldb_processor', '--summary', '--json', file_path],
                               capture_output=True, text=True, timeout=60)
        data = json.loads(result.stdout)

        return {
            'file': os.path.basename(file_path),
            'duration': data['dt_end'] - data['dt_start'],
            'tracks_count': len(data['tracks']),
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

### 7. 트랙 종류 확인

```python
def list_available_tracks(file_path):
    """VitalDB 파일의 모든 트랙 정보 확인"""
    result = subprocess.run(['./vitaldb_processor', '--list-tracks', '--json', file_path],
                           capture_output=True, text=True)
    tracks_info = json.loads(result.stdout)

    print("📊 사용 가능한 트랙들:")
    print("-" * 60)

    wave_tracks = []
    numeric_tracks = []
    string_tracks = []

    for name, info in tracks_info.items():
        track_type = info['type']
        if track_type == 1:  # WAVE
            wave_tracks.append(f"  • {name} ({info['unit']}, {info['sample_rate']} Hz)")
        elif track_type == 2:  # NUMERIC
            numeric_tracks.append(f"  • {name} ({info['unit']})")
        elif track_type == 5:  # STRING
            string_tracks.append(f"  • {name}")

    if wave_tracks:
        print("🌊 WAVE 트랙 (연속 파형):")
        print("\n".join(wave_tracks))
        print()

    if numeric_tracks:
        print("🔢 NUMERIC 트랙 (수치값):")
        print("\n".join(numeric_tracks))
        print()

    if string_tracks:
        print("📝 STRING 트랙 (이벤트/알람):")
        print("\n".join(string_tracks))

# 사용 예시
list_available_tracks('example.vital')
```

이러한 방법들을 통해 Python의 데이터 분석 생태계(pandas, numpy, matplotlib 등)와 Go의 고성능 파싱을 함께 활용할 수 있습니다.

## 예제 실행

```bash
# 예제 프로그램 실행
cd example
go run main.go /path/to/your/file.vital
```

## CLI 옵션 (향후 개발 예정)

Python에서 더 효과적으로 활용하기 위해 다음과 같은 CLI 옵션들을 추가할 예정입니다:

### 기본 사용법

```bash
./vitaldb_processor [options] <vital_file_path>
```

### 출력 형식 옵션

```bash
# JSON 형태로 출력
./vitaldb_processor --format json data.vital

# CSV 형태로 출력
./vitaldb_processor --format csv data.vital

# 요약 정보만 출력
./vitaldb_processor --summary data.vital
```

### 트랙 필터링 옵션

```bash
# 특정 트랙들만 추출
./vitaldb_processor --tracks "ECG_II,HR,PLETH" data.vital

# 트랙 타입별 필터링
./vitaldb_processor --track-type WAVE data.vital
./vitaldb_processor --track-type NUMERIC data.vital
./vitaldb_processor --track-type STRING data.vital

# 트랙 패턴 매칭
./vitaldb_processor --track-pattern "ECG*" data.vital
```

### 시간 범위 옵션

```bash
# 특정 시간 범위 추출
./vitaldb_processor --start-time 0 --end-time 300 data.vital

# 시간 단위 지정 (초, 분, 시간)
./vitaldb_processor --start 5m --end 10m data.vital
```

### 정보 조회 옵션

```bash
# 트랙 목록만 출력
./vitaldb_processor --list-tracks data.vital

# 파일 정보만 출력
./vitaldb_processor --info-only data.vital

# 디바이스 정보 출력
./vitaldb_processor --list-devices data.vital
```

### 출력 제어 옵션

```bash
# 샘플 개수 제한
./vitaldb_processor --max-samples 1000 data.vital

# 조용한 모드 (에러만 출력)
./vitaldb_processor --quiet data.vital

# 상세 모드
./vitaldb_processor --verbose data.vital
```

### 파일 출력 옵션

```bash
# 파일로 저장
./vitaldb_processor --output result.json data.vital

# 메모리 맵 파일로 출력 (고성능)
./vitaldb_processor --output-mmap result.mmap data.vital
```

### 사용 예시

```bash
# ECG 데이터만 처음 5분간 JSON으로 추출
./vitaldb_processor --tracks "ECG_II" --start-time 0 --end-time 300 --format json data.vital

# 모든 수치형 데이터를 CSV로 저장
./vitaldb_processor --track-type NUMERIC --format csv --output vitals.csv data.vital

# 파일 정보 빠르게 확인
./vitaldb_processor --info-only --quiet data.vital
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
