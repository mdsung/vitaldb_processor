package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mdsung/vitaldb_processor/vital"
)

func benchmarkGoVitalDB(filePaths []string) map[string]interface{} {
	startTime := time.Now()

	totalTracks := 0
	filesProcessed := 0

	for _, filePath := range filePaths {
		// Load the file
		vf, err := vital.NewVitalFile(filePath)
		if err != nil {
			log.Printf("Error loading %s: %v", filePath, err)
			continue
		}

		// Count tracks
		totalTracks += len(vf.Trks)
		filesProcessed++
	}

	elapsed := time.Since(startTime)

	return map[string]interface{}{
		"elapsed": elapsed.Seconds(),
		"files":   filesProcessed,
		"tracks":  totalTracks,
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

	// Sort for consistency
	sort.Strings(filePaths)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Go VitalDB Processor Performance Benchmark (Simple)")
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

	// Warm-up run
	fmt.Println("Warm-up run...")
	_ = benchmarkGoVitalDB(filePaths[:1])

	// Actual benchmark (run 3 times and take average)
	fmt.Println("Running benchmark (3 iterations)...")
	var times []float64
	var lastResult map[string]interface{}

	for i := 0; i < 3; i++ {
		result := benchmarkGoVitalDB(filePaths)
		elapsed := result["elapsed"].(float64)
		times = append(times, elapsed)
		fmt.Printf("  Iteration %d: %.3f seconds\n", i+1, elapsed)
		lastResult = result
	}

	// Calculate average
	var sum float64
	for _, t := range times {
		sum += t
	}
	avgTime := sum / float64(len(times))

	fmt.Println()
	fmt.Println("Results:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Files processed:    %d\n", lastResult["files"])
	fmt.Printf("Total tracks:       %d\n", lastResult["tracks"])
	fmt.Printf("Average time:       %.3f seconds\n", avgTime)
	throughput := float64(totalSize) / 1024 / 1024 / avgTime
	fmt.Printf("Throughput:         %.2f MB/s\n", throughput)
	fmt.Printf("Files per second:   %.2f\n", float64(lastResult["files"].(int))/avgTime)
	fmt.Println()
}
