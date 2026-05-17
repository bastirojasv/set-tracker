package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) showOnboarding() {
	selectedColor := "#00BFFF"

	bg := canvas.NewRectangle(ColorBG)

	title := canvas.NewText("Bienvenido a", ColorMuted)
	title.TextSize = 14
	title.Alignment = fyne.TextAlignCenter

	appName := canvas.NewText("SET·TRACKER", ColorUser)
	appName.TextSize = 48
	appName.TextStyle = fyne.TextStyle{Bold: true}
	appName.Alignment = fyne.TextAlignCenter

	subtitle := canvas.NewText("Configura tu perfil para empezar", ColorMuted)
	subtitle.TextSize = 13
	subtitle.Alignment = fyne.TextAlignCenter

	// Name input
	nameLabel := canvas.NewText("TU NOMBRE", ColorMuted)
	nameLabel.TextSize = 11

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Tu nombre de jugador")
	nameEntry.TextStyle = fyne.TextStyle{Bold: true}

	// Color picker
	colorLabel := canvas.NewText("TU COLOR", ColorMuted)
	colorLabel.TextSize = 11

	swatches := container.NewHBox()
	var updateSwatches func()
	updateSwatches = func() {
		swatches.Objects = nil
		for _, hex := range PresetColors {
			hex := hex
			active := hex == selectedColor
			sw := NewColorSwatch(hex, active, func(h string) {
				selectedColor = h
				updateSwatches()
			})
			swatches.Add(sw)
		}
		swatches.Refresh()
	}
	updateSwatches()

	// Preview avatar
	previewInitial := canvas.NewText("?", ColorText)
	previewInitial.TextSize = 32
	previewInitial.TextStyle = fyne.TextStyle{Bold: true}
	previewInitial.Alignment = fyne.TextAlignCenter

	previewBg := canvas.NewCircle(ColorUser)

	previewAvatar := container.NewStack(
		container.NewCenter(previewBg),
		container.NewCenter(previewInitial),
	)
	previewBg.Resize(fyne.NewSize(80, 80))

	nameEntry.OnChanged = func(s string) {
		if len(s) > 0 {
			previewInitial.Text = string([]rune(s)[0])
		} else {
			previewInitial.Text = "?"
		}
		previewInitial.Refresh()
	}

	startBtn := NewActionButton("Comenzar", ColorUser, func() {
		name := nameEntry.Text
		if name == "" {
			name = "Jugador"
		}
		p := &UserProfile{Name: name, Color: selectedColor}
		if err := a.db.SaveProfile(p); err != nil {
			return
		}
		a.profile = p
		a.buildMain()
		a.showNewSet()
	})

	form := container.NewVBox(
		nameLabel,
		nameEntry,
		widget.NewSeparator(),
		colorLabel,
		swatches,
		widget.NewSeparator(),
		container.NewCenter(previewAvatar),
		container.NewCenter(startBtn),
	)

	card := container.NewStack(
		newSurface(ColorSurface),
		container.NewPadded(container.NewPadded(form)),
	)

	content := container.NewCenter(
		container.NewVBox(
			container.NewCenter(title),
			container.NewCenter(appName),
			container.NewCenter(subtitle),
			widget.NewSeparator(),
			card,
		),
	)

	a.win.SetContent(container.NewStack(bg, content))
}
