package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/mdsung/vitaldb_processor/vital"
	"github.com/parquet-go/parquet-go"
	"github.com/vmihailenco/msgpack/v5"
)

type Config struct {
	Format       string  // "csv", "parquet", "text", "json", or "msgpack"
	Compact      bool    // Compact JSON (no indentation)
	ListTracks   bool    // 트랙 목록만 출력
	InfoOnly     bool    // 파일 정보만 출력
	ListDevices  bool    // 디바이스 목록만 출력
	Summary      bool    // 요약 정보만 출력
	Tracks       string  // 특정 트랙들만 출력 (쉼표로 구분)
	TrackPattern string  // 트랙 이름 패턴 필터 (glob 스타일: *, ?)
	TrackType    string  // 트랙 타입 필터 ("WAVE", "NUMERIC", "STRING")
	MaxTracks    int     // 최대 트랙 개수 제한 (0 = 무제한)
	MaxSamples   int     // 샘플 데이터 최대 개수
	StartTime    float64 // 시작 시간
	EndTime      float64 // 종료 시간
	Quiet        bool    // 조용한 모드
	Verbose      bool    // 상세 모드
	CPUProfile   string  // CPU 프로파일 출력 파일
	MemProfile   string  // 메모리 프로파일 출력 파일
}

type OutputData struct {
	FileInfo *FileInfo             `json:"file_info,omitempty"`
	Devices  map[string]DeviceInfo `json:"devices,omitempty"`
	Tracks   map[string]TrackInfo  `json:"tracks,omitempty"`
}

type FileInfo struct {
	StartTime    float64 `json:"dt_start"`
	EndTime      float64 `json:"dt_end"`
	Duration     float64 `json:"duration"`
	GMTOffset    int16   `json:"gmt_offset"`
	TracksCount  int     `json:"tracks_count"`
	DevicesCount int     `json:"devices_count"`
}

type DeviceInfo struct {
	Name     string `json:"name"`
	TypeName string `json:"type_name"`
	Port     string `json:"port"`
}

type TrackInfo struct {
	Name         string       `json:"name"`
	Type         uint8        `json:"type"`
	TypeName     string       `json:"type_name"`
	Fmt          uint8        `json:"fmt"`
	Unit         string       `json:"unit"`
	SampleRate   float32      `json:"sample_rate"`
	Gain         float64      `json:"gain"`
	Offset       float64      `json:"offset"`
	MinDisplay   float32      `json:"min_display"`
	MaxDisplay   float32      `json:"max_display"`
	Color        uint32       `json:"color"`
	MonitorType  uint8        `json:"monitor_type"`
	DeviceName   string       `json:"device_name"`
	RecordsCount int          `json:"records_count"`
	Records      []RecordInfo `json:"records,omitempty"`
}

type RecordInfo struct {
	Time  float64     `json:"dt"`
	Value interface{} `json:"val"`
}

// ParquetRow represents a single row in Parquet output
type ParquetRow struct {
	TrackName string  `parquet:"track_name,snappy"`
	Timestamp float64 `parquet:"timestamp,snappy"`
	Value     string  `parquet:"value,snappy"`
	Unit      string  `parquet:"unit,snappy"`
}

