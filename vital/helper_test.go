package vital

// min returns the minimum of two integers.
// This helper function is used across multiple test files.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
