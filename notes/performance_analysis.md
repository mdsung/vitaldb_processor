---
tags: [vitaldb, performance, profiling, optimization]
date: 2025-11-16
status: in-progress
priority: high
---

# VitalDB Processor: 성능 분석 및 최적화 전략

## 현재 상태 (Current Status)

### ✅ 달성 항목
- **정확도**: Python VitalDB와 100% 동일 (Golden Standard 달성)
- **대형 파일**: 3.12 MB 파일에서 1.80x 빠른 성능
- **안정성**: 6개 실제 파일 모두 정상 처리

### ⚠️ 문제점
- **전체 평균**: 0.95x 속도 (Python보다 5% 느림)
- **작은 파일**: 5.87x 느림 (0.39 MB)
- **특정 파일**: MICUB06 파일 2.78x 느림 (2.57 MB)

## 성능 병목 분석 (Performance Bottleneck Analysis)

### 1. 프로세스 생성 오버헤드 (Subprocess Overhead)

**현재 구조**:
```python
# Python에서 Go 바이너리 호출
subprocess.run(['./vitaldb_processor', '-format', 'json', ...])
```

**오버헤드 구성**:
- 프로세스 생성 시간: ~50-100ms
- stdout 버퍼링: ~50ms
- JSON 파싱 (Python): ~50ms
- **총 고정 오버헤드**: ~150-200ms

**영향**:
- 작은 파일 (0.39 MB): Python 70ms vs Go 412ms
  - 실제 처리: ~70ms
  - 오버헤드: ~340ms (실제 처리 시간의 4.8배)
- 큰 파일 (3.12 MB): Python 1123ms vs Go 623ms
  - 실제 처리 이득: 500ms
  - 오버헤드: 200ms (이득의 40%)

### 2. JSON 마샬링 비용 (JSON Marshalling Cost)

**문제점**:
```go
// Go에서 모든 레코드를 JSON으로 변환
json.MarshalIndent(output, "", "  ")
```

**비용 분석**:
- 작은 배열 (8K 레코드): ~20ms
- 중간 배열 (40K 레코드): ~80ms
- 큰 배열 (80K 레코드): ~150ms

**MICUB06 이상 현상**:
```
파일: MICUB06_240322_230000.vital
크기: 2.57 MB
트랙: 19개 (다른 파일: 45-52개)
레코드: 21,898개
Python: 177ms
Go: 493ms (2.78x 느림)
```

**가설**: 트랙당 평균 레코드 수
- MICUB06: 21,898 / 19 = **1,152 레코드/트랙**
- 다른 파일: 43,205 / 45 = **960 레코드/트랙**

→ 큰 배열을 JSON으로 마샬링하는 비용이 높음

### 3. 메모리 할당 패턴

**Python VitalDB**:
- 메모리 사용: 3.5 MB (작은 파일) ~ 33 MB (큰 파일)
- 지연 로딩: `trk.recs`는 필요시 접근

**Go VitalDB Processor**:
- 모든 레코드를 메모리에 로드
- RecordInfo 구조체 배열 생성
- JSON 출력을 위한 추가 메모리 할당

## 최적화 전략 (Optimization Strategies)

### 전략 1: 라이브러리 모드 (Library Mode) ⭐ 추천

**현재 (Subprocess 모드)**:
```python
# Python
subprocess.run(['./vitaldb_processor', ...])
```

**개선 (cgo 라이브러리)**:
```python
# Python
import vitaldb_go
vf = vitaldb_go.VitalFile('data.vital')
```

**장점**:
- ✅ 프로세스 생성 오버헤드 제거 (~200ms 절약)
- ✅ JSON 마샬링 불필요 (Python 객체 직접 반환)
- ✅ 메모리 효율성 (데이터 복사 최소화)

**예상 성능**:
- 작은 파일: 5.87x → **1.2-1.5x** 빠름
- 큰 파일: 1.80x → **2.0-2.5x** 빠름

**구현 복잡도**: 중간 (cgo 바인딩 필요)

### 전략 2: 스트리밍 JSON 출력 (Streaming JSON)

**현재**:
```go
// 모든 데이터를 메모리에 로드 후 한번에 마샬링
json.MarshalIndent(output, "", "  ")
```