func main() {
	config := parseFlags()

	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <vital_file_path>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// CPU 프로파일링 시작
	if config.CPUProfile != "" {
		f, err := os.Create(config.CPUProfile)
		if err != nil {
			log.Fatal("CPU 프로파일 생성 실패:", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("CPU 프로파일 시작 실패:", err)
		}
		defer pprof.StopCPUProfile()
	}

	filepath := flag.Args()[0]

	if !config.Quiet {
		fmt.Fprintf(os.Stderr, "Reading VitalDB file: %s\n", filepath)
	}

	vf, err := vital.NewVitalFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	output := processVitalFile(vf, config)

	// 출력 형식에 따라 처리
	switch config.Format {
	case "csv":
		if err := printCSVOutput(output, config); err != nil {
			log.Fatal(err)
		}
	case "parquet":
		if err := printParquetOutput(output, config); err != nil {
			log.Fatal(err)
		}
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		if !config.Compact {
			encoder.SetIndent("", "  ")
		}
		if err := encoder.Encode(output); err != nil {
			log.Fatal(err)
		}
	case "msgpack":
		// 버퍼링된 writer 사용 (syscall 오버헤드 감소)
		writer := bufio.NewWriterSize(os.Stdout, 256*1024) // 256KB 버퍼
		encoder := msgpack.NewEncoder(writer)
		if err := encoder.Encode(output); err != nil {
			log.Fatal(err)
		}
		if err := writer.Flush(); err != nil {
			log.Fatal(err)
		}
	case "text":
		printTextOutput(output, config)
	default:
		log.Fatalf("Unknown format: %s. Supported formats: csv, parquet, text, json, msgpack", config.Format)
	}

	// 메모리 프로파일링
	if config.MemProfile != "" {
		f, err := os.Create(config.MemProfile)
		if err != nil {
			log.Fatal("메모리 프로파일 생성 실패:", err)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("메모리 프로파일 작성 실패:", err)
		}
	}
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.Format, "format", "csv", "출력 형식 (csv, parquet, text, json, msgpack)")
	flag.BoolVar(&config.Compact, "compact", false, "Compact JSON (들여쓰기 없음)")
	flag.BoolVar(&config.ListTracks, "list-tracks", false, "트랙 목록만 출력")
	flag.BoolVar(&config.InfoOnly, "info-only", false, "파일 정보만 출력")
	flag.BoolVar(&config.ListDevices, "list-devices", false, "디바이스 목록만 출력")
	flag.BoolVar(&config.Summary, "summary", false, "요약 정보만 출력")
	flag.StringVar(&config.Tracks, "tracks", "", "특정 트랙들만 출력 (쉼표로 구분)")
	flag.StringVar(&config.TrackPattern, "track-pattern", "", "트랙 이름 패턴 필터 (glob 스타일: ECG*, *_II, 쉼표로 구분)")
	flag.StringVar(&config.TrackType, "track-type", "", "트랙 타입 필터 (WAVE, NUMERIC, STRING)")
	flag.IntVar(&config.MaxTracks, "max-tracks", 0, "최대 트랙 개수 제한 (0 = 무제한)")
	flag.IntVar(&config.MaxSamples, "max-samples", 3, "샘플 데이터 최대 개수")
	flag.Float64Var(&config.StartTime, "start-time", 0, "시작 시간")
	flag.Float64Var(&config.EndTime, "end-time", 0, "종료 시간 (0 = 파일 끝까지)")
	flag.BoolVar(&config.Quiet, "quiet", false, "조용한 모드 (에러만 출력)")
	flag.BoolVar(&config.Verbose, "verbose", false, "상세 모드")
	flag.StringVar(&config.CPUProfile, "cpuprofile", "", "CPU 프로파일 출력 파일")
	flag.StringVar(&config.MemProfile, "memprofile", "", "메모리 프로파일 출력 파일")

	flag.Parse()
	return config
}

func processVitalFile(vf *vital.VitalFile, config *Config) *OutputData {
	output := &OutputData{}

	// 파일 정보
	if !config.ListTracks && !config.ListDevices {
		output.FileInfo = &FileInfo{
			StartTime:    vf.DtStart,
			EndTime:      vf.DtEnd,
			Duration:     vf.DtEnd - vf.DtStart,
			GMTOffset:    vf.Dgmt,
			TracksCount:  len(vf.Trks),
			DevicesCount: len(vf.Devs),
		}
	}

	// 디바이스 정보
	if !config.InfoOnly && !config.ListTracks {
		output.Devices = make(map[string]DeviceInfo)
		for name, device := range vf.Devs {
			output.Devices[name] = DeviceInfo{
				Name:     device.Name,
				TypeName: device.TypeName,
				Port:     device.Port,
			}
		}
	}

	// 트랙 정보
	if !config.InfoOnly && !config.ListDevices {
		output.Tracks = processTracks(vf, config)
	}

	return output
}

