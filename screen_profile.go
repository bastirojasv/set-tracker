package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

func buildProfileContent(a *App, onClose func()) fyne.CanvasObject {
	profile := a.profile

	// Working copies — only written to profile on Save
	selectedColor := profile.Color
	photoPath := profile.PhotoPath
	name := profile.Name

	// --- Avatar canvas objects ---
	avatarInitial := canvas.NewText("", ColorText)
	avatarInitial.TextSize = 36
	avatarInitial.TextStyle = fyne.TextStyle{Bold: true}
	avatarInitial.Alignment = fyne.TextAlignCenter

	avatarCircle := canvas.NewCircle(ParseHexColor(selectedColor))

	// avatarStack is declared here so updateAvatar can reference it before it's initialized
	var avatarStack *fyne.Container

	updateAvatar := func() {
		if photoPath != "" {
			img := canvas.NewImageFromFile(photoPath)
			img.FillMode = canvas.ImageFillContain
			avatarStack.Objects = []fyne.CanvasObject{img}
		} else {
			n := name
			if len(n) > 0 {
				avatarInitial.Text = strings.ToUpper(string([]rune(n)[0]))
			} else {
				avatarInitial.Text = "?"
			}
			avatarCircle.FillColor = ParseHexColor(selectedColor)
			avatarInitial.Refresh()
			avatarCircle.Refresh()
			avatarStack.Objects = []fyne.CanvasObject{avatarCircle, container.NewCenter(avatarInitial)}
		}
		avatarStack.Refresh()
	}

	avatarStack = container.NewStack(avatarCircle, container.NewCenter(avatarInitial))
	avatarContainer := container.NewGridWrap(fyne.NewSize(88, 88), avatarStack)
	updateAvatar()

	// --- Stats ---
	records, _ := a.db.LoadHistory()
	total, wins, losses := len(records), 0, 0
	for _, r := range records {
		if r.UserWon != nil && *r.UserWon {
			wins++
		} else {
			losses++
		}
	}
	wr := 0
	if total > 0 {
		wr = wins * 100 / total
	}
	statsLbl := widget.NewLabel(fmt.Sprintf("%d SETS  ·  %dW  ·  %dL  ·  %d%% WR", total, wins, losses, wr))
	statsLbl.TextStyle = fyne.TextStyle{Monospace: true}

	// --- Photo ---
	photoNameLbl := canvas.NewText("", ColorMuted)
	photoNameLbl.TextSize = 11
	if photoPath != "" {
		photoNameLbl.Text = filepath.Base(photoPath)
	}

	photoBtn := widget.NewButton("Subir imagen", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			photoPath = reader.URI().Path()
			reader.Close()
			photoNameLbl.Text = filepath.Base(photoPath)
			photoNameLbl.Refresh()
			updateAvatar()
		}, a.win)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
		fd.Show()
	})
	photoBtn.Importance = widget.LowImportance

	nameLbl := canvas.NewText(strings.ToUpper(profile.Name), ParseHexColor(selectedColor))
	nameLbl.TextSize = 24
	nameLbl.TextStyle = fyne.TextStyle{Bold: true}

	avatarBlock := container.NewHBox(
		avatarContainer,
		container.NewVBox(
			nameLbl,
			statsLbl,
			container.NewHBox(photoBtn, container.NewCenter(photoNameLbl)),
		),
	)

	// --- Name field ---
	nameCaption := canvas.NewText("NOMBRE DE USUARIO", ColorMuted)
	nameCaption.TextSize = 11

	nameEntry := widget.NewEntry()
	nameEntry.SetText(profile.Name)
	nameEntry.TextStyle = fyne.TextStyle{Bold: true}
	nameEntry.OnChanged = func(s string) {
		name = s
		nameLbl.Text = strings.ToUpper(s)
		nameLbl.Refresh()
		updateAvatar()
	}

	// --- Color picker ---
	colorCaption := canvas.NewText("TU COLOR · IDENTIDAD PERMANENTE", ColorMuted)
	colorCaption.TextSize = 11

	swatchRow := container.NewHBox()
	var updateSwatches func()
	updateSwatches = func() {
		swatchRow.Objects = nil
		for _, hex := range PresetColors {
			hex := hex
			sw := NewColorSwatch(hex, hex == selectedColor, func(h string) {
				selectedColor = h
				nameLbl.Color = ParseHexColor(h)
				nameLbl.Refresh()
				updateAvatar()
				updateSwatches()
			})
			swatchRow.Add(sw)
		}
		swatchRow.Refresh()
	}
	updateSwatches()

	// --- Feedback label ---
	feedbackLbl := canvas.NewText("", color.Transparent)
	feedbackLbl.TextSize = 16
	feedbackLbl.TextStyle = fyne.TextStyle{Bold: true}
	feedbackLbl.Alignment = fyne.TextAlignCenter

	// --- Save button ---
	saveBtn := widget.NewButton("Guardar perfil", func() {
		n := nameEntry.Text
		if n == "" {
			n = "Jugador"
		}
		// Validate photo still exists
		if photoPath != "" {
			if _, err := os.Stat(photoPath); err != nil {
				photoPath = ""
				photoNameLbl.Text = ""
				photoNameLbl.Refresh()
				updateAvatar()
			}
		}
		profile.Name = n
		profile.Color = selectedColor
		profile.PhotoPath = photoPath

		if err := a.db.SaveProfile(profile); err != nil {
			feedbackLbl.Text = "Error al guardar: " + err.Error()
			feedbackLbl.Color = ColorOpponent
			feedbackLbl.Refresh()
			return
		}
		a.profile = profile
		a.refreshProfileWidget()

		feedbackLbl.Text = "✓  Perfil guardado correctamente"
		feedbackLbl.Color = ColorUser
		feedbackLbl.Refresh()
	})
	saveBtn.Importance = widget.HighImportance

	// --- Header with X close button ---
	titleLbl := canvas.NewText("EDITAR PERFIL", ColorMuted)
	titleLbl.TextSize = 11

	closeBtn := widget.NewButton("✕", func() { onClose() })
	closeBtn.Importance = widget.LowImportance

	header := container.NewBorder(nil, nil, nil, closeBtn,
		container.NewCenter(titleLbl),
	)

	return container.NewVBox(
		header,
		widget.NewSeparator(),
		avatarBlock,
		widget.NewSeparator(),
		nameCaption,
		nameEntry,
		widget.NewSeparator(),
		colorCaption,
		swatchRow,
		widget.NewSeparator(),
		container.NewCenter(feedbackLbl),
		saveBtn,
	)
}
