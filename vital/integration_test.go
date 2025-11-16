//go:build integration

package vital

import (
	"testing"
)

// TestVitalFileParse tests basic VitalFile parsing functionality.
// This test requires actual .vital files and should be run with -tags=integration.
func TestVitalFileParse(t *testing.T) {
	vf, err := NewVitalFile("../data_sample/MICUA01_240724_180000.vital")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	t.Logf("Devices found: %d", len(vf.Devs))
	for name, dev := range vf.Devs {
		t.Logf("Device: %s (%s)", name, dev.TypeName)
	}

	t.Logf("Tracks found: %d", len(vf.Trks))
	if len(vf.Trks) == 0 {
		t.Fatalf("no tracks parsed")
	}

	// Order 배열 순서대로 출력
	t.Logf("Track order: %d tracks", len(vf.Order))
	for i, trackName := range vf.Order {
		t.Logf("Track %d: %s", i+1, trackName)
		if i >= 9 { // 처음 10개만
			t.Logf("  ... (showing first 10 of %d tracks)", len(vf.Order))
			break
		}
	}
}

// TestREC데이터파싱 tests REC (record) data parsing.
func TestREC데이터파싱(t *testing.T) {
	vf, err := NewVitalFile("../data_sample/MICUA01_240724_180000.vital")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	t.Logf("=== REC 데이터 파싱 테스트 ===")

	totalRecs := 0
	for trackName, track := range vf.Trks {
		recCount := len(track.Recs)
		totalRecs += recCount
		if recCount > 0 {
			t.Logf("Track %s: %d records", trackName, recCount)

			// 첫 번째 레코드 예시 출력
			firstRec := track.Recs[0]
			switch track.Type {
			case 1: // WAVE
				if waveData, ok := firstRec.Val.([]byte); ok {
					t.Logf("  First wave record: dt=%f, samples=%d bytes", firstRec.Dt, len(waveData))
				}
			case 2: // NUMERIC
				if val, ok := firstRec.Val.(float64); ok {
					t.Logf("  First numeric record: dt=%f, val=%f", firstRec.Dt, val)
				}
			case 5: // STRING
				if val, ok := firstRec.Val.(string); ok {
					t.Logf("  First string record: dt=%f, val='%s'", firstRec.Dt, val)
				}
			}
		}
	}

	t.Logf("Total records across all tracks: %d", totalRecs)

	if totalRecs == 0 {
		t.Errorf("No REC data was parsed")
	}
}

// TestRECValuesDetail tests the accuracy of REC values in detail.
func TestRECValuesDetail(t *testing.T) {
	vf, err := NewVitalFile("../data_sample/MICUA01_240724_180000.vital")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	t.Logf("=== REC 값 상세 검증 ===")

	// 각 타입별 상세 테스트
	for trackName, track := range vf.Trks {
		if len(track.Recs) > 0 {
			switch track.Type {
			case 1: // WAVE 데이터
				t.Logf("WAVE Track: %s (fmt=%d, srate=%f)", trackName, track.Fmt, track.SRate)
				for i := 0; i < min(3, len(track.Recs)); i++ {
					rec := track.Recs[i]
					switch waveData := rec.Val.(type) {
					case []float32:
						t.Logf("  Record %d: dt=%f, samples=%d ([]float32)", i, rec.Dt, len(waveData))
						if len(waveData) > 0 {
							t.Logf("    First 4 samples: %v", waveData[:min(4, len(waveData))])
						}
					case []float64:
						t.Logf("  Record %d: dt=%f, samples=%d ([]float64)", i, rec.Dt, len(waveData))
						if len(waveData) > 0 {
							t.Logf("    First 4 samples: %v", waveData[:min(4, len(waveData))])
						}
					default:
						t.Errorf("  Record %d: expected []float32 or []float64, got %T", i, rec.Val)
					}
				}

			case 2: // NUMERIC 데이터
				t.Logf("NUMERIC Track: %s (fmt=%d, unit=%s)", trackName, track.Fmt, track.Unit)
				for i := 0; i < min(5, len(track.Recs)); i++ {
					rec := track.Recs[i]
					// fmt에 따라 다른 타입 기대
					switch track.Fmt {
					case 1: // float32
						if val, ok := rec.Val.(float32); ok {
							t.Logf("  Record %d: dt=%f, val=%f (float32)", i, rec.Dt, val)
							// 값이 합리적인 범위인지 확인
							if val < -1e10 || val > 1e10 {
								t.Logf("    Warning: value seems out of normal range")
							}
						} else {
							t.Errorf("  Record %d: expected float32, got %T", i, rec.Val)
						}
					case 2: // float64
						if val, ok := rec.Val.(float64); ok {
							t.Logf("  Record %d: dt=%f, val=%f (float64)", i, rec.Dt, val)
							// 값이 합리적인 범위인지 확인
							if val < -1e10 || val > 1e10 {
								t.Logf("    Warning: value seems out of normal range")
							}
						} else {
							t.Errorf("  Record %d: expected float64, got %T", i, rec.Val)
						}
					default: // 정수 타입들은 float64로 변환됨
						if val, ok := rec.Val.(float64); ok {
							t.Logf("  Record %d: dt=%f, val=%f (converted to float64)", i, rec.Dt, val)
						} else {
							t.Errorf("  Record %d: expected float64 (converted), got %T", i, rec.Val)
						}
					}
				}

			case 5: // STRING 데이터
				t.Logf("STRING Track: %s", trackName)
				for i := 0; i < min(3, len(track.Recs)); i++ {
					rec := track.Recs[i]
					if val, ok := rec.Val.(string); ok {
						t.Logf("  Record %d: dt=%f, val='%s'", i, rec.Dt, val)
						if len(val) == 0 {
							t.Logf("    Warning: empty string value")
						}
					} else {
						t.Errorf("  Record %d: expected string, got %T", i, rec.Val)
					}
				}
			}
		}
	}

	// 시간 순서 확인
	t.Logf("\n=== 시간 순서 검증 ===")
	for trackName, track := range vf.Trks {
		if len(track.Recs) >= 2 {
			isOrdered := true
			for i := 1; i < len(track.Recs); i++ {
				if track.Recs[i].Dt < track.Recs[i-1].Dt {
					isOrdered = false
					break
				}
			}
			if !isOrdered {
				t.Logf("Track %s: timestamps are NOT in ascending order", trackName)
			} else if len(track.Recs) > 10 {
				t.Logf("Track %s: timestamps are properly ordered (%d records)", trackName, len(track.Recs))
			}
		}
	}
}

