package models

import (
	"bytes"
	"fmt"
	"time"
)

// durationToMSStr converts a pointer to a time.Duration into a string suitable for an INSERT.
// The pq library doesn't support time.Duration -> interval PG type, so we have to convert it 
// to a string. We'll do this with millisecond granularity, allowing fractional pieces too.
func durationToMSStr(t *time.Duration) string {
	durationLiteral := "NULL"
	if t != nil {
		durationLiteral = fmt.Sprintf("'%d milliseconds'", t.Milliseconds())
	}

	return durationLiteral
}

// DurationString creates a human-readable duration string with a days component.
func DurationString(d time.Duration, short bool) string {
	minutes := uint64(d.Minutes())
	days := uint64(minutes / 1440)
	minutes -= days * 1440
	hours := uint64(minutes / 60)
	minutes -= hours * 60

	var buffer bytes.Buffer
	if days == 1 {
		if short {
			buffer.WriteString("1d")
		} else {
			buffer.WriteString("1 day")
		}
	} else if days > 1 {
		if short {
			buffer.WriteString(fmt.Sprintf("%dd", days))
		} else {
			buffer.WriteString(fmt.Sprintf("%d days", days))
		}
	}

	if hours >= 1 && days >= 1 {
		if !short {
			buffer.WriteString(", ")
		}
	}

	if hours == 1 {
		if short {
			buffer.WriteString("1h")
		} else {
			buffer.WriteString("1 hr")
		}
	} else if hours > 1 {
		if short{
			buffer.WriteString(fmt.Sprintf("%dh", hours))
		} else {
			buffer.WriteString(fmt.Sprintf("%d hrs", hours))
		}
	}

	if minutes >= 1 && hours >= 1 {
		if !short {
			buffer.WriteString(", ")
		}
	}

	if minutes == 1 {
		if short {
			buffer.WriteString("1m")
		} else {
			buffer.WriteString("1 min")
		}
	} else if minutes > 1 {
		if short {
			buffer.WriteString(fmt.Sprintf("%dm", minutes))
		} else {
			buffer.WriteString(fmt.Sprintf("%d mins", minutes))
		}
	}
	return buffer.String()
}
