package main

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) showH2H() {
	a.setActiveTab(a.tabH2H)
	a.setBody(a.buildH2HScreen(""))
}

func (a *App) buildH2HScreen(selectedOpponent string) fyne.CanvasObject {
	opponents, _ := a.db.LoadOpponents()

	titleLabel := canvas.NewText("ESTADÍSTICAS CONTRA", ColorMuted)
	titleLabel.TextSize = 11
	h2hTitle := canvas.NewText("Head — to — Head", ColorText)
	h2hTitle.TextSize = 40
	h2hTitle.TextStyle = fyne.TextStyle{Bold: true}
	titleBlock := container.NewVBox(titleLabel, h2hTitle)

	if len(opponents) == 0 {
		empty := canvas.NewText("Aún no hay sets registrados", ColorMuted)
		empty.TextSize = 16
		empty.Alignment = fyne.TextAlignCenter
		return container.NewStack(
			canvas.NewRectangle(ColorBG),
			container.NewBorder(
				container.NewPadded(container.NewPadded(titleBlock)),
				nil, nil, nil,
				container.NewCenter(empty),
			),
		)
	}

	if selectedOpponent == "" {
		selectedOpponent = opponents[0]
	}

	oppSelect := widget.NewSelect(opponents, func(s string) {
		a.setBody(a.buildH2HScreen(s))
	})
	// Asignar directamente sin disparar onChange (evita loop infinito)
	oppSelect.Selected = selectedOpponent
	oppSelect.Refresh()

	rivalLabel := canvas.NewText("Rival:", ColorMuted)
	rivalLabel.TextSize = 13
	headerRow := container.NewBorder(nil, nil, titleBlock,
		container.NewHBox(rivalLabel, oppSelect),
	)

	// Load data
	records, _ := a.db.LoadH2H(selectedOpponent)
	total := len(records)
	wins, losses := 0, 0
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

	oppColorHex := a.db.LoadOpponentColor(selectedOpponent)
	oppCol := ParseHexColor(oppColorHex)
	userCol := ParseHexColor(a.profile.Color)

	vsPanel := buildVSPanel(a.profile.Name, userCol, a.profile.PhotoPath, selectedOpponent, oppCol, wins, losses, total)
	winratePanel := buildWinrateBar(userCol, wr)
	formPanel := buildRecentForm(records)
	statsRow := container.NewGridWithColumns(2, winratePanel, formPanel)

	histLabel := canvas.NewText(fmt.Sprintf("Sets contra %s", selectedOpponent), ColorMuted)
	histLabel.TextSize = 11

	var histContent fyne.CanvasObject
	if total == 0 {
		none := canvas.NewText("Sin sets registrados", ColorMuted)
		none.Alignment = fyne.TextAlignCenter
		histContent = container.NewCenter(none)
	} else {
		list := widget.NewList(
			func() int { return total },
			func() fyne.CanvasObject { return newHistoryRow() },
			func(id widget.ListItemID, obj fyne.CanvasObject) {
				updateHistoryRow(obj, records[id], a.profile.Name, nil)
			},
		)
		list.OnSelected = func(_ widget.ListItemID) { list.UnselectAll() }
		histContent = list
	}

	// Panel VS + stats como sección fija arriba; lista llena el resto
	fixedTop := container.NewVBox(
		container.NewPadded(container.NewPadded(headerRow)),
		container.NewPadded(vsPanel),
		container.NewPadded(statsRow),
		widget.NewSeparator(),
		container.NewPadded(histLabel),
	)

	return container.NewStack(
		canvas.NewRectangle(ColorBG),
		container.NewBorder(
			fixedTop,
			nil, nil, nil,
			container.NewPadded(histContent),
		),
	)
}

func buildVSPanel(userName string, userCol color.NRGBA, userPhotoPath string, oppName string, oppCol color.NRGBA, wins, losses, total int) fyne.CanvasObject {
	// User side
	userAvatar := buildAvatarWidget(userName, userCol, 56, userPhotoPath)
	userNameLbl := canvas.NewText(strings.ToUpper(userName), userCol)
	userNameLbl.TextSize = 28
	userNameLbl.TextStyle = fyne.TextStyle{Bold: true}
	userSubLbl := canvas.NewText(fmt.Sprintf("%d VICTORIAS", wins), ColorMuted)
	userSubLbl.TextSize = 11
	userSide := container.NewHBox(
		userAvatar,
		container.NewVBox(canvas.NewText("TÚ", ColorMuted), userNameLbl, userSubLbl),
	)

	// Center
	totalLbl := canvas.NewText(fmt.Sprintf("%d sets jugados", total), ColorMuted)
	totalLbl.TextSize = 10
	totalLbl.Alignment = fyne.TextAlignCenter

	winTxt := canvas.NewText(fmt.Sprintf("%d", wins), ColorText)
	winTxt.TextSize = 80
	winTxt.TextStyle = fyne.TextStyle{Bold: true}
	dashTxt := canvas.NewText("—", ColorDim)
	dashTxt.TextSize = 56
	loseTxt := canvas.NewText(fmt.Sprintf("%d", losses), ColorText)
	loseTxt.TextSize = 80
	loseTxt.TextStyle = fyne.TextStyle{Bold: true}
	scoreRow := container.NewHBox(winTxt, dashTxt, loseTxt)
	centerBlock := container.NewVBox(
		container.NewCenter(totalLbl),
		container.NewCenter(scoreRow),
	)

	// Opponent side
	oppAvatar := buildAvatarWidget(oppName, oppCol, 56, "")
	oppNameLbl := canvas.NewText(strings.ToUpper(oppName), oppCol)
	oppNameLbl.TextSize = 28
	oppNameLbl.TextStyle = fyne.TextStyle{Bold: true}
	oppSubLbl := canvas.NewText(fmt.Sprintf("%d VICTORIAS", losses), ColorMuted)
	oppSubLbl.TextSize = 11
	oppSide := container.NewHBox(
		container.NewVBox(canvas.NewText("ADVERSARIO", ColorMuted), oppNameLbl, oppSubLbl),
		oppAvatar,
	)

	panelBg := canvas.NewRectangle(ColorSurface)
	panelBg.StrokeColor = ColorStroke
	panelBg.StrokeWidth = 1
	panelBg.CornerRadius = 3

	inner := container.New(&threeColLayout{centerW: 260, minH: 140}, userSide, centerBlock, oppSide)
	return container.NewStack(panelBg, container.NewPadded(container.NewPadded(inner)))
}

