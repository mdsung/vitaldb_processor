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

## 예제 실행

```bash
# 예제 프로그램 실행
cd example
go run main.go /path/to/your/file.vital
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
