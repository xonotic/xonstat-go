package models

import (
	"testing"
	"time"
)

func TestDurationString(t *testing.T) {
	tests := []struct {
		input  string
		shortOutput string
		longOutput string
	}{
		{"1m", "1m", "1 min"},
		{"1m1s", "1m", "1 min"},
		{"1h1m1s", "1h 1m", "1 hour, 1 min"},
		{"1h1m", "1h 1m", "1 hour, 1 min"},
		{"1h1s", "1h", "1 hour"},
		{"25h1m1s", "1d 1h 1m", "1 day, 1 hour, 1 min"},
		{"25h1m", "1d 1h 1m", "1 day, 1 hour, 1 min"},
	}

	for _, v := range tests {
		d, _ := time.ParseDuration(v.input)
		shortOutput := DurationString(d, "short")
		if shortOutput != v.shortOutput {
			t.Fatalf("wrong short format expected %s, got %s", v.shortOutput, shortOutput)
		}
		longOutput := DurationString(d, "long")
		if longOutput != v.longOutput {
			t.Fatalf("wrong long format expected %s, got %s", v.longOutput, longOutput)
		}
	}
}