func buildWinrateBar(userCol color.NRGBA, wr int) fyne.CanvasObject {
	label := canvas.NewText("TASA DE VICTORIA", ColorMuted)
	label.TextSize = 11
	pctLbl := canvas.NewText(fmt.Sprintf("%d%%  ·  %d%%", wr, 100-wr), ColorMuted)
	pctLbl.TextSize = 11
	pctLbl.Alignment = fyne.TextAlignTrailing
	labelRow := container.NewBorder(nil, nil, label, pctLbl)

	userBar := canvas.NewRectangle(userCol)
	oppBar := canvas.NewRectangle(ColorOpponent)
	barBg := canvas.NewRectangle(ColorSurface2)
	bar := container.New(&splitBarLayout{fraction: float32(wr) / 100.0}, barBg, userBar, oppBar)

	bg := canvas.NewRectangle(ColorSurface)
	bg.StrokeColor = ColorStroke
	bg.StrokeWidth = 1
	bg.CornerRadius = 3

	return container.NewStack(bg, container.NewPadded(container.NewPadded(
		container.NewVBox(labelRow, bar),
	)))
}

func buildRecentForm(records []SetRecord) fyne.CanvasObject {
	label := canvas.NewText("FORMA RECIENTE", ColorMuted)
	label.TextSize = 11
	pips := container.NewHBox()
	max := 7
	if len(records) < max {
		max = len(records)
	}
	for i := 0; i < max; i++ {
		r := records[i]
		won := r.UserWon != nil && *r.UserWon
		col := ColorOpponent
		letter := "L"
		if won {
			col = ColorUser
			letter = "W"
		}
		dot := canvas.NewRectangle(col)
		dot.CornerRadius = 2
		lbl := canvas.NewText(letter, ColorBG)
		lbl.TextSize = 11
		lbl.TextStyle = fyne.TextStyle{Bold: true}
		lbl.Alignment = fyne.TextAlignCenter
		pip := container.NewGridWrap(fyne.NewSize(24, 24),
			container.NewStack(dot, container.NewCenter(lbl)),
		)
		pips.Add(pip)
	}

	bg := canvas.NewRectangle(ColorSurface)
	bg.StrokeColor = ColorStroke
	bg.StrokeWidth = 1
	bg.CornerRadius = 3
	return container.NewStack(bg, container.NewPadded(container.NewPadded(
		container.NewBorder(nil, nil, label, pips),
	)))
}

func buildAvatarWidget(name string, col color.NRGBA, size float32, photoPath string) fyne.CanvasObject {
	if photoPath != "" {
		img := canvas.NewImageFromFile(photoPath)
		img.FillMode = canvas.ImageFillContain
		return container.NewGridWrap(fyne.NewSize(size, size), img)
	}
	initial := "?"
	if len(name) > 0 {
		initial = string([]rune(name)[0])
	}
	circle := canvas.NewCircle(col)
	circle.Resize(fyne.NewSize(size, size))
	lbl := canvas.NewText(strings.ToUpper(initial), ColorBG)
	lbl.TextSize = size * 0.4
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	lbl.Alignment = fyne.TextAlignCenter
	return container.NewGridWrap(fyne.NewSize(size, size),
		container.NewStack(circle, container.NewCenter(lbl)),
	)
}

// splitBarLayout
type splitBarLayout struct {
	fraction float32
}

func (l *splitBarLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	if len(objs) < 3 {
		return
	}
	objs[0].Resize(size)
	objs[0].Move(fyne.NewPos(0, 0))
	uw := size.Width * l.fraction
	objs[1].Resize(fyne.NewSize(uw, size.Height))
	objs[1].Move(fyne.NewPos(0, 0))
	objs[2].Resize(fyne.NewSize(size.Width-uw, size.Height))
	objs[2].Move(fyne.NewPos(uw, 0))
}

func (l *splitBarLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(100, 16)
}
