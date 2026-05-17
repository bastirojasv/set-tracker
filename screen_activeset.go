package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (a *App) showActiveSet() {
	a.setActiveTab(a.tabSet)
	if a.activeSet == nil {
		a.showNewSet()
		return
	}
	a.setBody(a.buildActiveSetScreen())
}

func (a *App) buildActiveSetScreen() fyne.CanvasObject {
	if a.activeSet.IsLibre() {
		return a.buildLibreSetScreen()
	}
	s := a.activeSet
	userCol := ParseHexColor(a.profile.Color)
	oppCol := ParseHexColor(s.OpponentColor)

	// Score displays
	userScoreTxt := canvas.NewText(strconv.Itoa(s.UserScore), userCol)
	userScoreTxt.TextSize = 180
	userScoreTxt.TextStyle = fyne.TextStyle{Bold: true}
	userScoreTxt.Alignment = fyne.TextAlignCenter

	oppScoreTxt := canvas.NewText(strconv.Itoa(s.OppScore), oppCol)
	oppScoreTxt.TextSize = 180
	oppScoreTxt.TextStyle = fyne.TextStyle{Bold: true}
	oppScoreTxt.Alignment = fyne.TextAlignCenter

	// Player names
	userNameLbl := canvas.NewText(strings.ToUpper(a.profile.Name), userCol)
	userNameLbl.TextSize = 30
	userNameLbl.TextStyle = fyne.TextStyle{Bold: true}

	userSubLbl := canvas.NewText("JUGADOR", ColorMuted)
	userSubLbl.TextSize = 10

	oppNameLbl := canvas.NewText(strings.ToUpper(s.OpponentName), oppCol)
	oppNameLbl.TextSize = 30
	oppNameLbl.TextStyle = fyne.TextStyle{Bold: true}
	oppNameLbl.Alignment = fyne.TextAlignTrailing

	oppSubLbl := canvas.NewText("ADVERSARIO", ColorMuted)
	oppSubLbl.TextSize = 10
	oppSubLbl.Alignment = fyne.TextAlignTrailing

	// Game history pips
	userPips := a.buildGamePips(s, true, userCol)
	oppPips := a.buildGamePips(s, false, oppCol)

	// Action buttons
	userBtn := NewActionButton("YO GANÉ", userCol, func() {
		a.recordGame("user")
	})
	oppBtn := NewActionButton("RIVAL GANÓ", oppCol, func() {
		a.recordGame("opponent")
	})

	// Player panels
	userBg := canvas.NewRectangle(WithAlpha(userCol, 12))
	leftPanel := container.NewStack(
		userBg,
		container.NewPadded(container.NewBorder(
			container.NewVBox(userSubLbl, userNameLbl),
			container.NewVBox(container.NewCenter(userPips), userBtn),
			nil, nil,
			container.NewCenter(userScoreTxt),
		)),
	)

	oppBg := canvas.NewRectangle(WithAlpha(oppCol, 12))
	rightPanel := container.NewStack(
		oppBg,
		container.NewPadded(container.NewBorder(
			container.NewVBox(oppSubLbl, oppNameLbl),
			container.NewVBox(container.NewCenter(oppPips), oppBtn),
			nil, nil,
			container.NewCenter(oppScoreTxt),
		)),
	)

	// VS divider
	// Set type badge (top center overlay)
	setTypeTxt := canvas.NewText(fmt.Sprintf("%s %d", s.SetType, s.SetNumber), ColorText)
	setTypeTxt.TextSize = 22
	setTypeTxt.TextStyle = fyne.TextStyle{Bold: true}

	editBtn := widget.NewButton("✎", func() {
		a.showChangeSetTypeDialog(setTypeTxt)
	})
	editBtn.Importance = widget.LowImportance

	setTypeBg := newSurface(ColorSurface)
	setTypeCard := container.NewStack(
		setTypeBg,
		container.NewHBox(container.NewPadded(setTypeTxt), editBtn),
	)
	setTypeCaption := canvas.NewText("TIPO DE SET", ColorMuted)
	setTypeCaption.TextSize = 10
	setTypeCaption.Alignment = fyne.TextAlignCenter
	topCenter := container.NewVBox(
		container.NewCenter(setTypeCaption),
		container.NewCenter(setTypeCard),
	)

	// Center column: línea divisoria + VS arriba + undo/cancel abajo
	vsTxt := canvas.NewText("VS", ColorMuted)
	vsTxt.TextSize = 13
	vsTxt.TextStyle = fyne.TextStyle{Bold: true}
	vsTxt.Alignment = fyne.TextAlignCenter

	vsCircle := canvas.NewCircle(ColorSurface)
	vsCircle.StrokeColor = ColorStroke
	vsCircle.StrokeWidth = 1
	vsWidget := container.NewGridWrap(fyne.NewSize(40, 40),
		container.NewStack(vsCircle, container.NewCenter(vsTxt)),
	)

	undoBtn := NewSmallButton("↺ Deshacer", func() { a.undoLastGame() })
	cancelBtn := NewSmallButton("✕ Cancelar", func() {
		a.confirmDialog(
			"Cancelar set",
			"¿Seguro? El set no se guardará en el historial.",
			func() {
				a.db.AbandonActiveSet()
				a.activeSet = nil
				a.showNewSet()
			},
		)
	})

	divLine := canvas.NewRectangle(ColorStroke)

	centerDiv := container.NewBorder(
		nil,
		container.NewVBox(undoBtn, cancelBtn),
		nil, nil,
		container.NewStack(
			container.NewCenter(divLine),
			container.NewCenter(vsWidget),
		),
	)

	// Main split: columna central más ancha para acomodar los botones
	mainSplit := container.New(&threeColLayout{centerW: 120}, leftPanel, centerDiv, rightPanel)

	// Solo el set type como overlay en el top
	return container.NewStack(
		canvas.NewRectangle(ColorBG),
		container.NewStack(
			mainSplit,
			container.NewBorder(
				container.NewCenter(topCenter),
				nil, nil, nil,
			),
		),
	)
}

