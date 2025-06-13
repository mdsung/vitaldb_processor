package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mdsung/vitaldb_processor/vital"
)

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

	// 첫 번째 트랙의 샘플 데이터 출력
	if len(vf.Order) > 0 {
		firstTrackName := vf.Order[0]
		if track, exists := vf.Trks[firstTrackName]; exists && len(track.Recs) > 0 {
			fmt.Printf("\n=== Sample Data from '%s' ===\n", firstTrackName)
			for i, rec := range track.Recs {
				if i >= 5 { // 처음 5개만 출력
					break
				}
				fmt.Printf("Time: %f, Value: %v\n", rec.Dt, rec.Val)
			}
			if len(track.Recs) > 5 {
				fmt.Printf("... and %d more records\n", len(track.Recs)-5)
			}
		}
	}
}
