//go:build !nogui

package main

import "github.com/NexzaDev/attackpath-visualizer/gui"

func launchGUI(dbPath string) error {
	return gui.Run(dbPath)
}

func launchChooser(dbPath string) error {
	return gui.ShowChooser(dbPath)
}
