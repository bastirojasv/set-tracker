package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (a *App) showNewSet() {
	a.setActiveTab(a.tabSet)
	a.setBody(a.buildNewSetScreen())
}

func (a *App) buildNewSetScreen() fyne.CanvasObject {
	selectedOppColor := "#FF4500"
	selectedSetType := "FT"
	selectedNum := 5

	opponents, _ := a.db.LoadOpponents()

	// --- Opponent name ---
	oppLabel := newLabel("NOMBRE DEL ADVERSARIO", 11, true, ColorMuted)
	oppEntry := widget.NewEntry()
	oppEntry.SetPlaceHolder("Nombre del rival")
	oppEntry.TextStyle = fyne.TextStyle{Bold: true}
	oppEntry.OnChanged = func(name string) {
		if name != "" {
			c := a.db.LoadOpponentColor(name)
			if c != "" {
				selectedOppColor = c
			}
		}
	}

	suggestBox := container.NewHBox()
	for _, opp := range opponents {
		opp := opp
		chip := widget.NewButton(opp, func() {
			oppEntry.SetText(opp)
			selectedOppColor = a.db.LoadOpponentColor(opp)
		})
		chip.Importance = widget.LowImportance
		suggestBox.Add(chip)
	}

	// --- Color picker ---
	colorLabel := newLabel("COLOR DEL ADVERSARIO", 11, true, ColorMuted)
	swatchRow := container.NewHBox()
	var updateSwatches func()
	updateSwatches = func() {
		swatchRow.Objects = nil
		for _, hex := range PresetColors {
			hex := hex
			sw := NewColorSwatch(hex, hex == selectedOppColor, func(h string) {
				selectedOppColor = h
				updateSwatches()
			})
			swatchRow.Add(sw)
		}
		swatchRow.Refresh()
	}
	updateSwatches()

	// --- Set type ---
	typeLabel := newLabel("TIPO DE SET", 11, true, ColorMuted)

	descLabel := canvas.NewText("", ColorMuted)
	descLabel.TextSize = 12

	noLimit := false
	savedNum := selectedNum // guarda el valor al activar sin límite

	var updateDesc func()
	updateDesc = func() {
		switch selectedSetType {
		case "FT":
			descLabel.Text = fmt.Sprintf("Primero en ganar %d juegos", selectedNum)
		case "BO":
			descLabel.Text = fmt.Sprintf("Mejor de %d — gana quien llegue a %d", selectedNum, (selectedNum+1)/2)
		case "LS":
			if noLimit {
				descLabel.Text = "Set libre · sin límite · termina manualmente"
			} else {
				descLabel.Text = fmt.Sprintf("Set libre · puntaje máximo %d · termina manualmente", selectedNum)
			}
		}
		descLabel.Refresh()
	}

	stepper := NewStepper(&selectedNum, 1, 99, func(_ int) { updateDesc() })

	noLimitCheck := widget.NewCheck("Sin límite", func(checked bool) {
		noLimit = checked
		if checked {
			savedNum = selectedNum
			selectedNum = 0
			stepper.Hide()
		} else {
			selectedNum = savedNum
			stepper.Show()
		}
		updateDesc()
	})
	noLimitCheck.Hide() // solo visible en modo LS

	toggle := NewSetTypeToggle(&selectedSetType, func(val string) {
		selectedSetType = val
		if val == "LS" {
			noLimitCheck.Show()
		} else {
			// al salir de LS, resetear estado sin límite
			if noLimit {
				noLimit = false
				noLimitCheck.Checked = false
				noLimitCheck.Refresh()
				selectedNum = savedNum
				stepper.Show()
			}
			noLimitCheck.Hide()
		}
		updateDesc()
	})

	updateDesc()

	typeRow := container.NewHBox(toggle, noLimitCheck, stepper, container.NewPadded(descLabel))

	// --- CTA ---
	startBtn := NewActionButton("Comenzar Set", ColorUser, func() {
		name := oppEntry.Text
		if name == "" {
			a.showError("Ingresa el nombre del adversario")
			return
		}
		state, err := a.db.StartSet(selectedSetType, selectedNum, name, selectedOppColor)
		if err != nil {
			a.showError("Error al iniciar set: " + err.Error())
			return
		}
		a.activeSet = state
		a.showActiveSet()
	})

	// --- Title ---
	titleTxt := canvas.NewText("CONFIGURAR PARTIDA", ColorMuted)
	titleTxt.TextSize = 11
	titleTxt.Alignment = fyne.TextAlignCenter
	mainTitle := canvas.NewText("Nuevo Set", ColorText)
	mainTitle.TextSize = 52
	mainTitle.TextStyle = fyne.TextStyle{Bold: true}
	mainTitle.Alignment = fyne.TextAlignCenter
	titleBlock := container.NewVBox(
		container.NewCenter(titleTxt),
		container.NewCenter(mainTitle),
	)

	form := container.NewVBox(
		titleBlock,
		widget.NewSeparator(),
		oppLabel,
		oppEntry,
		suggestBox,
		widget.NewSeparator(),
		colorLabel,
		swatchRow,
		widget.NewSeparator(),
		typeLabel,
		typeRow,
		widget.NewSeparator(),
		startBtn,
	)

	// Invisible sizer para dar ancho mínimo al card sin romper el layout
	sizer := canvas.NewRectangle(color.Transparent)
	sizer.SetMinSize(fyne.NewSize(560, 1))

	formCard := container.NewStack(
		sizer,
		newSurface(ColorSurface),
		container.NewPadded(container.NewPadded(form)),
	)

	// Centrado vertical y horizontal usando spacers
	centered := container.NewStack(
		canvas.NewRectangle(ColorBG),
		container.NewVBox(
			layout.NewSpacer(),
			container.NewHBox(layout.NewSpacer(), formCard, layout.NewSpacer()),
			layout.NewSpacer(),
		),
	)
	return centered
}
