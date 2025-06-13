package vital

import (
	"fmt"
)

// ParseVitalFile parses a VitalDB file and returns the structured data
func ParseVitalFile(filePath string) (*VitalFile, error) {
	// Parse using existing NewVitalFile function which accepts file path
	vitalFile, err := NewVitalFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vital file: %w", err)
	}

	return vitalFile, nil
}
