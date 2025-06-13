# 개발 계획 (TODO)

## 🔧 테스트 코드 개선 (우선순위: 높음)

### 현재 문제점

- `vital/vital_test.go` 파일이 13KB, 440줄로 너무 큼
- 하드코딩된 파일 경로로 인한 테스트 실패: `../../data/sample_vitalfiles/MICUA01_241114_130000.vital`
- 외부 파일 의존성으로 인한 CI/CD 불안정성
- 통합 테스트와 단위 테스트가 혼재

### 개선 계획

#### 1. 테스트 파일 분리

```
vital/
├── unit_test.go          # 단위 테스트 (mock 데이터)
├── integration_test.go   # 통합 테스트 (실제 파일)
├── benchmark_test.go     # 성능 테스트
└── testdata/            # 테스트용 데이터
    ├── small_sample.vital
    ├── mock_data.go
    └── README.md
```

#### 2. Mock 데이터 시스템 구축

- [ ] `testdata/mock_data.go` 생성
  - `CreateMockVitalFile()` 함수
  - `CreateMockWaveTrack()` 함수
  - `CreateMockNumericTrack()` 함수
  - `CreateMockStringTrack()` 함수
- [ ] `go:embed`를 사용한 테스트 데이터 임베딩
- [ ] 작은 크기의 실제 VitalDB 샘플 파일들 생성

#### 3. 테스트 코드 리팩토링

- [ ] **unit_test.go**: Mock 데이터를 사용한 빠른 단위 테스트
  - 데이터 파싱 정확성 테스트
  - 타입 안전성 테스트
  - 에러 핸들링 테스트
- [ ] **integration_test.go**: 실제 파일을 사용한 통합 테스트
  - `//go:build integration` 태그 사용
  - `go test -tags=integration` 명령어로 실행
- [ ] **benchmark_test.go**: 성능 측정 테스트
  - 파일 크기별 성능 벤치마크
  - 메모리 사용량 측정

#### 4. 테이블 드리븐 테스트 도입

```go
func TestDataParsing(t *testing.T) {
    tests := []struct {
        name     string
        input    []byte
        expected VitalFile
    }{
        {"WAVE track", mockWaveData, expectedWaveResult},
        {"NUMERIC track", mockNumericData, expectedNumericResult},
        {"STRING track", mockStringData, expectedStringResult},
    }
    // ...
}
```

#### 5. CI/CD 파이프라인 개선

- [ ] GitHub Actions 워크플로우 설정
- [ ] 단위 테스트: 모든 PR에서 실행
- [ ] 통합 테스트: 조건부 실행 (테스트 데이터 존재 시)
- [ ] 코드 커버리지 리포팅

---

## ✅ CLI 기능 개발 (완료됨!)

### Python 활용을 위한 명령줄 옵션 구현

#### 1. Flag 패키지 도입

- [x] 현재 `os.Args[1]` 방식을 `flag` 패키지로 교체
- [ ] `cobra` CLI 라이브러리 검토 (더 풍부한 기능)

#### 2. 출력 형식 옵션

- [x] `--format json` : JSON 출력
- [ ] `--format csv` : CSV 출력
- [x] `--summary` : 요약 정보만 출력

#### 3. 트랙 필터링 옵션

- [x] `--tracks "ECG_II,HR,PLETH"` : 특정 트랙만 추출
- [x] `--track-type WAVE|NUMERIC|STRING` : 타입별 필터링
- [ ] `--track-pattern "ECG*"` : 패턴 매칭

#### 4. 시간 범위 옵션

- [x] `--start-time 0 --end-time 300` : 시간 범위 지정
- [ ] `--start 5m --end 10m` : 시간 단위 지원

#### 5. 정보 조회 옵션

- [x] `--list-tracks` : 트랙 목록만 출력
- [x] `--info-only` : 파일 정보만 출력
- [x] `--list-devices` : 디바이스 정보 출력

#### 6. 출력 제어 옵션