// matchesAnyPattern checks if a name matches any of the provided glob patterns.
// Returns true if patterns is empty (no filtering) or if name matches at least one pattern.
// Supports wildcards (* and ?) that can match across path separators.
func matchesAnyPattern(name string, patterns []string) bool {
	// No patterns = no filtering, match everything
	if len(patterns) == 0 {
		return true
	}

	// Check if name matches any pattern
	for _, pattern := range patterns {
		// First try filepath.Match for exact path matching
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}

		// Also try matching against just the base name (after last /)
		// This allows patterns like "*_HR" to match "Bx50/ART1_HR"
		baseName := filepath.Base(name)
		matched, err = filepath.Match(pattern, baseName)
		if err == nil && matched {
			return true
		}

		// Try matching the pattern with simple string contains for wildcards
		// This handles cases like "*_HR" matching "Bx50/ART1_HR"
		if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
			// Convert glob pattern to regex-like matching
			if matchGlobPattern(name, pattern) {
				return true
			}
		}
	}

	return false
}

// matchGlobPattern performs simple glob matching that works across path separators
func matchGlobPattern(str, pattern string) bool {
	// Handle simple suffix patterns like "*_HR"
	if strings.HasPrefix(pattern, "*") && !strings.Contains(pattern[1:], "*") {
		suffix := pattern[1:]
		return strings.HasSuffix(str, suffix)
	}

	// Handle simple prefix patterns like "ECG*"
	if strings.HasSuffix(pattern, "*") && !strings.Contains(pattern[:len(pattern)-1], "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(str, prefix)
	}

	// Handle patterns with * in the middle
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			// Pattern like "Bx50/*_HR"
			return strings.HasPrefix(str, parts[0]) && strings.HasSuffix(str, parts[1])
		}
	}

	// Fall back to exact match
	return str == pattern
}

