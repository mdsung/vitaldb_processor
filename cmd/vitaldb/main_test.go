package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mdsung/vitaldb_processor/vital"
)

func TestMatchesAnyPattern(t *testing.T) {
	tests := []struct {
		name     string
		trackName string
		patterns []string
		expected bool
	}{
		{
			name:      "Empty patterns returns true",
			trackName: "ECG_II",
			patterns:  []string{},
			expected:  true,
		},
		{
			name:      "Nil patterns returns true",
			trackName: "ECG_II",
			patterns:  nil,
			expected:  true,
		},
		{
			name:      "Exact match",
			trackName: "ECG_II",
			patterns:  []string{"ECG_II"},
			expected:  true,
		},
		{
			name:      "Wildcard prefix match",
			trackName: "ECG_II",
			patterns:  []string{"ECG*"},
			expected:  true,
		},
		{
			name:      "Wildcard suffix match",
			trackName: "Bx50/ART1_HR",
			patterns:  []string{"*_HR"},
			expected:  true,
		},
		{
			name:      "Wildcard middle match",
			trackName: "Bx50/ART1_HR",
			patterns:  []string{"Bx50/*_HR"},
			expected:  true,
		},
		{
			name:      "Full wildcard match",
			trackName: "Bx50/ART1_HR",
			patterns:  []string{"Bx50/ART*"},
			expected:  true,
		},
		{
			name:      "Multiple patterns - first matches",
			trackName: "ECG_II",
			patterns:  []string{"ECG*", "HR*"},
			expected:  true,
		},
		{
			name:      "Multiple patterns - second matches",
			trackName: "HR",
			patterns:  []string{"ECG*", "HR*"},
			expected:  true,
		},
		{
			name:      "No match",
			trackName: "HR",
			patterns:  []string{"ECG*"},
			expected:  false,
		},
		{
			name:      "Question mark wildcard",
			trackName: "ECG_I",
			patterns:  []string{"ECG_?"},
			expected:  true,
		},
		{
			name:      "Complex pattern with device prefix",
			trackName: "Bx50/PLETH_HR",
			patterns:  []string{"Bx50/*ETH*"},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesAnyPattern(tt.trackName, tt.patterns)
			if result != tt.expected {
				t.Errorf("matchesAnyPattern(%q, %v) = %v, expected %v",
					tt.trackName, tt.patterns, result, tt.expected)
			}
		})
	}
}

func TestTrackPatternFiltering(t *testing.T) {
	// Find test data file
	testFile := filepath.Join("..", "data_sample", "MICUA01_240724_180000.vital")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test data file not found, skipping integration test")
	}

	tests := []struct {
		name          string
		trackPattern  string
		minExpected   int
		maxExpected   int
		shouldContain []string
	}{
		{
			name:          "ART pattern matches ART tracks",
			trackPattern:  "Bx50/ART*",
			minExpected:   4,
			maxExpected:   10,
			shouldContain: []string{"Bx50/ART1_HR", "Bx50/ART1_DBP", "Bx50/ART1_MBP", "Bx50/ART1_SBP"},
		},
		{
			name:          "HR pattern matches all HR tracks",
			trackPattern:  "Bx50/*_HR",
			minExpected:   2,
			maxExpected:   5,
			shouldContain: []string{"Bx50/ART1_HR", "Bx50/PLETH_HR"},
		},
		{
			name:         "Multiple patterns with OR logic",
			trackPattern: "Bx50/ART1_HR,Bx50/PLETH_HR",
			minExpected:  2,
			maxExpected:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a config with the pattern
			config := &Config{
				TrackPattern: tt.trackPattern,
				Quiet:        true,
				MaxSamples:   0, // Don't load samples for faster test
			}

			// Load the vital file
			vf, err := vital.NewVitalFile(testFile)
			if err != nil {
				t.Fatalf("Failed to load vital file: %v", err)
			}

			// Process tracks with pattern filtering
			tracks := processTracks(vf, config)

			// Check track count
			trackCount := len(tracks)
			if trackCount < tt.minExpected || trackCount > tt.maxExpected {
				t.Errorf("Expected %d-%d tracks, got %d tracks",
					tt.minExpected, tt.maxExpected, trackCount)
			}

			// Check specific tracks are present
			for _, trackName := range tt.shouldContain {
				if _, found := tracks[trackName]; !found {
					t.Errorf("Expected track %q not found in results", trackName)
				}
			}
		})
	}
}