// TestAllSampleFiles tests parsing of all available sample files.
func TestAllSampleFiles(t *testing.T) {
	sampleFiles := []string{
		"../data_sample/MICUA01_240724_180000.vital",
		"../data_sample/MICUB06_240322_230000.vital",
		"../data_sample/MICUA01_240724_190000.vital",
		"../data_sample/MICUB08_240520_230000.vital",
	}

	for _, filePath := range sampleFiles {
		t.Run(filePath, func(t *testing.T) {
			vf, err := NewVitalFile(filePath)
			if err != nil {
				t.Fatalf("parse error for %s: %v", filePath, err)
			}

			t.Logf("File: %s", filePath)
			t.Logf("  Devices: %d", len(vf.Devs))
			t.Logf("  Tracks: %d", len(vf.Trks))
			t.Logf("  Start time: %f", vf.DtStart)
			t.Logf("  End time: %f", vf.DtEnd)
			t.Logf("  DGMT: %d", vf.Dgmt)

			if len(vf.Trks) == 0 {
				t.Errorf("no tracks parsed in %s", filePath)
			}

			// 디바이스별 트랙 수 카운트
			deviceTrackCount := make(map[string]int)
			for trackName := range vf.Trks {
				for deviceName := range vf.Devs {
					if len(trackName) > len(deviceName) && trackName[:len(deviceName)] == deviceName {
						deviceTrackCount[deviceName]++
						break
					}
				}
			}

			for device, count := range deviceTrackCount {
				t.Logf("  Device %s: %d tracks", device, count)
			}
		})
	}
}

// TestGainOffsetAnalysis analyzes gain/offset values in NUMERIC tracks.
// This test provides statistical analysis of track configurations.
func TestGainOffsetAnalysis(t *testing.T) {
	sampleFiles := []string{
		"../data_sample/MICUA01_240724_180000.vital",
		"../data_sample/MICUB06_240322_230000.vital",
		"../data_sample/MICUA01_240724_190000.vital",
		"../data_sample/MICUB08_240520_230000.vital",
	}

	for _, filePath := range sampleFiles {
		t.Logf("\n=== 파일: %s ===", filePath)

		vf, err := NewVitalFile(filePath)
		if err != nil {
			t.Logf("파일 파싱 실패: %v", err)
			continue
		}

		gainOffsetCounts := make(map[string]int)
		integerTypesWithGainOffset := 0
		totalIntegerTypes := 0
		fmtCounts := make(map[uint8]int)

		for trackName, track := range vf.Trks {
			if track.Type == 2 { // NUMERIC only
				fmtCounts[track.Fmt]++

				isIntegerType := track.Fmt >= 3 && track.Fmt <= 8
				if isIntegerType {
					totalIntegerTypes++
					t.Logf("정수 타입 발견: %s (fmt=%d, gain=%f, offset=%f)",
						trackName, track.Fmt, track.Gain, track.Offset)
				}

				// Categorize gain/offset combinations
				key := ""
				if track.Gain == 1.0 && track.Offset == 0.0 {
					key = "identity" // gain=1, offset=0 (no conversion needed)
				} else if track.Gain != 1.0 && track.Offset == 0.0 {
					key = "scale_only" // only scaling
				} else if track.Gain == 1.0 && track.Offset != 0.0 {
					key = "offset_only" // only offset
				} else {
					key = "both" // both gain and offset
				}

				gainOffsetCounts[key]++

				if isIntegerType && (track.Gain != 1.0 || track.Offset != 0.0) {
					integerTypesWithGainOffset++
				}
			}
		}

		t.Logf("포맷 분포:")
		for fmt, count := range fmtCounts {
			switch fmt {
			case 1:
				t.Logf("  fmt=1 (float32): %d tracks", count)
			case 2:
				t.Logf("  fmt=2 (float64): %d tracks", count)
			case 3:
				t.Logf("  fmt=3 (int8): %d tracks", count)
			case 4:
				t.Logf("  fmt=4 (uint8): %d tracks", count)
			case 5:
				t.Logf("  fmt=5 (int16): %d tracks", count)
			case 6:
				t.Logf("  fmt=6 (uint16): %d tracks", count)
			case 7:
				t.Logf("  fmt=7 (int32): %d tracks", count)
			case 8:
				t.Logf("  fmt=8 (uint32): %d tracks", count)
			}
		}

		t.Logf("Gain/Offset 분포:")
		for category, count := range gainOffsetCounts {
			t.Logf("  %s: %d tracks", category, count)
		}

		if totalIntegerTypes > 0 {
			t.Logf("정수 타입 트랙: %d개 중 %d개가 gain/offset 적용 필요 (%.1f%%)",
				totalIntegerTypes, integerTypesWithGainOffset,
				float64(integerTypesWithGainOffset)/float64(totalIntegerTypes)*100)
		} else {
			t.Logf("정수 타입 NUMERIC 트랙이 없음 (모두 fmt=1 또는 fmt=2)")
		}
	}
}