func (a *App) buildLibreSetScreen() fyne.CanvasObject {
	s := a.activeSet
	userCol := ParseHexColor(a.profile.Color)
	oppCol := ParseHexColor(s.OpponentColor)

	// Font size adaptativa: 1 dígito → 180, 2 dígitos → 130, 3 dígitos → 90
	adaptSize := func(n int) float32 {
		if n >= 100 {
			return 90
		}
		if n >= 10 {
			return 130
		}
		return 180
	}

	userScoreTxt := canvas.NewText(strconv.Itoa(s.UserScore), userCol)
	userScoreTxt.TextSize = adaptSize(s.UserScore)
	userScoreTxt.TextStyle = fyne.TextStyle{Bold: true}
	userScoreTxt.Alignment = fyne.TextAlignCenter

	oppScoreTxt := canvas.NewText(strconv.Itoa(s.OppScore), oppCol)
	oppScoreTxt.TextSize = adaptSize(s.OppScore)
	oppScoreTxt.TextStyle = fyne.TextStyle{Bold: true}
	oppScoreTxt.Alignment = fyne.TextAlignCenter

	// Nombres
	userNameLbl := canvas.NewText(strings.ToUpper(a.profile.Name), userCol)
	userNameLbl.TextSize = 30
	userNameLbl.TextStyle = fyne.TextStyle{Bold: true}
	userSubLbl := canvas.NewText("JUGADOR", ColorMuted)
	userSubLbl.TextSize = 10

	oppNameLbl := canvas.NewText(strings.ToUpper(s.OpponentName), oppCol)
	oppNameLbl.TextSize = 30
	oppNameLbl.TextStyle = fyne.TextStyle{Bold: true}
	oppNameLbl.Alignment = fyne.TextAlignTrailing
	oppSubLbl := canvas.NewText("ADVERSARIO", ColorMuted)
	oppSubLbl.TextSize = 10
	oppSubLbl.Alignment = fyne.TextAlignTrailing

	// Progreso — si SetNumber==0 es sin límite
	progText := func(score int) string {
		if s.SetNumber == 0 {
			return fmt.Sprintf("%d pts", score)
		}
		return fmt.Sprintf("%d / %d pts", score, s.SetNumber)
	}

	userProgLbl := canvas.NewText(progText(s.UserScore), ColorMuted)
	userProgLbl.TextSize = 11
	userProgLbl.Alignment = fyne.TextAlignCenter

	oppProgLbl := canvas.NewText(progText(s.OppScore), ColorMuted)
	oppProgLbl.TextSize = 11
	oppProgLbl.Alignment = fyne.TextAlignCenter

	// Botones +1 (sin límite cuando SetNumber==0)
	userBtn := NewActionButton("+1 PUNTO", userCol, func() {
		if s.SetNumber == 0 || s.UserScore < s.SetNumber {
			a.recordGame("user")
		}
	})
	oppBtn := NewActionButton("+1 PUNTO", oppCol, func() {
		if s.SetNumber == 0 || s.OppScore < s.SetNumber {
			a.recordGame("opponent")
		}
	})

	// Paneles laterales
	userBg := canvas.NewRectangle(WithAlpha(userCol, 12))
	leftPanel := container.NewStack(
		userBg,
		container.NewPadded(container.NewBorder(
			container.NewVBox(userSubLbl, userNameLbl),
			container.NewVBox(container.NewCenter(userProgLbl), userBtn),
			nil, nil,
			container.NewCenter(userScoreTxt),
		)),
	)

	oppBg := canvas.NewRectangle(WithAlpha(oppCol, 12))
	rightPanel := container.NewStack(
		oppBg,
		container.NewPadded(container.NewBorder(
			container.NewVBox(oppSubLbl, oppNameLbl),
			container.NewVBox(container.NewCenter(oppProgLbl), oppBtn),
			nil, nil,
			container.NewCenter(oppScoreTxt),
		)),
	)

	// Badge central con tipo
	badgeText := "Set Libre · sin límite"
	if s.SetNumber > 0 {
		badgeText = fmt.Sprintf("Set Libre · máx %d pts", s.SetNumber)
	}
	setTypeTxt := canvas.NewText(badgeText, ColorText)
	setTypeTxt.TextSize = 18
	setTypeTxt.TextStyle = fyne.TextStyle{Bold: true}
	setTypeBg := newSurface(ColorSurface)
	setTypeCard := container.NewStack(setTypeBg, container.NewPadded(setTypeTxt))
	setTypeCaption := canvas.NewText("MODO", ColorMuted)
	setTypeCaption.TextSize = 10
	setTypeCaption.Alignment = fyne.TextAlignCenter
	topCenter := container.NewVBox(
		container.NewCenter(setTypeCaption),
		container.NewCenter(setTypeCard),
	)

	// VS circle
	vsTxt := canvas.NewText("VS", ColorMuted)
	vsTxt.TextSize = 13
	vsTxt.TextStyle = fyne.TextStyle{Bold: true}
	vsTxt.Alignment = fyne.TextAlignCenter
	vsCircle := canvas.NewCircle(ColorSurface)
	vsCircle.StrokeColor = ColorStroke
	vsCircle.StrokeWidth = 1
	vsWidget := container.NewGridWrap(fyne.NewSize(40, 40),
		container.NewStack(vsCircle, container.NewCenter(vsTxt)),
	)

	undoBtn := NewSmallButton("↺ Deshacer", func() { a.undoLastGame() })
	finalizarBtn := NewSmallButton("✓ Finalizar", func() { a.confirmFinalizeLibreSet() })
	cancelBtn := NewSmallButton("✕ Cancelar", func() {
		a.confirmDialog(
			"Cancelar set",
			"¿Seguro? El set no se guardará en el historial.",
			func() {
				a.db.AbandonActiveSet()
				a.activeSet = nil
				a.showNewSet()
			},
		)
	})

	divLine := canvas.NewRectangle(ColorStroke)
	centerDiv := container.NewBorder(
		nil,
		container.NewVBox(undoBtn, finalizarBtn, cancelBtn),
		nil, nil,
		container.NewStack(
			container.NewCenter(divLine),
			container.NewCenter(vsWidget),
		),
	)

	mainSplit := container.New(&threeColLayout{centerW: 150}, leftPanel, centerDiv, rightPanel)

	return container.NewStack(
		canvas.NewRectangle(ColorBG),
		container.NewStack(
			mainSplit,
			container.NewBorder(container.NewCenter(topCenter), nil, nil, nil),
		),
	)
}

