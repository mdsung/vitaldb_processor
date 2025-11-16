---
tags: [vitaldb, profiling, performance, bottleneck]
date: 2025-11-16
status: completed
priority: critical
---

# VitalDB Processor: 프로파일링 결과 및 병목 분석

## 실행 환경

**테스트 파일**: MICUB08_240520_230000.vital
- 크기: 3.12 MB
- 트랙 수: 52개
- 레코드 수: 80,037개

**실행 명령**:
```bash
./vitaldb_processor -format json -max-tracks 0 -max-samples 0 \
  -cpuprofile cpu.prof -memprofile mem.prof data.vital > /dev/null
```

**총 실행 시간**: 605.28ms

## CPU 프로파일링 결과 (420ms 샘플)

### 🔴 병목 #1: JSON 마샬링 (71.43%)

| 함수 | 시간 | 비율 | 설명 |
|------|------|------|------|
| `encoding/json.MarshalIndent` | 300ms | **71.43%** | JSON 전체 변환 |
| `encoding/json.appendIndent` | 150ms | 35.71% | 들여쓰기 처리 |
| `encoding/json.floatEncoder` | 140ms | 33.33% | float64 변환 |
| `encoding/json.interfaceEncoder` | 120ms | 28.57% | interface{} 타입 처리 |
| `encoding/json.arrayEncoder` | 150ms | 35.71% | 배열 인코딩 |

**문제점**:
- 전체 실행 시간의 **50%가 JSON 마샬링**
- float64를 JSON 문자열로 변환하는 비용이 매우 큼
- 들여쓰기(pretty print)로 인한 추가 오버헤드

### ⚪ VitalDB 파싱 (11.90%)

| 함수 | 시간 | 비율 | 설명 |
|------|------|------|------|
| `compress/flate` 관련 | 50ms | 11.90% | gzip 압축 해제 |
| `vital.NewVitalFile` | 50ms | 11.90% | 파일 읽기 및 파싱 |

**결론**: VitalDB 파일 파싱은 매우 빠르고 효율적

### 기타

| 함수 | 시간 | 비율 |
|------|------|------|
| `runtime.memmove` | 90ms | 21.43% |
| `strconv.fmtF` | 70ms | 16.67% |

## 메모리 프로파일링 결과 (192.90MB 할당)

### 🔴 병목 #1: JSON 마샬링 (83.55%)

| 함수 | 할당량 | 비율 | 설명 |
|------|--------|------|------|
| `encoding/json.MarshalIndent` | 161.17MB | **83.55%** | JSON 버퍼 전체 |
| `bytes.growSlice` | 61.95MB | 32.11% | JSON 버퍼 확장 |
| `encoding/json.appendNewline` | 45.10MB | 23.38% | 줄바꿈 문자 추가 |
| `encoding/json.Marshal` | 79.99MB | 41.47% | 실제 인코딩 |

**문제점**:
- 전체 메모리의 **84%가 JSON 출력 생성**
- `bytes.Buffer`가 반복적으로 확장 (2배씩)
- 들여쓰기와 줄바꿈으로 메모리 2배 증가

### ⚪ VitalDB 파싱 (약 12%)

| 함수 | 할당량 | 비율 | 설명 |
|------|--------|------|------|
| `vital.NewVitalFile` | 24.51MB | 12.71% | 파일 로딩 |
| `vital.parseWaveData` | 15.47MB | 8.02% | 파형 데이터 |
| `vital.parseNumericData` | 3.54MB | 1.83% | 수치 데이터 |
| `main.processTracks` | 6.07MB | 3.15% | 트랙 처리 |

**결론**: 실제 데이터 파싱은 메모리 효율적 (~25MB)

## 핵심 발견 (Key Findings)

### 1. JSON 마샬링이 압도적 병목

```
전체 성능 분해:
├─ VitalDB 파일 파싱: 50ms (8.3%)  ✅ 빠름
├─ 데이터 처리: ~50ms (8.3%)      ✅ 빠름
└─ JSON 마샬링: 300ms (50%)        ❌ 병목!
   └─ 프로세스 오버헤드: ~200ms    ❌ 병목!
```

### 2. 메모리 사용 패턴

```
메모리 할당 분해:
├─ VitalDB 데이터: ~25MB (13%)    ✅ 효율적
└─ JSON 출력: ~162MB (84%)         ❌ 비효율적
```

### 3. float64 변환이 비쌈

