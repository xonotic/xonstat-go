package util

// Ratio calculates the ratio between two numbers, handling divide-by-zero errors.
func Ratio(numerator, denominator int) float32 {
	if denominator == 0 {
		denominator = 1
	}

	return float32(numerator) / float32(denominator)
}

// Percentage calculates the percentage of the given numbers.
func Percentage(numerator, denominator int) float32 {
	return Ratio(numerator, denominator) * 100.0
}
