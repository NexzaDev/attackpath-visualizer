package notifier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
)

type deliveryState struct {
	Delivered map[string]bool `json:"delivered"`
}

func statePath(databasePath string) (string, error) {
	home, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find notifier state directory: %w", err)
	}
	absolute, err := filepath.Abs(databasePath)
	if err != nil {
		return "", fmt.Errorf("resolve database path for notifier state: %w", err)
	}
	digest := sha256.Sum256([]byte(absolute))
	return filepath.Join(home, "AttackPathVisualizer", "notifier", hex.EncodeToString(digest[:])+".json"), nil
}

func loadDeliveryState(path string) (deliveryState, error) {
	state := deliveryState{Delivered: make(map[string]bool)}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return state, fmt.Errorf("read notifier delivery state: %w", err)
	}
	if err = json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("parse notifier delivery state: %w", err)
	}
	if state.Delivered == nil {
		state.Delivered = make(map[string]bool)
	}
	return state, nil
}

func saveDeliveryState(path string, state deliveryState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create notifier state directory: %w", err)
	}
	keys := make([]string, 0, len(state.Delivered))
	for key, delivered := range state.Delivered {
		if delivered {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for len(keys) > 2000 {
		delete(state.Delivered, keys[0])
		keys = keys[1:]
	}
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode notifier delivery state: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".notifier-state-*")
	if err != nil {
		return fmt.Errorf("create notifier state file: %w", err)
	}
	name := temporary.Name()
	if err = temporary.Chmod(0600); err == nil {
		_, err = temporary.Write(data)
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(name)
		return fmt.Errorf("write notifier state file: %w", err)
	}
	if err = os.Rename(name, path); err != nil {
		os.Remove(path)
		if err = os.Rename(name, path); err != nil {
			os.Remove(name)
			return fmt.Errorf("save notifier state file: %w", err)
		}
	}
	return nil
}

func notificationKey(item repository.Notification) string {
	return fmt.Sprintf("%020d:%s", item.ID, item.CreatedAt.UTC().Format(time.RFC3339Nano))
}