**80,037개 레코드 × 평균 10개 float 값 = 약 800,000개 float64를 JSON 문자열로 변환**

`encoding/json.floatEncoder`: 140ms (33%)

## 최적화 전략 (Optimization Strategies)

### 전략 1: 스트리밍 JSON 인코더 ⭐

**현재**:
```go
jsonData, err := json.MarshalIndent(output, "", "  ")
fmt.Println(string(jsonData))
```

**개선**:
```go
encoder := json.NewEncoder(os.Stdout)
encoder.SetIndent("", "  ")
encoder.Encode(output)
```

**예상 효과**:
- 메모리 할당: 162MB → **80MB** (50% 감소)
- 버퍼 확장 오버헤드 제거
- CPU 시간: 10-15% 감소

### 전략 2: 들여쓰기 제거 (옵션)

**현재**: Pretty-print JSON (2칸 들여쓰기)
**개선**: Compact JSON

**예상 효과**:
- `appendIndent` 150ms 제거
- `appendNewline` 45MB 할당 제거
- **CPU 25% 감소, 메모리 30% 감소**

**트레이드오프**: JSON 가독성 저하 (Python에서는 상관없음)

### 전략 3: 메모리 프리할당

**현재**:
```go
records := make([]RecordInfo, 0)  // 용량 0
```

**개선**:
```go
records := make([]RecordInfo, 0, len(track.Recs))  // 용량 사전 확보
```

**예상 효과**: 5-10% 성능 향상

### 전략 4: 라이브러리 모드 (장기)

JSON 마샬링 + 프로세스 오버헤드 = **500ms (83%)**

**개선**: cgo로 Python에서 직접 호출
- JSON 불필요
- 프로세스 생성 불필요

**예상 효과**: **5-10배 빠름**

## 즉시 실행 가능한 최적화 (Quick Wins)

### 1단계: 스트리밍 JSON (5분)
```go
// example/main.go:112
encoder := json.NewEncoder(os.Stdout)
if !config.Compact {
    encoder.SetIndent("", "  ")
}
encoder.Encode(output)
```

**예상 효과**:
- 현재: 605ms → **530ms** (12% 빠름)
- 메모리: 193MB → **110MB** (43% 감소)

### 2단계: Compact JSON 옵션 추가 (2분)
```go
flag.BoolVar(&config.Compact, "compact", false, "Compact JSON (no indentation)")
```

**예상 효과** (Compact 모드):
- 현재: 605ms → **380ms** (37% 빠름)
- 메모리: 193MB → **80MB** (59% 감소)

### 3단계: 메모리 프리할당 (5분)
```go
// example/main.go:230
records := make([]RecordInfo, 0, len(track.Recs))
```

**예상 효과**: 추가 5-8% 빠름

### 종합 예상 성능 (Quick Wins 적용 후)

| 항목 | 현재 | 개선 후 | 비율 |
|------|------|---------|------|
| **CPU 시간** | 605ms | **350ms** | **1.73x** |
| **메모리** | 193MB | **75MB** | **2.57x** |

**Python 대비 성능** (3.12MB 파일):
- 현재: Python 1123ms vs Go 623ms = 1.80x
- 개선 후: Python 1123ms vs Go **350ms** = **3.21x** ✅

## 다음 단계 (Next Steps)

1. ✅ **즉시**: 스트리밍 JSON + Compact 옵션 구현
2. 🔄 **단기**: 메모리 프리할당 적용
3. 📊 **중기**: 재프로파일링 및 추가 최적화
4. 🚀 **장기**: cgo 라이브러리 모드 구현

## 프로파일 파일 위치

- CPU 프로파일: `benchmark/cpu.prof`
- 메모리 프로파일: `benchmark/mem.prof`

### 프로파일 분석 명령어

```bash
# CPU 프로파일 (Top 함수)
go tool pprof -top -cum benchmark/cpu.prof

# CPU 프로파일 (Flame graph)
go tool pprof -http=:8080 benchmark/cpu.prof

# 메모리 프로파일 (할당량)
go tool pprof -top -alloc_space benchmark/mem.prof

# 메모리 프로파일 (현재 사용량)
go tool pprof -top -inuse_space benchmark/mem.prof
```

---

**분석 완료일**: 2025-11-16
**테스트 파일**: MICUB08_240520_230000.vital (3.12 MB)
**핵심 발견**: JSON 마샬링이 CPU 50%, 메모리 84% 차지
**즉시 조치**: 스트리밍 JSON 인코더로 30-40% 성능 향상 가능
