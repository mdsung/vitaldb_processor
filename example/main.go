package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mdsung/vitaldb_processor/vital"
)

type Config struct {
	Format      string  // "text" or "json"
	ListTracks  bool    // 트랙 목록만 출력
	InfoOnly    bool    // 파일 정보만 출력
	ListDevices bool    // 디바이스 목록만 출력
	Summary     bool    // 요약 정보만 출력
	Tracks      string  // 특정 트랙들만 출력 (쉼표로 구분)
	TrackType   string  // 트랙 타입 필터 ("WAVE", "NUMERIC", "STRING")
	MaxTracks   int     // 최대 트랙 개수 제한 (0 = 무제한)
	MaxSamples  int     // 샘플 데이터 최대 개수
	StartTime   float64 // 시작 시간
	EndTime     float64 // 종료 시간
	Quiet       bool    // 조용한 모드
	Verbose     bool    // 상세 모드
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
	Name        string       `json:"name"`
	Type        uint8        `json:"type"`
	TypeName    string       `json:"type_name"`
	Unit        string       `json:"unit"`
	SampleRate  float32      `json:"sample_rate"`
	Gain        float64      `json:"gain"`
	Offset      float64      `json:"offset"`
	MinDisplay  float32      `json:"min_display"`
	MaxDisplay  float32      `json:"max_display"`
	Color       uint32       `json:"color"`
	MonitorType uint8        `json:"monitor_type"`
	DeviceName  string       `json:"device_name"`
	Records     []RecordInfo `json:"records,omitempty"`
}

type RecordInfo struct {
	Time  float64     `json:"dt"`
	Value interface{} `json:"val"`
}

func main() {
	config := parseFlags()

	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <vital_file_path>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
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

	if config.Format == "json" {
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonData))
	} else {
		printTextOutput(output, config)
	}
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.Format, "format", "text", "출력 형식 (text, json)")
	flag.BoolVar(&config.ListTracks, "list-tracks", false, "트랙 목록만 출력")
	flag.BoolVar(&config.InfoOnly, "info-only", false, "파일 정보만 출력")
	flag.BoolVar(&config.ListDevices, "list-devices", false, "디바이스 목록만 출력")
	flag.BoolVar(&config.Summary, "summary", false, "요약 정보만 출력")
	flag.StringVar(&config.Tracks, "tracks", "", "특정 트랙들만 출력 (쉼표로 구분)")
	flag.StringVar(&config.TrackType, "track-type", "", "트랙 타입 필터 (WAVE, NUMERIC, STRING)")
	flag.IntVar(&config.MaxTracks, "max-tracks", 0, "최대 트랙 개수 제한 (0 = 무제한)")
	flag.IntVar(&config.MaxSamples, "max-samples", 3, "샘플 데이터 최대 개수")
	flag.Float64Var(&config.StartTime, "start-time", 0, "시작 시간")
	flag.Float64Var(&config.EndTime, "end-time", 0, "종료 시간 (0 = 파일 끝까지)")
	flag.BoolVar(&config.Quiet, "quiet", false, "조용한 모드 (에러만 출력)")
	flag.BoolVar(&config.Verbose, "verbose", false, "상세 모드")

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
			records := make([]RecordInfo, 0)
			for i, rec := range track.Recs {
				if i >= config.MaxSamples {
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
