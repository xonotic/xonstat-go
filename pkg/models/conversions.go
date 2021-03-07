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

// DurationString a formatted duration string. Formats include "short" and "long".
func DurationString(d time.Duration, format string) string {
	// The smallest grain is minutes, so let's get the total number of those.
	// As we take out chunks for the larger grained items, this gets decremented.
	minutes := uint64(d.Minutes())

	days := uint64(minutes / 1440)
	minutes -= days * 1440

	hours := uint64(minutes / 60)
	minutes -= hours * 60

	var separator, prefix, daySuffix, hourSuffix, minSuffix string
	if format == "long" {
		separator = ", "
		prefix = " "
		daySuffix = "day"
		hourSuffix = "hour"
		minSuffix = "min"
	} else {
		separator = " "
		prefix = ""
		daySuffix = "d"
		hourSuffix = "h"
		minSuffix = "m"
	}

	var buffer bytes.Buffer
	if days > 0 {
		buffer.WriteString(fmt.Sprintf("%d%s%s", days, prefix, daySuffix))
		if days > 1 && format == "long" {
			buffer.WriteString("s")
		}
		buffer.WriteString(separator)
	}

	if hours > 0 {
		buffer.WriteString(fmt.Sprintf("%d%s%s", hours, prefix, hourSuffix))
		if hours > 1 && format == "long" {
			buffer.WriteString("s")
		}
		buffer.WriteString(separator)

	}

	if minutes > 0 {
		buffer.WriteString(fmt.Sprintf("%d%s%s", minutes, prefix, minSuffix))
		if minutes > 1 && format == "long" {
			buffer.WriteString("s")
		}
		buffer.WriteString(separator)
	}

	return strings.TrimRight(buffer.String(), separator)
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
		Short:        DurationString(d, "short"),
		Long:         DurationString(d, "long"),
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
