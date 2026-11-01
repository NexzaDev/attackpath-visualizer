//go:build !nogui

package main

import (
	"flag"
	"fmt"
	"os"

	"fyne.io/fyne/v2/test"
	"github.com/NexzaDev/attackpath-visualizer/gui"
)

func main() {
	output := flag.String("out", "artifacts/ui", "screenshot output directory")
	flag.Parse()
	application := test.NewApp()
	if err := gui.RenderScreenshots(*output, application); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