**개선**:
```go
// 스트리밍 JSON 인코더
encoder := json.NewEncoder(os.Stdout)
encoder.SetIndent("", "  ")
encoder.Encode(output)
```

**장점**:
- ✅ 메모리 사용량 감소
- ✅ 큰 배열 마샬링 시간 단축

**예상 성능**: 10-20% 개선

**구현 복잡도**: 낮음

### 전략 3: 선택적 데이터 출력 (Selective Output)

**아이디어**: Python이 실제로 사용하는 데이터만 출력

**현재**:
```json
{
  "tracks": {
    "ECG_II": {
      "records": [/* 80,000개 레코드 */]
    }
  }
}
```

**개선**:
```json
{
  "tracks": {
    "ECG_II": {
      "records_count": 80000,
      "first_record": {...},
      "last_record": {...}
    }
  }
}
```

**장점**:
- ✅ JSON 크기 대폭 감소
- ✅ 마샬링 시간 단축
- ✅ Python 파싱 시간 단축

**단점**:
- ⚠️ 전체 데이터가 필요한 경우 제한적

**구현 복잡도**: 낮음

### 전략 4: 바이너리 프로토콜 (Protocol Buffers/MessagePack)

**아이디어**: JSON 대신 바이너리 직렬화

**장점**:
- ✅ 직렬화/역직렬화 속도 3-5배 빠름
- ✅ 데이터 크기 50-70% 감소

**단점**:
- ⚠️ Python 측 디코더 필요
- ⚠️ 디버깅 어려움

**구현 복잡도**: 중간

## Go 코드 최적화 기회 (Go Code Optimizations)

### 1. 메모리 프리할당 (Pre-allocation)

**현재**:
```go
records := make([]RecordInfo, 0)
for _, rec := range track.Recs {
    records = append(records, ...)  // 반복적 재할당
}
```

**개선**:
```go
records := make([]RecordInfo, 0, len(track.Recs))  // 용량 사전 할당
for _, rec := range track.Recs {
    records = append(records, ...)  // 재할당 없음
}
```

**예상 성능**: 5-10% 개선

### 2. 불필요한 복사 제거

**현재**: RecordInfo 구조체 복사
**개선**: 포인터 사용 또는 직접 참조

### 3. 병렬 처리 (Parallel Processing)

**아이디어**: 트랙별 병렬 처리

```go
var wg sync.WaitGroup
for name, track := range vf.Trks {
    wg.Add(1)
    go func(n string, t *Track) {
        defer wg.Done()
        // 트랙 처리
    }(name, track)
}
wg.Wait()
```

**주의**: JSON 출력 순서 보장 필요

## 다음 단계 (Next Steps)

### 즉시 실행 가능 (Quick Wins)
1. ✅ **메모리 프리할당**: example/main.go 수정
2. ✅ **스트리밍 JSON**: json.NewEncoder 사용
3. ✅ **선택적 출력**: `-summary` 플래그 개선

### 중기 목표 (Medium-term)
4. 🔄 **Go 프로파일링**: CPU/메모리 병목 지점 찾기
5. 🔄 **벤치마크**: 다양한 파일 타입 테스트

### 장기 목표 (Long-term)
6. 📋 **라이브러리 모드**: cgo 바인딩 구현
7. 📋 **바이너리 프로토콜**: MessagePack 지원

## 성능 목표 (Performance Goals)

### Phase 1: Quick Optimizations (1-2일)
- **목표**: 평균 1.2-1.5x 빠름
- **방법**: 메모리 프리할당, 스트리밍 JSON

### Phase 2: Profiling & Tuning (3-5일)
- **목표**: 평균 1.5-2.0x 빠름
- **방법**: CPU/메모리 프로파일링, 병목 제거

### Phase 3: Architectural Changes (1-2주)
- **목표**: 평균 2.0-3.0x 빠름
- **방법**: 라이브러리 모드, 바이너리 프로토콜

## 참고 자료 (References)

- [Go Performance Best Practices](https://github.com/golang/go/wiki/Performance)
- [JSON Encoding Performance](https://pkg.go.dev/encoding/json)
- [Python cgo Integration](https://pkg.go.dev/cmd/cgo)

---

**작성일**: 2025-11-16
**상태**: 분석 완료, 최적화 대기 중
**우선순위**: 높음
