package models

import (
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
