//go:build nogui

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/NexzaDev/attackpath-visualizer/tui"
)

func launchGUI(string) error {
	return errors.New("the graphical dashboard is unavailable in this nogui build; rebuild without -tags nogui")
}

func launchChooser(dbPath string) error {
	fmt.Fprintln(os.Stderr, "The graphical dashboard is unavailable in this nogui build; opening the terminal interface.")
	return tui.Run(dbPath)
}
