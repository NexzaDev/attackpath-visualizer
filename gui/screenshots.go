//go:build !nogui

package gui

import (
	"context"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/software"
	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
	"github.com/NexzaDev/attackpath-visualizer/internal/service"
)

func RenderScreenshots(outputDir string, app fyne.App) error {
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return fmt.Errorf("create screenshot directory: %w", err)
	}
	store, err := repository.Open(":memory:")
	if err != nil {
		return err
	}
	defer store.Close()
	if err = (service.Service{Store: store}).LoadDemo(context.Background()); err != nil {
		return err
	}
	dashboard, err := NewDashboard(app, nil, store, nil)
	if err != nil {
		return err
	}
	for _, category := range Categories {
		dashboard.current = category.ID
		canvas := software.NewCanvas()
		canvas.SetPadded(false)
		canvas.Resize(fyne.NewSize(1440, 900))
		canvas.SetContent(dashboard.Content())
		imageData := software.RenderCanvas(canvas, productTheme{})
		file, err := os.Create(filepath.Join(outputDir, category.ID+".png"))
		if err != nil {
			return fmt.Errorf("create %s screenshot: %w", category.ID, err)
		}
		if err = png.Encode(file, imageData); err != nil {
			file.Close()
			return fmt.Errorf("write %s screenshot: %w", category.ID, err)
		}
		if err = file.Close(); err != nil {
			return fmt.Errorf("close %s screenshot: %w", category.ID, err)
		}
	}
	return nil
}
