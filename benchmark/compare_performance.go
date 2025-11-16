package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mdsung/vitaldb_processor/vital"
)

func benchmarkGoVitalDB(filePaths []string) map[string]interface{} {
	startTime := time.Now()

	totalTracks := 0
	totalRecords := 0

	for _, filePath := range filePaths {
		// Load the file
		vf, err := vital.NewVitalFile(filePath)
		if err != nil {
			log.Printf("Error loading %s: %v", filePath, err)
			continue
		}

		// Count tracks and records
		totalTracks += len(vf.Trks)

		for _, track := range vf.Trks {
			totalRecords += len(track.Recs)
		}
	}

	elapsed := time.Since(startTime)

	return map[string]interface{}{
		"elapsed": elapsed.Seconds(),
		"files":   len(filePaths),
		"tracks":  totalTracks,
		"records": totalRecords,
	}
}

func main() {
	// Find all .vital files
	dataDir := "../data_sample"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		log.Fatalf("ERROR: %s directory not found", dataDir)
	}

	pattern := filepath.Join(dataDir, "*.vital")
	filePaths, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatalf("ERROR: Failed to glob files: %v", err)
	}

	if len(filePaths) == 0 {
		log.Fatalf("ERROR: No .vital files found in %s", dataDir)
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Go VitalDB Processor Performance Benchmark")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Files to process: %d\n", len(filePaths))

	// Calculate total size
	var totalSize int64
	for _, path := range filePaths {
		info, err := os.Stat(path)
		if err == nil {
			totalSize += info.Size()
		}
	}
	fmt.Printf("Total size: %.2f MB\n", float64(totalSize)/1024/1024)
	fmt.Println()

	// Run benchmark
	fmt.Println("Running benchmark...")
	result := benchmarkGoVitalDB(filePaths)

	fmt.Println()
	fmt.Println("Results:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Files processed:    %d\n", result["files"])
	fmt.Printf("Total tracks:       %d\n", result["tracks"])
	fmt.Printf("Total records:      %d\n", result["records"])
	fmt.Printf("Time elapsed:       %.3f seconds\n", result["elapsed"])
	throughput := float64(totalSize) / 1024 / 1024 / result["elapsed"].(float64)
	fmt.Printf("Throughput:         %.2f MB/s\n", throughput)
	fmt.Printf("Files per second:   %.2f\n", float64(result["files"].(int))/result["elapsed"].(float64))
	fmt.Println()
}
