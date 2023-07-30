package util

import "time"
import "fmt"

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

func IsLeapYear(year int) bool {
	if year % 400 == 0 || (year % 4 == 0 && year % 100 != 0) {
		return true
	} else {
		return false
	}
}

// IsAnniversary determines if two dates occurred on the same day in different years.
// t1 is considered the "reference" time. 
func IsAnniversary(t1, t2 time.Time) bool {
	t1Year, t1Month, t1Day := t1.Date()
	t2Year, t2Month, t2Day := t2.Date()

	fmt.Printf("t1: %d-%d-%d", t1Year, t1Month, t1Day)
	fmt.Printf("t2: %d-%d-%d", t2Year, t2Month, t2Day)

	if t1Month == t2Month && t1Day == t2Day && t1Year < t2Year {
		return true
	}

	// Leap day folks (those who have a reference time on a leap day) get their cakes on March 1st,
	// but only in non-leap years! On leap years they get it on the actual day.
	if t1Month == 2 && t1Day == 29 && !IsLeapYear(t2Year) {
		if t2Month == 3 && t2Day == 1 {
			return true
		}
	}

	return false
}