package notifier

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func Entry(goos, executable, home string, database ...string) (string, []byte, error) {
	databasePath := ""
	if len(database) > 0 {
		databasePath = database[0]
	}
	if home == "" || executable == "" {
		return "", nil, errors.New("home directory and application executable are required")
	}
	switch goos {
	case "linux":
		path := filepath.Join(home, ".config", "autostart", "attackpath-visualizer-notifier.desktop")
		command := escapeDesktop(executable) + " notifier run"
		if databasePath != "" {
			command += " --db " + escapeDesktop(databasePath)
		}
		return path, []byte("[Desktop Entry]\nType=Application\nName=AttackPath Visualizer Notifications\nComment=Opt-in local graph notifications\nExec=" + command + "\nTerminal=false\nX-GNOME-Autostart-enabled=true\n"), nil
	case "darwin":
		path := filepath.Join(home, "Library", "LaunchAgents", "dev.nexza.attackpathvisualizer.notifier.plist")
		args := "<string>" + escapeXML(executable) + "</string><string>notifier</string><string>run</string>"
		if databasePath != "" {
			args += "<string>--db</string><string>" + escapeXML(databasePath) + "</string>"
		}
		return path, []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n<plist version=\"1.0\"><dict><key>Label</key><string>dev.nexza.attackpathvisualizer.notifier</string><key>ProgramArguments</key><array>" + args + "</array><key>RunAtLoad</key><true/><key>StartInterval</key><integer>900</integer></dict></plist>\n"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		path := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup", "attackpath-visualizer-notifier.cmd")
		quoted := strings.ReplaceAll(executable, "\"", "\"\"")
		arguments := "notifier run"
		if databasePath != "" {
			databaseQuoted := strings.ReplaceAll(databasePath, "\"", "\"\"")
			arguments += " --db \"" + databaseQuoted + "\""
		}
		return path, []byte("@echo off\r\nstart \"\" \"" + quoted + "\" " + arguments + "\r\n"), nil
	default:
		return "", nil, fmt.Errorf("background notifier installation is unsupported on %s", goos)
	}
}

func Install(executable string, database ...string) (string, error) {
	databasePath := ""
	if len(database) > 0 {
		databasePath = database[0]
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find user home: %w", err)
	}
	path, contents, err := Entry(runtime.GOOS, executable, home, databasePath)
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", fmt.Errorf("create autostart directory: %w", err)
	}
	mode := os.FileMode(0600)
	if runtime.GOOS == "linux" {
		mode = 0700
	}
	if err = os.WriteFile(path, contents, mode); err != nil {
		return "", fmt.Errorf("install notifier autostart entry: %w", err)
	}
	return path, nil
}

func Uninstall() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find user home: %w", err)
	}
	path, _, err := Entry(runtime.GOOS, "attackpath-visualizer", home)
	if err != nil {
		return "", err
	}
	if err = os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("remove notifier autostart entry: %w", err)
	}
	return path, nil
}

func Status() (string, bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false, err
	}
	path, _, err := Entry(runtime.GOOS, "attackpath-visualizer", home)
	if err != nil {
		return "", false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return path, true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return path, false, nil
	}
	return path, false, err
}

func escapeDesktop(value string) string {
	return "\"" + strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "`", "\\`", "$", "\\$", "%", "%%").Replace(value) + "\""
}

func escapeXML(value string) string {
	var output strings.Builder
	_ = xml.EscapeText(&output, []byte(value))
	return output.String()
}
