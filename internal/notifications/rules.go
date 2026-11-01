package notifications

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Event struct {
	Type     string
	Severity int
	Title    string
	Detail   string
	Related  string
	Key      string
}

type Rule struct {
	Enabled         bool
	MinimumSeverity int
	QuietStart      string
	QuietEnd        string
	DetailedText    bool
}

func Key(event Event) string {
	if event.Key != "" {
		return event.Key
	}
	sum := sha256.Sum256([]byte(event.Type + "\x00" + event.Related + "\x00" + event.Title))
	return hex.EncodeToString(sum[:])
}

func Allowed(event Event, rule Rule, now time.Time) bool {
	if !rule.Enabled || event.Severity < rule.MinimumSeverity {
		return false
	}
	return !Quiet(now, rule.QuietStart, rule.QuietEnd)
}

func Quiet(now time.Time, start, end string) bool {
	startAt, startErr := time.Parse("15:04", start)
	endAt, endErr := time.Parse("15:04", end)
	if startErr != nil || endErr != nil || start == end {
		return false
	}
	current := now.Hour()*60 + now.Minute()
	startMinute := startAt.Hour()*60 + startAt.Minute()
	endMinute := endAt.Hour()*60 + endAt.Minute()
	if startMinute < endMinute {
		return current >= startMinute && current < endMinute
	}
	return current >= startMinute || current < endMinute
}

func Deduplicated(last, now time.Time, window time.Duration) bool {
	return !last.IsZero() && now.Sub(last) >= 0 && now.Sub(last) < window
}
