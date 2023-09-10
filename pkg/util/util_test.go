package util

import (
	"testing"
	"time"
)

func TestIsLeapYear(t *testing.T) {
	leapYears := []int{
		2000, 2004, 2008, 2012, 2016, 2020, 2024, 2028, 2032, 2036, 2040, 2044, 2048,
	}

	for _, year := range leapYears {
		if !IsLeapYear(year) {
			t.Fatalf("expected %d to be marked as a leapyear", year)
		}
	}
}

func TestIsNotLeapYear(t *testing.T) {
	nonLeapYears := []int{
		2001, 2005, 2007, 2013, 2017, 2021, 2025, 2029, 2033, 2037, 2041, 2045, 2049,
	}

	for _, year := range nonLeapYears {
		if IsLeapYear(year) {
			t.Fatalf("expected %d to NOT be marked as a leapyear", year)
		}
	}
}

func TestIsAnniversary(t *testing.T) {
	layout := "2006-01-02" // Input is in YYYY-MM-DD format

	tests := []struct {
		t1       string
		t2       string
		expected bool
	}{
		{"2000-07-29", "2004-07-29", true}, // happy path, no leap year case needed
		{"2000-02-29", "2004-02-29", true}, // same day, two different leap years
		{"2000-02-29", "2005-03-01", true}, // should follow March 1 logic since both are not leap years
		{"2000-02-29", "2004-03-01", false}, // should not follow March 1 logic since both are leap years
	}

	for _, test := range tests {
		t1, _ := time.Parse(layout, test.t1)
		t2, _ := time.Parse(layout, test.t2)
		if IsAnniversary(t1, t2) != test.expected {
			t.Fatalf("expected %s and %s to be %v", test.t1, test.t2, test.expected)
		}
	}
}
