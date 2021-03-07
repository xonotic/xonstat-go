package models

import (
	"testing"
	"time"
)

func TestShortDurationString(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{"1m", "1m"},
		{"1m1s", "1m"},
		{"1h1m1s", "1h 1m"},
		{"1h1m", "1h 1m"},
		{"1h1s", "1h"},
		{"25h1m1s", "1d 1h 1m"},
		{"25h1m", "1d 1h 1m"},
	}

	for _, v := range tests {
		d, _ := time.ParseDuration(v.input)
		output := ShortDurationString(d)
		if output != v.output {
			t.Fatalf("expected %s, got %s", v.output, output)
		}
	}
}
