package notifier

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NexzaDev/attackpath-visualizer/internal/notifications"
	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
)

func TestDeliverPendingUsesMinimalTextAndPersistsDeliveryOutsideDatabase(t *testing.T) {
	isolateDeliveryState(t)
	store, err := repository.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	if _, err = store.AddNotification(ctx, notifications.Event{Type: "critical_path", Severity: 5, Title: "New critical potential path", Detail: "Never shown in minimal text", Related: "paths"}, time.Minute); err != nil {
		t.Fatal(err)
	}
	var message string
	count, err := DeliverPending(ctx, store, func(_, text string) error {
		message = text
		return nil
	})
	if err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	if message != "A new critical potential path was found" {
		t.Fatalf("unexpected notification text %q", message)
	}
	pending, err := store.PendingNotifications(ctx)
	if err != nil || len(pending) != 1 {
		t.Fatalf("read-only delivery unexpectedly changed the database: pending=%v err=%v", pending, err)
	}
	secondCount, err := DeliverPending(ctx, store, func(string, string) error {
		t.Fatal("already delivered notification was sent twice")
		return nil
	})
	if err != nil || secondCount != 0 {
		t.Fatalf("second delivery count=%d err=%v", secondCount, err)
	}
}

func isolateDeliveryState(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", directory)
	t.Setenv("APPDATA", directory)
	t.Setenv("HOME", directory)
}

func TestDetailedNotificationPreferenceUsesConfiguredText(t *testing.T) {
	isolateDeliveryState(t)
	store, err := repository.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	if err = store.SaveNotificationRule(ctx, "validation_failed", notifications.Rule{Enabled: true, MinimumSeverity: 1, DetailedText: true}); err != nil {
		t.Fatal(err)
	}
	if _, err = store.AddNotification(ctx, notifications.Event{Type: "validation_failed", Severity: 3, Title: "Import failed", Detail: "The JSON field assets is missing.", Related: "imports"}, time.Minute); err != nil {
		t.Fatal(err)
	}
	var message string
	if _, err = DeliverPending(ctx, store, func(_, text string) error {
		message = text
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if message != "The JSON field assets is missing." {
		t.Fatalf("detailed notification text = %q", message)
	}
}

func TestDeliveryFailureDoesNotRecordDeliveredState(t *testing.T) {
	isolateDeliveryState(t)
	store, err := repository.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	if _, err = store.AddNotification(ctx, notifications.Event{Type: "validation_failed", Severity: 3, Title: "Import failed validation", Related: "imports"}, time.Minute); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("notification service unavailable")
	if _, err = DeliverPending(ctx, store, func(string, string) error { return failure }); !errors.Is(err, failure) {
		t.Fatalf("expected sender failure, got %v", err)
	}
	pending, err := store.PendingNotifications(ctx)
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending=%v err=%v", pending, err)
	}
}
