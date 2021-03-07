package models

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/antzucaro/qstr"
	"github.com/nleeper/goment"
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

// ShortDurationString returns a "short" form of a duration string. Components include
// days, hours, and minutes (no seconds).
func ShortDurationString(d time.Duration) string {
	// The smallest grain is minutes, so let's get the total number of those.
	// As we take out chunks for the larger grained items, this gets decremented.
	minutes := uint64(d.Minutes())

	days := uint64(minutes / 1440)
	minutes -= days * 1440

	hours := uint64(minutes / 60)
	minutes -= hours * 60

	var buffer bytes.Buffer
	if days > 0 {
		buffer.WriteString(fmt.Sprintf("%dd ", days))
	}

	if hours > 0 {
		buffer.WriteString(fmt.Sprintf("%dh ", hours))
	}

	if minutes > 0 {
		buffer.WriteString(fmt.Sprintf("%dm ", minutes))
	}

	return strings.TrimRight(buffer.String(), " ")
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
		if short {
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

// MultiDt takes a normal time.Time object and converts it into several commonly-used values.
type MultiDt struct {
	Dt     time.Time
	Epoch  int64
	UTCStr string
	Fuzzy  string
}

// NewMultiDt takes a time.Time and computes several commonly-used formats for it.
func NewMultiDt(dt time.Time) (*MultiDt, error) {
	dtUTC := dt.UTC()

	fuzzyDt, err := goment.New(dtUTC)
	if err != nil {
		return nil, err
	}

	epoch := dt.Unix()
	dtUTCStr := dtUTC.Format("Mon, 2 Jan 2006 15:04:05 MST")
	fuzzy := fuzzyDt.FromNow()

	return &MultiDt{
		Dt:     dt,
		Epoch:  epoch,
		UTCStr: dtUTCStr,
		Fuzzy:  fuzzy,
	}, nil
}

// MultiDuration is like MultiDt, but for time.Duration-s.
type MultiDuration struct {
	Duration     time.Duration
	Seconds      float64
	Milliseconds int64
	Short        string
	Long         string
}

// NewMultiDuration creates a new MultiDuration.
func NewMultiDuration(d time.Duration) *MultiDuration {
	return &MultiDuration{
		Duration:     d,
		Milliseconds: d.Milliseconds(),
		Seconds:      d.Seconds(),
		Short:        DurationString(d, true),
		Long:         DurationString(d, false),
	}
}

// MultiNick provides some common formats for a player's nick.
type MultiNick struct {
	Nick         string
	NickStripped string
	NickHTML     template.HTML
}

// NewMultiNick creates a MultiNick
func NewMultiNick(nick string) *MultiNick {
	n := qstr.QStr(nick)
	return &MultiNick{
		Nick:         nick,
		NickStripped: n.Stripped(),
		NickHTML:     n.HTML(),
	}
}
