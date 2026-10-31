//go:build !nogui

package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/NexzaDev/attackpath-visualizer/assets"
)

type productTheme struct {
	dark bool
}

func (t productTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	dark := t.dark || variant == theme.VariantDark
	if dark {
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 15, G: 24, B: 39, A: 255}
		case theme.ColorNameForeground:
			return color.NRGBA{R: 231, G: 237, B: 245, A: 255}
		case theme.ColorNamePrimary, theme.ColorNameFocus, theme.ColorNameHyperlink:
			return color.NRGBA{R: 125, G: 151, B: 255, A: 255}
		case theme.ColorNameButton, theme.ColorNameHeaderBackground, theme.ColorNameInputBackground:
			return color.NRGBA{R: 29, G: 43, B: 62, A: 255}
		case theme.ColorNameSeparator, theme.ColorNameInputBorder:
			return color.NRGBA{R: 55, G: 72, B: 93, A: 255}
		case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
			return color.NRGBA{R: 147, G: 161, B: 180, A: 255}
		case theme.ColorNameSelection:
			return color.NRGBA{R: 45, G: 75, B: 118, A: 255}
		case theme.ColorNameError:
			return color.NRGBA{R: 255, G: 128, B: 126, A: 255}
		default:
			return color.NRGBA{R: 23, G: 35, B: 52, A: 255}
		}
	}
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 247, G: 249, B: 252, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 28, G: 41, B: 59, A: 255}
	case theme.ColorNamePrimary, theme.ColorNameFocus, theme.ColorNameHyperlink:
		return color.NRGBA{R: 55, G: 88, B: 176, A: 255}
	case theme.ColorNameButton, theme.ColorNameHeaderBackground, theme.ColorNameInputBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameSeparator, theme.ColorNameInputBorder:
		return color.NRGBA{R: 218, G: 225, B: 234, A: 255}
	case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 95, G: 110, B: 131, A: 255}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 223, G: 232, B: 250, A: 255}
	case theme.ColorNameError:
		return color.NRGBA{R: 176, G: 38, B: 48, A: 255}
	default:
		return color.NRGBA{R: 242, G: 245, B: 249, A: 255}
	}
}

func (productTheme) Font(style fyne.TextStyle) fyne.Resource {
	name := "fonts/NotoSans-Regular.ttf"
	if style.Bold && style.Italic {
		name = "fonts/NotoSans-BoldItalic.ttf"
	} else if style.Bold {
		name = "fonts/NotoSans-Bold.ttf"
	} else if style.Italic {
		name = "fonts/NotoSans-Italic.ttf"
	}
	data, err := assets.Files.ReadFile(name)
	if err != nil {
		return theme.DefaultTheme().Font(style)
	}
	return fyne.NewStaticResource(name, data)
}

func (productTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (productTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 6
	case theme.SizeNameInlineIcon:
		return 18
	case theme.SizeNameText:
		return 13
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameHeadingText:
		return 25
	case theme.SizeNameSubHeadingText:
		return 17
	case theme.SizeNameLineSpacing:
		return 4
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameScrollBar:
		return 8
	case theme.SizeNameScrollBarSmall:
		return 4
	default:
		return 6
	}
}