func (a *App) confirmFinalizeLibreSet() {
	s := a.activeSet
	if s.UserScore == s.OppScore {
		a.showTiebreakDialog()
		return
	}
	// El ganador se determina por score; la pantalla de victoria llama a CompleteSet
	a.showVictory()
}

func (a *App) showTiebreakDialog() {
	s := a.activeSet

	var dlg dialog.Dialog

	userBtn := widget.NewButton(fmt.Sprintf("🏆  %s", a.profile.Name), func() {
		dlg.Hide()
		s.UserScore++ // romper empate en memoria para que Winner() funcione
		a.showVictory()
	})
	userBtn.Importance = widget.HighImportance

	oppBtn := widget.NewButton(fmt.Sprintf("🏆  %s", s.OpponentName), func() {
		dlg.Hide()
		s.OppScore++ // romper empate en memoria para que Winner() funcione
		a.showVictory()
	})
	oppBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("El set terminó %d — %d (empate). ¿Quién ganó?", s.UserScore, s.OppScore)),
		widget.NewSeparator(),
		container.NewGridWithColumns(2, userBtn, oppBtn),
	)
	dlg = dialog.NewCustomWithoutButtons("Empate", content, a.win)
	dlg.Show()
}

func (a *App) buildGamePips(s *ActiveSetState, forUser bool, col color.NRGBA) fyne.CanvasObject {
	needed := s.WinsNeeded()

	// Count only this player's wins
	wins := 0
	for _, g := range s.Games {
		if (g == "user") == forUser {
			wins++
		}
	}
	if wins > needed {
		wins = needed
	}

	row := container.NewHBox()
	for i := 0; i < needed; i++ {
		var dot *canvas.Circle
		if i < wins {
			dot = canvas.NewCircle(col) // victoria
		} else {
			dot = canvas.NewCircle(color.Transparent) // pendiente
			dot.StrokeColor = ColorDim
			dot.StrokeWidth = 1.5
		}
		pip := container.NewGridWrap(fyne.NewSize(22, 22), container.NewStack(dot))
		row.Add(pip)
	}
	return row
}