func processTracks(vf *vital.VitalFile, config *Config) map[string]TrackInfo {
	tracks := make(map[string]TrackInfo)

	// 트랙 필터링
	selectedTracks := make([]string, 0)
	if config.Tracks != "" {
		selectedTracks = strings.Split(config.Tracks, ",")
		for i := range selectedTracks {
			selectedTracks[i] = strings.TrimSpace(selectedTracks[i])
		}
	}

	// 패턴 필터링
	trackPatterns := make([]string, 0)
	if config.TrackPattern != "" {
		trackPatterns = strings.Split(config.TrackPattern, ",")
		for i := range trackPatterns {
			trackPatterns[i] = strings.TrimSpace(trackPatterns[i])
		}
	}

	count := 0
	for name, track := range vf.Trks {
		// 트랙 개수 제한 확인
		if config.MaxTracks > 0 && count >= config.MaxTracks {
			break
		}

		// 특정 트랙 필터링
		if len(selectedTracks) > 0 {
			found := false
			for _, selectedTrack := range selectedTracks {
				if selectedTrack == name {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 패턴 필터링
		if !matchesAnyPattern(name, trackPatterns) {
			continue
		}

		// 트랙 타입 필터링
		if config.TrackType != "" {
			typeMatch := false
			switch strings.ToUpper(config.TrackType) {
			case "WAVE":
				typeMatch = track.Type == 1
			case "NUMERIC":
				typeMatch = track.Type == 2
			case "STRING":
				typeMatch = track.Type == 5
			}
			if !typeMatch {
				continue
			}
		}

		trackInfo := TrackInfo{
			Name:        track.Name,
			Type:        track.Type,
			TypeName:    getTypeName(track.Type),
			Fmt:         track.Fmt,
			Unit:        track.Unit,
			SampleRate:  track.SRate,
			Gain:        track.Gain,
			Offset:      track.Offset,
			MinDisplay:  track.Mindisp,
			MaxDisplay:  track.Maxdisp,
			Color:       track.Col,
			MonitorType: track.Montype,
			DeviceName:  track.DName,
		}

		// 레코드 데이터 (요약 모드가 아닌 경우만)
		if !config.Summary && !config.ListTracks {
			// 메모리 프리할당: 최대 필요 용량 사전 확보
			expectedSize := len(track.Recs)
			if config.MaxSamples > 0 && config.MaxSamples < expectedSize {
				expectedSize = config.MaxSamples
			}
			records := make([]RecordInfo, 0, expectedSize)

			for i, rec := range track.Recs {
				if config.MaxSamples > 0 && i >= config.MaxSamples {
					break
				}

				// 시간 범위 필터링
				if config.StartTime > 0 && rec.Dt < config.StartTime {
					continue
				}
				if config.EndTime > 0 && rec.Dt > config.EndTime {
					break
				}

				records = append(records, RecordInfo{
					Time:  rec.Dt,
					Value: rec.Val,
				})
			}
			trackInfo.Records = records
			trackInfo.RecordsCount = len(records)
		} else {
			// Summary mode: still count total records
			trackInfo.RecordsCount = len(track.Recs)
		}

		tracks[name] = trackInfo
		count++
	}

	return tracks
}

func getTypeName(trackType uint8) string {
	switch trackType {
	case 1:
		return "WAVE"
	case 2:
		return "NUMERIC"
	case 5:
		return "STRING"
	default:
		return "UNKNOWN"
	}
}

func printParquetOutput(output *OutputData, config *Config) error {
	// Collect all rows first
	var rows []ParquetRow

	if output.Tracks != nil {
		for trackName, track := range output.Tracks {
			for _, record := range track.Records {
				// Convert value to string
				valueStr := fmt.Sprintf("%v", record.Value)

				rows = append(rows, ParquetRow{
					TrackName: trackName,
					Timestamp: record.Time,
					Value:     valueStr,
					Unit:      track.Unit,
				})
			}
		}
	}

	// Create Parquet writer with buffering
	writer := bufio.NewWriterSize(os.Stdout, 256*1024) // 256KB buffer
	defer writer.Flush()

	parquetWriter := parquet.NewGenericWriter[ParquetRow](writer)
	defer parquetWriter.Close()

	// Write all rows at once for better performance
	if _, err := parquetWriter.Write(rows); err != nil {
		return fmt.Errorf("failed to write Parquet data: %w", err)
	}

	return nil
}

func printCSVOutput(output *OutputData, config *Config) error {
	// CSV writer with buffering for performance
	writer := bufio.NewWriterSize(os.Stdout, 256*1024) // 256KB buffer
	csvWriter := csv.NewWriter(writer)
	defer func() {
		csvWriter.Flush()
		writer.Flush()
	}()

	// Write header
	header := []string{"track_name", "timestamp", "value", "unit"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	if output.Tracks != nil {
		for trackName, track := range output.Tracks {
			for _, record := range track.Records {
				// Convert value to string
				valueStr := fmt.Sprintf("%v", record.Value)

				row := []string{
					trackName,
					fmt.Sprintf("%.6f", record.Time),
					valueStr,
					track.Unit,
				}

				if err := csvWriter.Write(row); err != nil {
					return fmt.Errorf("failed to write CSV row: %w", err)
				}
			}
		}
	}

	return nil
}

func printTextOutput(output *OutputData, config *Config) {
	if output.FileInfo != nil {
		fmt.Printf("=== File Information ===\n")
		fmt.Printf("Start Time: %f\n", output.FileInfo.StartTime)
		fmt.Printf("End Time: %f\n", output.FileInfo.EndTime)
		fmt.Printf("Duration: %.2f seconds\n", output.FileInfo.Duration)
		fmt.Printf("GMT Offset: %d\n", output.FileInfo.GMTOffset)
		fmt.Printf("Number of Tracks: %d\n", output.FileInfo.TracksCount)
		fmt.Printf("Number of Devices: %d\n", output.FileInfo.DevicesCount)
		fmt.Println()
	}

	if output.Devices != nil && len(output.Devices) > 0 {
		fmt.Printf("=== Devices ===\n")
		for name, device := range output.Devices {
			fmt.Printf("- %s: %s (Port: %s)\n", name, device.TypeName, device.Port)
		}
		fmt.Println()
	}

	if output.Tracks != nil && len(output.Tracks) > 0 {
		if config.ListTracks {
			fmt.Printf("=== Available Tracks ===\n")
			for name, track := range output.Tracks {
				fmt.Printf("- %s: %s (%s), Rate: %.1f Hz\n",
					name, track.TypeName, track.Unit, track.SampleRate)
			}
		} else {
			fmt.Printf("=== Tracks ===\n")
			for name, track := range output.Tracks {
				fmt.Printf("- %s: %s, Rate: %.1f Hz, Records: %d\n",
					name, track.Unit, track.SampleRate, len(track.Records))

				if config.Verbose && len(track.Records) > 0 {
					fmt.Printf("  Sample data:\n")
					for i, rec := range track.Records {
						if i >= 3 {
							break
						}
						fmt.Printf("    [%d] Time: %.6f, Value: %v\n", i+1, rec.Time, rec.Value)
					}
				}
			}
		}
	}
}
