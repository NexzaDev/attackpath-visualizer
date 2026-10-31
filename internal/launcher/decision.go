package launcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Options struct {
	UI       string
	EnvUI    string
	Saved    string
	Ask      bool
	NoPrompt bool
	TTY      bool
	Headless bool
}

type Decision struct {
	Mode    string
	Prompt  bool
	Chooser bool
	Message string
}

func Decide(options Options) (Decision, error) {
	mode := strings.ToLower(options.UI)
	if mode == "" {
		mode = strings.ToLower(options.EnvUI)
	}
	if mode != "" && mode != "gui" && mode != "tui" {
		return Decision{}, fmt.Errorf("UI must be gui or tui, got %q", mode)
	}
	if mode != "" && !options.Ask {
		return Decision{Mode: mode}, nil
	}
	if options.Ask {
		if options.Headless && !options.TTY {
			return Decision{Mode: "tui", Message: "No display server or interactive terminal detected; opening the terminal interface."}, nil
		}
		if options.TTY {
			return Decision{Mode: preferred(options.Saved), Prompt: true}, nil
		}
		return Decision{Mode: preferred(options.Saved), Chooser: true}, nil
	}
	if options.Headless {
		return Decision{Mode: "tui", Message: "No display server detected; opening the terminal interface."}, nil
	}
	if options.NoPrompt {
		return Decision{Mode: preferred(options.Saved)}, nil
	}
	if options.TTY {
		return Decision{Mode: preferred(options.Saved), Prompt: true}, nil
	}
	return Decision{Mode: preferred(options.Saved), Chooser: true}, nil
}

func preferred(saved string) string {
	if saved == "gui" || saved == "tui" {
		return saved
	}
	return "gui"
}

func ConfigPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find configuration directory: %w", err)
	}
	return filepath.Join(base, "AttackPathVisualizer", "launcher.json"), nil
}

func ReadPreference() string {
	path, err := ConfigPath()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var preference struct {
		Mode string `json:"mode"`
	}
	if json.Unmarshal(data, &preference) != nil || (preference.Mode != "gui" && preference.Mode != "tui") {
		return ""
	}
	return preference.Mode
}

func SavePreference(mode string) error {
	if mode != "gui" && mode != "tui" && mode != "" {
		return errors.New("saved UI must be gui, tui, or empty")
	}
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create preferences directory: %w", err)
	}
	if mode == "" {
		if err = os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("clear launcher preference: %w", err)
		}
		return nil
	}
	data, err := json.Marshal(struct {
		Mode string `json:"mode"`
	}{Mode: mode})
	if err != nil {
		return err
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("save launcher preference: %w", err)
	}
	if err = os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("restrict launcher preference permissions: %w", err)
	}
	return nil
}

func Headless() bool {
	if runtime.GOOS == "windows" {
		return false
	}
	if runtime.GOOS == "darwin" {
		return false
	}
	return os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == ""
}

func InteractiveStdin() bool {
	file, err := os.Stdin.Stat()
	return err == nil && file.Mode()&os.ModeCharDevice != 0
}
