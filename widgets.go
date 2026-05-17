package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// --- ActionButton: large colored button ---

type ActionButton struct {
	widget.BaseWidget
	label   string
	bgColor color.NRGBA
	onTap   func()
}

func NewActionButton(label string, bgColor color.NRGBA, onTap func()) *ActionButton {
	b := &ActionButton{label: label, bgColor: bgColor, onTap: onTap}
	b.ExtendBaseWidget(b)
	return b
}

func (b *ActionButton) Tapped(_ *fyne.PointEvent) {
	if b.onTap != nil {
		b.onTap()
	}
}
func (b *ActionButton) TappedSecondary(_ *fyne.PointEvent) {}

func (b *ActionButton) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(b.bgColor)
	bg.CornerRadius = 2

	lbl := canvas.NewText("▶ "+b.label, ColorBG)
	lbl.TextSize = 22
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	lbl.Alignment = fyne.TextAlignCenter

	return widget.NewSimpleRenderer(container.NewStack(bg, container.NewCenter(lbl)))
}

func (b *ActionButton) MinSize() fyne.Size { return fyne.NewSize(220, 64) }

// --- SmallButton: ghost/secondary button ---

type SmallButton struct {
	widget.BaseWidget
	label string
	onTap func()
}

func NewSmallButton(label string, onTap func()) *SmallButton {
	b := &SmallButton{label: label, onTap: onTap}
	b.ExtendBaseWidget(b)
	return b
}

func (b *SmallButton) Tapped(_ *fyne.PointEvent) {
	if b.onTap != nil {
		b.onTap()
	}
}
func (b *SmallButton) TappedSecondary(_ *fyne.PointEvent) {}

func (b *SmallButton) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeWidth = 1
	border.StrokeColor = ColorStroke
	border.CornerRadius = 2

	lbl := canvas.NewText(b.label, ColorMuted)
	lbl.TextSize = 11
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	lbl.Alignment = fyne.TextAlignCenter

	return widget.NewSimpleRenderer(container.NewStack(bg, border, container.NewCenter(lbl)))
}

func (b *SmallButton) MinSize() fyne.Size { return fyne.NewSize(160, 36) }

// --- ColorSwatch: color picker tile ---

type ColorSwatch struct {
	widget.BaseWidget
	hexColor string
	active   bool
	onTap    func(hex string)
}

func NewColorSwatch(hex string, active bool, onTap func(string)) *ColorSwatch {
	s := &ColorSwatch{hexColor: hex, active: active, onTap: onTap}
	s.ExtendBaseWidget(s)
	return s
}

func (s *ColorSwatch) Tapped(_ *fyne.PointEvent) {
	if s.onTap != nil {
		s.onTap(s.hexColor)
	}
}
func (s *ColorSwatch) TappedSecondary(_ *fyne.PointEvent) {}

func (s *ColorSwatch) CreateRenderer() fyne.WidgetRenderer {
	col := ParseHexColor(s.hexColor)
	bg := canvas.NewRectangle(col)
	bg.CornerRadius = 2
	if s.active {
		bg.StrokeColor = ColorText
		bg.StrokeWidth = 2
	} else {
		bg.StrokeColor = ColorStroke
		bg.StrokeWidth = 1
	}
	return widget.NewSimpleRenderer(bg)
}

func (s *ColorSwatch) MinSize() fyne.Size { return fyne.NewSize(40, 40) }

// --- SetTypeToggle: FT / BO segmented buttons ---

func NewSetTypeToggle(selectedType *string, onChange func(string)) fyne.CanvasObject {
	var ftBtn, boBtn, lsBtn *widget.Button

	update := func() {
		ftBtn.Importance = widget.LowImportance
		boBtn.Importance = widget.LowImportance
		lsBtn.Importance = widget.LowImportance
		switch *selectedType {
		case "FT":
			ftBtn.Importance = widget.HighImportance
		case "BO":
			boBtn.Importance = widget.HighImportance
		case "LS":
			lsBtn.Importance = widget.HighImportance
		}
		ftBtn.Refresh()
		boBtn.Refresh()
		lsBtn.Refresh()
	}

	ftBtn = widget.NewButton("FT", func() {
		*selectedType = "FT"
		update()
		if onChange != nil {
			onChange("FT")
		}
	})

	boBtn = widget.NewButton("BO", func() {
		*selectedType = "BO"
		update()
		if onChange != nil {
			onChange("BO")
		}
	})

	lsBtn = widget.NewButton("Libre", func() {
		*selectedType = "LS"
		update()
		if onChange != nil {
			onChange("LS")
		}
	})

	update()
	return container.NewHBox(ftBtn, boBtn, lsBtn)
}

// --- Stepper ---

func NewStepper(val *int, min, max int, onChange func(int)) fyne.CanvasObject {
	display := canvas.NewText(fmt.Sprintf("%d", *val), ColorText)
	display.TextSize = 24
	display.TextStyle = fyne.TextStyle{Bold: true}
	display.Alignment = fyne.TextAlignCenter

	minus := widget.NewButton("−", func() {
		if *val > min {
			*val--
			display.Text = fmt.Sprintf("%d", *val)
			display.Refresh()
			if onChange != nil {
				onChange(*val)
			}
		}
	})
	minus.Importance = widget.LowImportance

	plus := widget.NewButton("+", func() {
		if *val < max {
			*val++
			display.Text = fmt.Sprintf("%d", *val)
			display.Refresh()
			if onChange != nil {
				onChange(*val)
			}
		}
	})
	plus.Importance = widget.LowImportance

	displayBox := container.NewGridWrap(fyne.NewSize(48, 40), container.NewCenter(display))
	return container.NewHBox(minus, displayBox, plus)
}

// --- Label helpers ---

func newLabel(text string, size float32, bold bool, col color.NRGBA) *canvas.Text {
	t := canvas.NewText(text, col)
	t.TextSize = size
	t.TextStyle = fyne.TextStyle{Bold: bold}
	return t
}

func newLabelCenter(text string, size float32, bold bool, col color.NRGBA) *canvas.Text {
	t := newLabel(text, size, bold, col)
	t.Alignment = fyne.TextAlignCenter
	return t
}

// --- tapAbsorber: absorbs taps so clicks inside a PopUp card don't close it ---

type tapAbsorber struct {
	widget.BaseWidget
	content fyne.CanvasObject
}

func (t *tapAbsorber) Tapped(_ *fyne.PointEvent)          {}
func (t *tapAbsorber) TappedSecondary(_ *fyne.PointEvent) {}
func (t *tapAbsorber) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.content)
}

// --- Surface card ---

func newSurface(col color.NRGBA) *canvas.Rectangle {
	r := canvas.NewRectangle(col)
	r.CornerRadius = 2
	r.StrokeColor = ColorStroke
	r.StrokeWidth = 1
	return r
}