func (a *App) recordGame(winner string) {
	s := a.activeSet
	if winner == "user" {
		s.UserScore++
	} else {
		s.OppScore++
	}
	s.Games = append(s.Games, winner)

	if err := a.db.RecordGame(s, winner); err != nil {
		a.showError("Error al guardar juego")
		return
	}

	if s.IsFinished() {
		a.showVictory()
	} else {
		func() {
			defer func() {
				if r := recover(); r != nil {
					a.showError(fmt.Sprintf("Error al actualizar pantalla: %v", r))
				}
			}()
			a.setBody(a.buildActiveSetScreen())
		}()
	}
}

func (a *App) undoLastGame() {
	s := a.activeSet
	if len(s.Games) == 0 {
		return
	}
	last := s.Games[len(s.Games)-1]
	if last == "user" {
		s.UserScore--
	} else {
		s.OppScore--
	}
	s.Games = s.Games[:len(s.Games)-1]

	if err := a.db.UndoLastGame(s); err != nil {
		a.showError("Error al deshacer")
		return
	}
	a.setBody(a.buildActiveSetScreen())
}

func (a *App) showChangeSetTypeDialog(setTypeTxt *canvas.Text) {
	s := a.activeSet
	selType := s.SetType
	selNum := s.SetNumber

	toggle := NewSetTypeToggle(&selType, nil)
	stepper := NewStepper(&selNum, 1, 15, nil)

	content := container.NewVBox(
		widget.NewLabel("El puntaje actual se mantiene."),
		toggle,
		stepper,
	)

	dialog.ShowCustomConfirm(
		"Cambiar tipo de set", "Confirmar", "Cancelar",
		content,
		func(ok bool) {
			if !ok {
				return
			}
			s.SetType = selType
			s.SetNumber = selNum
			if err := a.db.UpdateSetType(s, selType, selNum); err != nil {
				a.showError("Error al actualizar")
				return
			}
			setTypeTxt.Text = fmt.Sprintf("%s %d", s.SetType, s.SetNumber)
			setTypeTxt.Refresh()
			a.setBody(a.buildActiveSetScreen())
		},
		a.win,
	)
}

// threeColLayout: gives left/right equal space, center a fixed width
// minH overrides the default 300px minimum height (0 = use default)
type threeColLayout struct {
	centerW float32
	minH    float32
}

func (l *threeColLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	if len(objs) < 3 {
		return
	}
	sideW := (size.Width - l.centerW) / 2
	objs[0].Resize(fyne.NewSize(sideW, size.Height))
	objs[0].Move(fyne.NewPos(0, 0))
	objs[1].Resize(fyne.NewSize(l.centerW, size.Height))
	objs[1].Move(fyne.NewPos(sideW, 0))
	objs[2].Resize(fyne.NewSize(sideW, size.Height))
	objs[2].Move(fyne.NewPos(sideW+l.centerW, 0))
}

func (l *threeColLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	minH := l.minH
	if minH <= 0 {
		minH = 300
	}
	return fyne.NewSize(l.centerW+300, minH)
}
