package repository

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/NexzaDev/attackpath-visualizer/internal/notifications"
)

func TestOpenReadOnlyAllowsReadsAndRejectsWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "graph.db")
	writer, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err = writer.AddNotification(ctx, notifications.Event{Type: "critical_path", Severity: 5, Title: "Critical", Related: "paths"}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := OpenReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if reader.Path() != path {
		t.Fatalf("read-only store path = %q", reader.Path())
	}
	items, err := reader.Notifications(ctx, 10)
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%v err=%v", items, err)
	}
	if err = reader.MarkNotificationDelivered(ctx, items[0].ID); err == nil {
		t.Fatal("read-only database accepted a write")
	}
}
