package notifier

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestEntryForEachPlatform(t *testing.T) {
	tests := []struct {
		goos string
		path string
		text string
	}{
		{goos: "linux", path: ".config/autostart/attackpath-visualizer-notifier.desktop", text: "notifier run"},
		{goos: "darwin", path: "Library/LaunchAgents/dev.nexza.attackpathvisualizer.notifier.plist", text: "StartInterval"},
		{goos: "windows", path: "Startup/attackpath-visualizer-notifier.cmd", text: "@echo off"},
	}
	for _, test := range tests {
		t.Run(test.goos, func(t *testing.T) {
			path, body, err := Entry(test.goos, "/Users/Example App/attackpath-visualizer", "/Users/example", "/var/data/graph file.db")
			if err != nil {
				t.Fatal(err)
			}
			portablePath := strings.ReplaceAll(path, "\\", "/")
			if !strings.Contains(portablePath, test.path) || !strings.Contains(string(body), test.text) || !strings.Contains(string(body), "graph file.db") {
				t.Fatalf("path=%q body=%q", path, body)
			}
		})
	}
}

func TestMacLaunchAgentEscapesArgumentsAsXML(t *testing.T) {
	_, body, err := Entry("darwin", "/Applications/A&B<app>.app", "/Users/example", "/data/one&two.db")
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Arguments []string `xml:"dict>array>string"`
	}
	if err = xml.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid launch agent XML: %v", err)
	}
	want := []string{"/Applications/A&B<app>.app", "notifier", "run", "--db", "/data/one&two.db"}
	if strings.Join(parsed.Arguments, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("arguments = %q", parsed.Arguments)
	}
}

func TestEntryRejectsIncompleteAndUnknownPlatforms(t *testing.T) {
	if _, _, err := Entry("linux", "", "/tmp/home"); err == nil {
		t.Fatal("expected missing executable error")
	}
	if _, _, err := Entry("plan9", "/app", "/tmp/home"); err == nil {
		t.Fatal("expected unsupported OS error")
	}
}
