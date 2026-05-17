package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type appTheme struct{}

var _ fyne.Theme = (*appTheme)(nil)

func (t *appTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return ColorBG
	case theme.ColorNameButton:
		return ColorSurface
	case theme.ColorNameDisabledButton:
		return ColorSurface2
	case theme.ColorNameDisabled:
		return ColorMuted
	case theme.ColorNameError:
		return ColorOpponent
	case theme.ColorNameFocus:
		return ColorUser
	case theme.ColorNameForeground:
		return ColorText
	case theme.ColorNameHover:
		return ColorSurface2
	case theme.ColorNameInputBackground:
		return ColorSurface2
	case theme.ColorNameInputBorder:
		return ColorDim
	case theme.ColorNameMenuBackground:
		return ColorSurface
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 8, G: 9, B: 14, A: 220}
	case theme.ColorNamePlaceHolder:
		return ColorMuted
	case theme.ColorNamePressed:
		return ColorSurface2
	case theme.ColorNameScrollBar:
		return ColorDim
	case theme.ColorNameSeparator:
		return ColorStroke
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 180}
	case theme.ColorNamePrimary:
		return ColorUser
	case theme.ColorNameSuccess:
		return ColorUser
	case theme.ColorNameWarning:
		return color.NRGBA{R: 255, G: 184, B: 0, A: 255}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t *appTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *appTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *appTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 13
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameInnerPadding:
		return 4
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameScrollBar:
		return 4
	case theme.SizeNameScrollBarSmall:
		return 2
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameInputRadius:
		return 2
	case theme.SizeNameSelectionRadius:
		return 2
	}
	return theme.DefaultTheme().Size(name)
}
