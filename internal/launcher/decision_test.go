package launcher

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDecideTable(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		want    Decision
	}{
		{name: "explicit flag", options: Options{UI: "tui", TTY: true}, want: Decision{Mode: "tui"}},
		{name: "environment selection", options: Options{EnvUI: "gui", TTY: true}, want: Decision{Mode: "gui"}},
		{name: "saved terminal preference prompt", options: Options{Saved: "tui", TTY: true}, want: Decision{Mode: "tui", Prompt: true}},
		{name: "desktop chooser", options: Options{Saved: "gui"}, want: Decision{Mode: "gui", Chooser: true}},
		{name: "headless terminal", options: Options{Headless: true}, want: Decision{Mode: "tui", Message: "No display server detected; opening the terminal interface."}},
		{name: "forced prompt", options: Options{UI: "gui", Ask: true, TTY: true}, want: Decision{Mode: "gui", Prompt: true}},
		{name: "ask without display or terminal", options: Options{Ask: true, Headless: true}, want: Decision{Mode: "tui", Message: "No display server or interactive terminal detected; opening the terminal interface."}},
		{name: "noninteractive terminal fallback", options: Options{}, want: Decision{Mode: "gui", Chooser: true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Decide(test.options)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("got %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestSavedLauncherPreferenceUsesPrivateUserConfig(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", directory)
	t.Setenv("APPDATA", directory)
	t.Setenv("HOME", directory)
	if err := SavePreference("tui"); err != nil {
		t.Fatal(err)
	}
	if ReadPreference() != "tui" {
		t.Fatal("saved terminal preference was not loaded")
	}
	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
		t.Fatalf("launcher preference is not private: %v", info.Mode().Perm())
	}
	if err = SavePreference(""); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Clean(path)); !os.IsNotExist(err) {
		t.Fatalf("preference was not removed: %v", err)
	}
}

func TestDecideRejectsUnknownUI(t *testing.T) {
	if _, err := Decide(Options{UI: "browser"}); err == nil {
		t.Fatal("expected invalid UI error")
	}
}