- [x] `--max-samples 1000` : 샘플 개수 제한
- [x] `--quiet` : 에러만 출력
- [x] `--verbose` : 상세 출력

#### 7. 파일 출력 옵션

- [ ] `--output result.json` : 파일로 저장
- [ ] `--output-mmap result.mmap` : 메모리 맵 파일 출력

---

## 📚 문서화 개선 (우선순위: 낮음)

### API 문서

- [ ] GoDoc 스타일 주석 추가
- [ ] 예제 코드가 포함된 함수 문서
- [ ] `godoc` 서버로 문서 확인

### 사용 가이드

- [ ] `docs/` 디렉토리 생성
- [ ] `docs/getting-started.md`
- [ ] `docs/python-integration.md`
- [ ] `docs/api-reference.md`

### 개발자 가이드

- [ ] `CONTRIBUTING.md` 생성
- [ ] 코딩 스타일 가이드
- [ ] PR 템플릿 추가

---

## 🔧 코드 품질 개선 (우선순위: 낮음)

### 리팩토링

- [ ] `vital_optimized_v3_fixed.go` 파일명 정리
- [ ] 중복 코드 제거
- [ ] 함수 분리 (너무 긴 함수들)

### 에러 핸들링

- [ ] 커스텀 에러 타입 정의
- [ ] 에러 래핑 (`fmt.Errorf` → `errors.Wrap`)
- [ ] 에러 메시지 일관성

### 성능 최적화

- [ ] 메모리 풀 사용 검토
- [ ] 불필요한 메모리 할당 최소화
- [ ] 프로파일링 도구 활용

---

## 🛠 빌드 및 배포 (우선순위: 낮음)

### 크로스 플랫폼 빌드

- [ ] Makefile 생성
- [ ] Linux, macOS, Windows 바이너리 생성
- [ ] Docker 이미지 생성

### 패키지 관리

- [ ] Go 모듈 버전 관리
- [ ] Semantic versioning 도입
- [ ] Release 자동화

---

## 📝 완료된 작업

- [x] Python 활용 예시를 README.md에 추가
- [x] CLI 옵션 계획을 README.md에 문서화
- [x] 기본적인 VitalDB 파일 파싱 기능 구현
- [x] 타입 안전성을 고려한 데이터 구조 설계
- [x] **CLI 기능 대폭 개선 (2024.06.13)**
  - [x] Flag 패키지 도입으로 완전한 CLI 인터페이스 구현
  - [x] JSON 출력 형식 지원으로 Python 연동 최적화
  - [x] 트랙 제한 문제 해결 (`-max-tracks 0`으로 모든 트랙 출력)
  - [x] 디바이스 파싱 문제 해결
  - [x] 트랙 필터링 기능 (`-tracks`, `-track-type`)
  - [x] 시간 범위 필터링 (`-start-time`, `-end-time`)
  - [x] 정보 조회 옵션 (`-list-tracks`, `-info-only`, `-list-devices`)
  - [x] 출력 제어 옵션 (`-quiet`, `-verbose`, `-summary`, `-max-samples`)
  - [x] 데모 스크립트 생성 (`example/demo.py`)
  - [x] README.md 대폭 업데이트 (새로운 기능 및 Python 연동 가이드)

---

## 📅 예상 일정

### ~~Phase 1: CLI 기능 (완료!)~~

- [x] Flag 패키지 도입 및 기본 옵션 구현
- [x] JSON 출력 기능 추가
- [x] 트랙 필터링 및 시간 범위 기능

### Phase 2: 테스트 개선 (1-2주)

- 테스트 코드 분리 및 Mock 데이터 시스템 구축
- CI/CD 파이프라인 설정

### Phase 3: 고급 기능 (2-3주)

- CSV 출력 기능
- 패턴 매칭 필터링
- 성능 최적화

### Phase 4: 문서화 및 배포 (1-2주)

- 문서 정리 및 릴리스 준비

---

_이 문서는 지속적으로 업데이트됩니다. 완료된 항목은 체크박스에 표시하고 완료된 작업 섹션으로 이동시켜 주세요._
