package main

import (
	"fmt"
	"image/color"
)

// Design tokens — matching SetTracker mockup palette
var (
	ColorBG       = color.NRGBA{R: 8, G: 9, B: 14, A: 255}
	ColorSurface  = color.NRGBA{R: 17, G: 19, B: 27, A: 255}
	ColorSurface2 = color.NRGBA{R: 24, G: 27, B: 38, A: 255}
	ColorText     = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	ColorMuted    = color.NRGBA{R: 122, G: 125, B: 140, A: 255}
	ColorDim      = color.NRGBA{R: 74, G: 77, B: 92, A: 255}
	ColorStroke   = color.NRGBA{R: 255, G: 255, B: 255, A: 18}
	ColorUser     = color.NRGBA{R: 0, G: 191, B: 255, A: 255}   // #00BFFF
	ColorOpponent = color.NRGBA{R: 255, G: 69, B: 0, A: 255}    // #FF4500
)

// Preset colors for the color picker
var PresetColors = []string{
	"#FF4500", "#FFB800", "#22D36F", "#00BFFF",
	"#A06CFF", "#FF3D7F", "#FFFFFF", "#7A7D8C",
}

func ParseHexColor(s string) color.NRGBA {
	var r, g, b uint8
	fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}

func ColorToHex(c color.NRGBA) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// WithAlpha returns the color with reduced alpha for tinted backgrounds
func WithAlpha(c color.NRGBA, a uint8) color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: a}
}
