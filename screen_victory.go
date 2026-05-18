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
	glowAnim := glowLoop(glow, winCol, 15, 35)
	a.trackAnim(glowAnim)
	glowAnim.Start()

	// Labels
	finalLbl := canvas.NewText("SET FINALIZADO", ColorMuted)
	finalLbl.TextSize = 12
	finalLbl.Alignment = fyne.TextAlignCenter

	headline := canvas.NewText("¡"+strings.ToUpper(winnerName)+" GANÓ!", winCol)
	headline.TextSize = 88
	headline.TextStyle = fyne.TextStyle{Bold: true}
	headline.Alignment = fyne.TextAlignCenter
	headlineAnim := fadeText(headline, 450*ms, easeOutF)
	after(200*ms, headlineAnim.Start)

	// Score row
	winScoreTxt := canvas.NewText("0", userCol)
	winScoreTxt.TextSize = 88
	winScoreTxt.TextStyle = fyne.TextStyle{Bold: true}
	if !userWon {
		winScoreTxt.Color = ColorMuted
	}

	dashTxt := canvas.NewText("—", ColorDim)
	dashTxt.TextSize = 60

	loseScoreTxt := canvas.NewText("0", oppCol)
	loseScoreTxt.TextSize = 88
	loseScoreTxt.TextStyle = fyne.TextStyle{Bold: true}
	if userWon {
		loseScoreTxt.Color = ColorMuted
	}
	after(300*ms, func() {
		countUp(func(n int) {
			winScoreTxt.Text = fmt.Sprintf("%d", n)
			winScoreTxt.Refresh()
		}, userScore, 800*ms).Start()
		countUp(func(n int) {
			loseScoreTxt.Text = fmt.Sprintf("%d", n)
			loseScoreTxt.Refresh()
		}, oppScore, 800*ms).Start()
	})

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
	trophyLbl.TextSize = 0
	trophyLbl.Alignment = fyne.TextAlignCenter
	fyne.NewAnimation(500*ms, func(t float32) {
		trophyLbl.TextSize = 64 * overshootF(t)
		trophyLbl.Refresh()
	}).Start()

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
	fadeRect(tintedBg, 10, 400*ms).Start()

	return container.NewStack(
		overlayBg,
		tintedBg,
		container.NewStack(
			container.NewCenter(container.NewGridWrap(fyne.NewSize(600, 600), container.NewCenter(glow))),
			container.NewCenter(content),
		),
	)
}
