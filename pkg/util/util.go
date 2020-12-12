package util

// Percentage calculates the percentage of the given numbers.
func Percentage(numerator, denominator int) float32 {
	if denominator == 0 {
		denominator = 1
	}

	return float32(numerator) / float32(denominator) * 100.0
}