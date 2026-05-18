package main

import (
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

const ms = time.Millisecond

// ── Easing ───────────────────────────────────────────────────────────────────

func easeInF(t float32) float32    { return t * t }
func easeOutF(t float32) float32   { return t * (2 - t) }
func easeInOutF(t float32) float32 { return t * t * (3 - 2*t) }

// overshootF rises past 1.0 then settles — used for bounce-in.
func overshootF(t float32) float32 {
	if t < 0.7 {
		return easeOutF(t / 0.7)
	}
	return 1 + float32(math.Sin(float64((t-0.7)/0.3*math.Pi)))*0.12
}

func lerpF(a, b, t float32) float32 { return a + (b-a)*t }

func lerpColor(from, to color.NRGBA, t float32) color.NRGBA {
	return color.NRGBA{
		R: uint8(lerpF(float32(from.R), float32(to.R), t)),
		G: uint8(lerpF(float32(from.G), float32(to.G), t)),
		B: uint8(lerpF(float32(from.B), float32(to.B), t)),
		A: uint8(lerpF(float32(from.A), float32(to.A), t)),
	}
}

func transparent(col color.NRGBA) color.NRGBA {
	return color.NRGBA{R: col.R, G: col.G, B: col.B, A: 0}
}

func toNRGBA(c color.Color) color.NRGBA {
	if nc, ok := c.(color.NRGBA); ok {
		return nc
	}
	r, g, b, a := c.RGBA()
	if a == 0 {
		return color.NRGBA{}
	}
	return color.NRGBA{
		R: uint8(r * 0xff / a),
		G: uint8(g * 0xff / a),
		B: uint8(b * 0xff / a),
		A: uint8(a >> 8),
	}
}

// fadeText fades txt from alpha 0 to its current Color.
// Sets txt.Color to transparent immediately. Caller calls .Start() (or via after()).
func fadeText(txt *canvas.Text, duration time.Duration, ease func(float32) float32) *fyne.Animation {
	to := toNRGBA(txt.Color)
	txt.Color = transparent(to)
	txt.Refresh()
	a := fyne.NewAnimation(duration, func(t float32) {
		txt.Color = lerpColor(transparent(to), to, ease(t))
		txt.Refresh()
	})
	a.Curve = fyne.AnimationLinear
	return a
}

// fadeRect animates rect.FillColor alpha from 0 to targetAlpha.
func fadeRect(rect *canvas.Rectangle, targetAlpha uint8, duration time.Duration) *fyne.Animation {
	c := toNRGBA(rect.FillColor)
	from := color.NRGBA{R: c.R, G: c.G, B: c.B, A: 0}
	to := color.NRGBA{R: c.R, G: c.G, B: c.B, A: targetAlpha}
	rect.FillColor = from
	rect.Refresh()
	a := fyne.NewAnimation(duration, func(t float32) {
		rect.FillColor = lerpColor(from, to, easeInOutF(t))
		rect.Refresh()
	})
	a.Curve = fyne.AnimationLinear
	return a
}

// popText pulses txt.TextSize with a sine envelope (up then back to baseSize).
func popText(txt *canvas.Text, baseSize, scaleTo float32, duration time.Duration) *fyne.Animation {
	a := fyne.NewAnimation(duration, func(t float32) {
		scale := 1 + (scaleTo-1)*float32(math.Sin(float64(t)*math.Pi))
		txt.TextSize = baseSize * scale
		txt.Refresh()
	})
	a.Curve = fyne.AnimationLinear
	return a
}

// flashRect pulses rect.FillColor alpha baseAlpha → peakAlpha → baseAlpha.
func flashRect(rect *canvas.Rectangle, baseAlpha, peakAlpha uint8, duration time.Duration) *fyne.Animation {
	c := toNRGBA(rect.FillColor)
	a := fyne.NewAnimation(duration, func(t float32) {
		alpha := baseAlpha + uint8(float32(peakAlpha-baseAlpha)*float32(math.Sin(float64(t)*math.Pi)))
		rect.FillColor = color.NRGBA{R: c.R, G: c.G, B: c.B, A: alpha}
		rect.Refresh()
	})
	a.Curve = fyne.AnimationLinear
	return a
}

// glowLoop returns an infinite alpha-breathing animation for a canvas.Circle.
func glowLoop(c *canvas.Circle, base color.NRGBA, minA, maxA uint8) *fyne.Animation {
	a := fyne.NewAnimation(2000*ms, func(t float32) {
		alpha := uint8(float32(minA) + float32(maxA-minA)*
			float32((math.Sin(float64(t)*2*math.Pi)+1)*0.5))
		c.FillColor = color.NRGBA{R: base.R, G: base.G, B: base.B, A: alpha}
		c.Refresh()
	})
	a.Curve = fyne.AnimationLinear
	a.RepeatCount = fyne.AnimationRepeatForever
	return a
}

// countUp animates an integer from 0 to target, calling setText each tick.
func countUp(setText func(int), target int, duration time.Duration) *fyne.Animation {
	a := fyne.NewAnimation(duration, func(t float32) {
		setText(int(lerpF(0, float32(target), easeOutF(t))))
	})
	a.Curve = fyne.AnimationLinear
	return a
}

// after calls fn after delay (synchronously if delay <= 0).
func after(delay time.Duration, fn func()) {
	if delay <= 0 {
		fn()
		return
	}
	time.AfterFunc(delay, fn)
}
