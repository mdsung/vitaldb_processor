---
tags: [vitaldb, python, bugfix, critical]
date: 2025-11-16
status: documented
priority: critical
---

# Python VitalDB 라이브러리 버퍼 오류 수정

## 문제 (Problem)

Python VitalDB 라이브러리 (vitaldb-python)를 사용할 때, 특정 트랙 포맷 (타입 7, 8)에서 **"buffer is too small"** 오류가 발생합니다.

## 필수 해결 방법 (Critical Fix)

Python VitalDB 라이브러리를 사용하기 **전에 반드시** 다음 코드를 실행해야 합니다:

```python
import vitaldb

# 필수: 버퍼 오류 방지를 위한 포맷 타입 길이 설정
vitaldb.utils.FMT_TYPE_LEN[7] = ("i", 4)  # signed integer, 4 bytes
vitaldb.utils.FMT_TYPE_LEN[8] = ("I", 4)  # unsigned integer, 4 bytes

# 이제 안전하게 VitalDB 파일 로드 가능
vf = vitaldb.VitalFile('data.vital')
```

## 기술적 배경 (Technical Background)

### FMT_TYPE_LEN이란?

`vitaldb.utils.FMT_TYPE_LEN`은 VitalDB 파일의 바이너리 데이터를 파싱할 때 사용되는 포맷 타입과 길이 매핑 딕셔너리입니다.

- **Key**: 포맷 타입 번호 (0-8)
- **Value**: `(struct_format, byte_length)` 튜플
  - `struct_format`: Python `struct` 모듈의 포맷 문자열 (`i` = signed int, `I` = unsigned int)
  - `byte_length`: 바이트 단위 크기

### 왜 타입 7과 8이 문제인가?

Python VitalDB 라이브러리의 기본 구현에서:
- 타입 7 (signed integer)과 타입 8 (unsigned integer)의 길이가 잘못 설정되어 있거나
- 아예 정의되지 않아서

바이너리 버퍼를 읽을 때 "buffer is too small" 오류가 발생합니다.

### 수정 코드의 의미

```python
vitaldb.utils.FMT_TYPE_LEN[7] = ("i", 4)
# 타입 7: 4바이트 signed integer (int32)

vitaldb.utils.FMT_TYPE_LEN[8] = ("I", 4)
# 타입 8: 4바이트 unsigned integer (uint32)
```

## 영향을 받는 파일

이 버퍼 오류는 특정 VitalDB 파일에서만 발생합니다:
- 타입 7 또는 타입 8 포맷을 사용하는 트랙을 포함한 파일
- 실제 VitalDB 데이터베이스에서 다운로드한 일부 파일들

## Go VitalDB Processor vs Python VitalDB

| 항목 | Python VitalDB | Go VitalDB Processor |
|------|----------------|----------------------|
| **버퍼 오류** | ❌ 발생 (수정 필요) | ✅ 없음 |
| **수정 필요** | ✅ 필수 | ❌ 불필요 |
| **신뢰성** | ⚠️ 패치 후 안정 | ✅ 기본적으로 안정 |

**Go VitalDB Processor는 이러한 버퍼 문제가 없으며**, 모든 포맷 타입을 올바르게 처리합니다.

## 권장 사항 (Recommendations)

### 1. Go Binary 사용 (최우선 권장)

```python
import subprocess
import json

def load_vital_data(file_path):
    cmd = ['./vitaldb_processor', '-format', 'json', file_path]
    result = subprocess.run(cmd, capture_output=True, text=True)
    return json.loads(result.stdout)

# 버퍼 오류 걱정 없이 사용 가능
data = load_vital_data('data.vital')
```

**장점**:
- ✅ 버퍼 오류 없음
- ✅ 패치 불필요
- ✅ 더 빠른 성능
- ✅ 100% 트랙 로딩 성공률

### 2. Python VitalDB 직접 사용 (필요시)

```python
import vitaldb

# 필수: 버퍼 오류 수정 (매번 import 후 실행)
vitaldb.utils.FMT_TYPE_LEN[7] = ("i", 4)
vitaldb.utils.FMT_TYPE_LEN[8] = ("I", 4)

# 이제 사용 가능
vf = vitaldb.VitalFile('data.vital')
```

**주의사항**:
- ⚠️ 매번 import 후 반드시 설정 필요
- ⚠️ 다른 Python VitalDB 관련 이슈 존재 가능

## 코드 예시: 완전한 Python 스크립트

### 패치 적용 예시

```python
#!/usr/bin/env python3
"""
VitalDB 파일을 Python VitalDB 라이브러리로 읽는 예시
버퍼 오류 수정 포함
"""

import vitaldb
import sys

def fix_vitaldb_buffer():
    """VitalDB 버퍼 오류 수정 (필수)"""
    vitaldb.utils.FMT_TYPE_LEN[7] = ("i", 4)
    vitaldb.utils.FMT_TYPE_LEN[8] = ("I", 4)
    print("✓ VitalDB buffer fix applied")

def load_vital_file(filepath):
    """VitalDB 파일 로드"""
    try:
        vf = vitaldb.VitalFile(filepath)
        print(f"✓ Loaded: {filepath}")
        print(f"  Tracks: {len(vf.trks)}")
        print(f"  Duration: {vf.dtend - vf.dtstart:.2f}s")
        return vf
    except Exception as e:
        print(f"✗ Error loading file: {e}")
        return None

if __name__ == '__main__':
    # 1. 필수: 버퍼 오류 수정 적용
    fix_vitaldb_buffer()

    # 2. 파일 로드
    if len(sys.argv) < 2:
        print("Usage: python script.py <vital_file>")
        sys.exit(1)

    vf = load_vital_file(sys.argv[1])

    # 3. 데이터 사용
    if vf:
        for track_name, track_data in vf.trks.items():
            print(f"  • {track_name}: {len(track_data)} records")
```

## 관련 문서

- README.md: "Python에서 활용하기" > "중요: Python VitalDB 라이브러리 버퍼 오류 수정"
- CLAUDE.md: "Python Integration Pattern" > "Option 2: Using Python VitalDB Library Directly"

## 업데이트 이력

- 2025-11-16: 초기 문서 작성, 버퍼 오류 수정 방법 기록

---

**중요**: 이 정보는 프로젝트 전체에서 Python VitalDB를 사용하는 모든 경우에 필수적으로 적용되어야 합니다.
