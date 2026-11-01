package notifier

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/NexzaDev/attackpath-visualizer/internal/config"
	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
	"github.com/gen2brain/beeep"
)

const DefaultInterval = 15 * time.Minute

type Sender func(string, string) error

func Run(ctx context.Context, requestedDB string, interval time.Duration, sender Sender) error {
	if interval <= 0 || interval > time.Hour {
		return errors.New("notifier interval must be between one nanosecond and one hour")
	}
	path, err := config.DatabasePath(requestedDB)
	if err != nil {
		return err
	}
	store, err := repository.OpenReadOnly(path)
	if err != nil {
		return err
	}
	defer store.Close()
	if sender == nil {
		sender = func(title, message string) error { return beeep.Notify(title, message, nil) }
	}
	for {
		if _, err = DeliverPending(ctx, store, sender); err != nil {
			return err
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func RunUntilSignal(requestedDB string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	err := Run(ctx, requestedDB, DefaultInterval, nil)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func DeliverPending(ctx context.Context, store *repository.Store, sender Sender) (int, error) {
	items, err := store.PendingNotifications(ctx)
	if err != nil {
		return 0, err
	}
	rules, err := store.NotificationRules(ctx)
	if err != nil {
		return 0, err
	}
	path, err := statePath(store.Path())
	if err != nil {
		return 0, err
	}
	state, err := loadDeliveryState(path)
	if err != nil {
		return 0, err
	}
	delivered := 0
	for _, item := range items {
		key := notificationKey(item)
		if state.Delivered[key] {
			continue
		}
		title, message := "AttackPath Visualizer", "New local findings are available"
		if item.Type == "validation_failed" {
			message = "An import failed validation"
		} else {
			message = "A new critical potential path was found"
		}
		if rule, ok := rules[item.Type]; ok && rule.DetailedText && item.Detail != "" {
			message = item.Detail
		}
		if err := sender(title, message); err != nil {
			return delivered, fmt.Errorf("send operating-system notification: %w", err)
		}
		state.Delivered[key] = true
		if err := saveDeliveryState(path, state); err != nil {
			return delivered, err
		}
		delivered++
	}
	return delivered, nil
}
