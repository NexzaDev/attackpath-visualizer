//go:build !nogui

package gui

import (
	"fmt"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/NexzaDev/attackpath-visualizer/assets"
	"github.com/NexzaDev/attackpath-visualizer/internal/config"
	"github.com/NexzaDev/attackpath-visualizer/internal/launcher"
	"github.com/NexzaDev/attackpath-visualizer/internal/meta"
	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
)

func Run(requestedDB string) error {
	application := app.NewWithID(meta.ID)
	application.Settings().SetTheme(productTheme{})
	path, err := config.DatabasePath(requestedDB)
	if err != nil {
		return err
	}
	store, err := repository.Open(path)
	if err != nil {
		return err
	}
	defer store.Close()
	window := application.NewWindow(meta.Name)
	setAppIcon(application, window)
	dashboard, err := NewDashboard(application, window, store, func(mode string) error {
		if mode == "tui" {
			if err := launcher.StartTerminal(); err != nil {
				return err
			}
			application.Quit()
		}
		return nil
	})
	if err != nil {
		return err
	}
	window.Resize(savedWindowSize(application))
	window.CenterOnScreen()
	window.SetOnClosed(func() { saveWindowSize(application, window) })
	window.SetContent(dashboard.Content())
	setMenu(window, dashboard)
	setShortcuts(window, dashboard)
	window.ShowAndRun()
	return nil
}

func ShowChooser(requestedDB string) error {
	application := app.NewWithID(meta.ID)
	application.Settings().SetTheme(productTheme{})
	window := application.NewWindow("How do you want to open AttackPath Visualizer?")
	setAppIcon(application, window)
	remember := widget.NewCheck("Remember my choice", nil)
	remember.SetChecked(launcher.ReadPreference() != "")
	launch := func(mode string) {
		if remember.Checked {
			if err := launcher.SavePreference(mode); err != nil {
				dialog.ShowError(err, window)
				return
			}
		} else if err := launcher.SavePreference(""); err != nil {
			dialog.ShowError(err, window)
			return
		}
		if mode == "tui" {
			if err := launcher.StartTerminal(); err != nil {
				dialog.ShowInformation("Terminal not found", err.Error(), window)
				return
			}
			application.Quit()
			return
		}
		path, err := config.DatabasePath(requestedDB)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		store, err := repository.Open(path)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		dashboardWindow := application.NewWindow(meta.Name)
		setAppIcon(application, dashboardWindow)
		dashboard, err := NewDashboard(application, dashboardWindow, store, func(mode string) error {
			if mode == "tui" {
				if err := launcher.StartTerminal(); err != nil {
					return err
				}
				application.Quit()
			}
			return nil
		})
		if err != nil {
			store.Close()
			dialog.ShowError(err, window)
			return
		}
		window.Hide()
		dashboardWindow.Resize(savedWindowSize(application))
		dashboardWindow.CenterOnScreen()
		dashboardWindow.SetContent(dashboard.Content())
		setMenu(dashboardWindow, dashboard)
		setShortcuts(dashboardWindow, dashboard)
		dashboardWindow.SetOnClosed(func() {
			saveWindowSize(application, dashboardWindow)
			if closeErr := store.Close(); closeErr != nil {
				fmt.Fprintln(os.Stderr, "close database:", closeErr)
			}
			application.Quit()
		})
		dashboardWindow.Show()
	}
	dashboardButton := widget.NewButton("Dashboard (graphical window)", func() { launch("gui") })
	dashboardButton.Importance = widget.HighImportance
	terminalButton := widget.NewButton("Terminal (text interface)", func() { launch("tui") })
	quitButton := widget.NewButton("Quit", application.Quit)
	iconData, _ := assets.Files.ReadFile("icons/icon.png")
	logo := widget.NewIcon(fyne.NewStaticResource("attackpath-visualizer.png", iconData))
	window.SetContent(container.NewCenter(container.NewVBox(container.NewCenter(logo), widget.NewLabelWithStyle("How do you want to open AttackPath Visualizer?", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}), dashboardButton, terminalButton, remember, quitButton)))
	window.Resize(fyne.NewSize(480, 330))
	window.CenterOnScreen()
	window.ShowAndRun()
	return nil
}

func savedWindowSize(application fyne.App) fyne.Size {
	width := application.Preferences().FloatWithFallback("window.width", 1360)
	height := application.Preferences().FloatWithFallback("window.height", 900)
	if width < 900 || width > 4000 {
		width = 1360
	}
	if height < 600 || height > 3000 {
		height = 900
	}
	return fyne.NewSize(float32(width), float32(height))
}

func saveWindowSize(application fyne.App, window fyne.Window) {
	size := window.Canvas().Size()
	if size.Width >= 900 && size.Height >= 600 {
		application.Preferences().SetFloat("window.width", float64(size.Width))
		application.Preferences().SetFloat("window.height", float64(size.Height))
	}
}

func setAppIcon(application fyne.App, window fyne.Window) {
	iconData, err := assets.Files.ReadFile("icons/icon.png")
	if err == nil {
		resource := fyne.NewStaticResource("attackpath-visualizer.png", iconData)
		application.SetIcon(resource)
		window.SetIcon(resource)
	}
}

func setMenu(window fyne.Window, dashboard *Dashboard) {
	fileMenu := fyne.NewMenu("Application", fyne.NewMenuItem("Export graph…", func() { dashboard.saveReport("graph") }), fyne.NewMenuItem("Switch to Terminal", func() {
		if dashboard.switchMode != nil {
			if err := dashboard.switchMode("tui"); err != nil {
				dialog.ShowError(err, window)
			}
		}
	}), fyne.NewMenuItem("Quit", func() { window.Close() }))
	helpMenu := fyne.NewMenu("Help", fyne.NewMenuItem("Usage guide  F1", func() { dashboard.show("about") }), fyne.NewMenuItem("About", func() { dashboard.show("about") }))
	window.SetMainMenu(fyne.NewMainMenu(fileMenu, helpMenu))
}

func setShortcuts(window fyne.Window, dashboard *Dashboard) {
	modifier := fyne.KeyModifierControl
	if runtime.GOOS == "darwin" {
		modifier = fyne.KeyModifierSuper
	}
	for index, category := range Categories[:9] {
		id := category.ID
		window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyName(fmt.Sprint(index + 1)), Modifier: modifier}, func(fyne.Shortcut) { dashboard.show(id) })
	}
	window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: modifier}, func(fyne.Shortcut) {
		dashboard.searchBox.FocusGained()
		window.Canvas().Focus(dashboard.searchBox)
	})
	window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: modifier}, func(fyne.Shortcut) { dashboard.show("settings") })
	window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyF1}, func(fyne.Shortcut) { dashboard.show("about") })
}
