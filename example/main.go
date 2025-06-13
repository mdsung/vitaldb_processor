package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mdsung/vitaldb_processor/vital"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <vital_file_path>")
		os.Exit(1)
	}

	filepath := os.Args[1]

	// VitalDB 파일 읽기
	fmt.Printf("Reading VitalDB file: %s\n", filepath)
	vf, err := vital.NewVitalFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	// 기본 정보 출력
	fmt.Printf("\n=== File Information ===\n")
	fmt.Printf("Start Time: %f\n", vf.DtStart)
	fmt.Printf("End Time: %f\n", vf.DtEnd)
	fmt.Printf("GMT Offset: %d\n", vf.Dgmt)
	fmt.Printf("Number of Tracks: %d\n", len(vf.Trks))
	fmt.Printf("Number of Devices: %d\n", len(vf.Devs))

	// 디바이스 정보 출력
	fmt.Printf("\n=== Devices ===\n")
	for name, device := range vf.Devs {
		fmt.Printf("- %s: %s (Port: %s)\n", name, device.TypeName, device.Port)
	}

	// 트랙 정보 출력 (처음 10개만)
	fmt.Printf("\n=== Tracks (first 10) ===\n")
	count := 0
	for name, track := range vf.Trks {
		if count >= 10 {
			break
		}
		fmt.Printf("- %s: %s, Rate: %.1f Hz, Records: %d\n",
			name, track.Unit, track.SRate, len(track.Recs))
		count++
	}

	if len(vf.Trks) > 10 {
		fmt.Printf("... and %d more tracks\n", len(vf.Trks)-10)
	}

	// 첫 번째 트랙의 샘플 데이터 출력 (타입 안전성을 고려한 접근)
	if len(vf.Order) > 0 {
		firstTrackName := vf.Order[0]
		if track, exists := vf.Trks[firstTrackName]; exists && len(track.Recs) > 0 {
			fmt.Printf("\n=== Sample Data from '%s' ===\n", firstTrackName)
			for i, rec := range track.Recs {
				if i >= 3 { // 처음 3개만 출력 (비교를 위해)
					break
				}

				fmt.Printf("\n--- Record %d ---\n", i+1)

				// 기존 방식 (raw value)
				fmt.Printf("기존 방식 - Raw Value: %v\n", rec.Val)
				fmt.Printf("기존 방식 - Type: %T\n", rec.Val)

				// 새로운 타입 안전성 방식
				if numVal, ok := rec.GetNumericValue(); ok {
					fmt.Printf("새 방식 - Numeric Value: %.6f\n", numVal)
				} else if strVal, ok := rec.AsString(); ok {
					fmt.Printf("새 방식 - String Value: %s\n", strVal)
				} else if arr32, ok := rec.AsFloat32Array(); ok {
					fmt.Printf("새 방식 - Float32 Array (%d samples): 처음 5개 = %v\n", len(arr32), arr32[:min(5, len(arr32))])
					fmt.Printf("기존 방식과 비교 - 첫 번째 값: %.6f vs %.6f\n", arr32[0], rec.Val.([]float32)[0])
				} else if arr64, ok := rec.AsFloat64Array(); ok {
					fmt.Printf("새 방식 - Float64 Array (%d samples): 처음 5개 = %v\n", len(arr64), arr64[:min(5, len(arr64))])
					fmt.Printf("기존 방식과 비교 - 첫 번째 값: %.6f vs %.6f\n", arr64[0], rec.Val.([]float64)[0])
				} else {
					fmt.Printf("새 방식 - Unknown type: %v\n", rec.Val)
				}
			}
		}
	}
}
