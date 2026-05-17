package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) showVictory() {
	a.setBody(a.buildVictoryScreen())
}

func (a *App) buildVictoryScreen() fyne.CanvasObject {
	s := a.activeSet
	userWon := s.Winner() == "user"

	userCol := ParseHexColor(a.profile.Color)
	oppCol := ParseHexColor(s.OpponentColor)

	var winnerName string
	var winCol = userCol
	if userWon {
		winnerName = a.profile.Name
	} else {
		winnerName = s.OpponentName
		winCol = oppCol
	}

	userScore, oppScore := s.UserScore, s.OppScore

	// Background
	overlayBg := canvas.NewRectangle(ColorBG)

	// Radial glow (large tinted circle)
	glow := canvas.NewCircle(WithAlpha(winCol, 25))
	glow.Resize(fyne.NewSize(600, 600))

	// Labels
	finalLbl := canvas.NewText("SET FINALIZADO", ColorMuted)
	finalLbl.TextSize = 12
	finalLbl.Alignment = fyne.TextAlignCenter

	headline := canvas.NewText("¡"+strings.ToUpper(winnerName)+" GANÓ!", winCol)
	headline.TextSize = 88
	headline.TextStyle = fyne.TextStyle{Bold: true}
	headline.Alignment = fyne.TextAlignCenter

	// Score row
	winScoreTxt := canvas.NewText(fmt.Sprintf("%d", userScore), userCol)
	winScoreTxt.TextSize = 88
	winScoreTxt.TextStyle = fyne.TextStyle{Bold: true}
	if !userWon {
		winScoreTxt.Color = ColorMuted
	}

	dashTxt := canvas.NewText("—", ColorDim)
	dashTxt.TextSize = 60

	loseScoreTxt := canvas.NewText(fmt.Sprintf("%d", oppScore), oppCol)
	loseScoreTxt.TextSize = 88
	loseScoreTxt.TextStyle = fyne.TextStyle{Bold: true}
	if userWon {
		loseScoreTxt.Color = ColorMuted
	}

	scoreRow := container.NewCenter(container.NewHBox(winScoreTxt, dashTxt, loseScoreTxt))

	// Players line
	userLbl := canvas.NewText(strings.ToUpper(a.profile.Name), userCol)
	userLbl.TextSize = 14
	userLbl.TextStyle = fyne.TextStyle{Bold: true}

	vsDivTxt := canvas.NewText("VS", ColorDim)
	vsDivTxt.TextSize = 12

	oppLbl := canvas.NewText(strings.ToUpper(s.OpponentName), oppCol)
	oppLbl.TextSize = 14
	oppLbl.TextStyle = fyne.TextStyle{Bold: true}

	typeLbl := canvas.NewText(fmt.Sprintf("%s%d", s.SetType, s.SetNumber), ColorMuted)
	typeLbl.TextSize = 12

	playersRow := container.NewCenter(container.NewHBox(userLbl, vsDivTxt, oppLbl, typeLbl))

	// Trophy
	trophyLbl := canvas.NewText("🏆", ColorText)
	trophyLbl.TextSize = 64
	trophyLbl.Alignment = fyne.TextAlignCenter

	// Continue
	continueBtn := NewActionButton("Continuar →", winCol, func() {
		if err := a.db.CompleteSet(s, userWon); err != nil {
			a.showError("Error al guardar set: " + err.Error())
		}
		a.activeSet = nil
		a.showNewSet()
	})

	a.setActiveTab(a.tabSet)

	content := container.NewVBox(
		widget.NewSeparator(),
		container.NewCenter(trophyLbl),
		container.NewCenter(finalLbl),
		container.NewCenter(headline),
		scoreRow,
		playersRow,
		widget.NewSeparator(),
		container.NewCenter(continueBtn),
	)

	tintedBg := canvas.NewRectangle(WithAlpha(winCol, 10))

	return container.NewStack(
		overlayBg,
		tintedBg,
		container.NewStack(
			container.NewCenter(container.NewGridWrap(fyne.NewSize(600, 600), container.NewCenter(glow))),
			container.NewCenter(content),
		),
	)
}
