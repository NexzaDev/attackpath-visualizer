package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func DatabasePath(requested string) (string, error) {
	if requested != "" {
		return requested, nil
	}
	if configured := os.Getenv("ATTACKPATH_VISUALIZER_DB"); configured != "" {
		return configured, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find user configuration directory: %w", err)
	}
	return filepath.Join(base, "AttackPathVisualizer", "graph.db"), nil
}
