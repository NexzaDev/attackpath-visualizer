package notifications

import (
	"testing"
	"time"
)

func TestAllowedSeverityAndQuietHours(t *testing.T) {
	now := time.Date(2026, 10, 31, 23, 0, 0, 0, time.UTC)
	rule := Rule{Enabled: true, MinimumSeverity: 3, QuietStart: "22:00", QuietEnd: "07:00"}
	if Allowed(Event{Severity: 5}, rule, now) {
		t.Fatal("quiet hours must suppress notification")
	}
	now = time.Date(2026, 10, 31, 8, 0, 0, 0, time.UTC)
	if !Allowed(Event{Severity: 5}, rule, now) {
		t.Fatal("eligible event should pass")
	}
	if Allowed(Event{Severity: 2}, rule, now) {
		t.Fatal("low severity should be filtered")
	}
}

func TestDeduplicationWindow(t *testing.T) {
	now := time.Now()
	if !Deduplicated(now.Add(-time.Minute), now, 10*time.Minute) {
		t.Fatal("recent duplicate should be suppressed")
	}
	if Deduplicated(now.Add(-time.Hour), now, 10*time.Minute) {
		t.Fatal("old event should not be suppressed")
	}
}

func TestKeyIsStableAndDistinct(t *testing.T) {
	first := Key(Event{Type: "critical_path", Related: "asset-a", Title: "Critical"})
	if first != Key(Event{Type: "critical_path", Related: "asset-a", Title: "Critical"}) {
		t.Fatal("same event must have same key")
	}
	if first == Key(Event{Type: "validation_failed", Related: "asset-a", Title: "Critical"}) {
		t.Fatal("different events must differ")
	}
}
