package main

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (a *App) showHistory() {
	a.setActiveTab(a.tabHist)
	a.setBody(a.buildHistoryScreen())
}

func (a *App) buildHistoryScreen() fyne.CanvasObject {
	records, err := a.db.LoadHistory()
	if err != nil {
		return widget.NewLabel("Error cargando historial: " + err.Error())
	}

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

	// --- Header ---
	titleLabel := canvas.NewText("REGISTRO COMPLETO", ColorMuted)
	titleLabel.TextSize = 11
	histTitle := canvas.NewText("Historial", ColorText)
	histTitle.TextSize = 48
	histTitle.TextStyle = fyne.TextStyle{Bold: true}
	titleBlock := container.NewVBox(titleLabel, histTitle)

	statsRow := container.NewHBox(
		buildStatWidget("Total", fmt.Sprintf("%d", total), ColorText),
		buildStatWidget("Victorias", fmt.Sprintf("%d", wins), ColorUser),
		buildStatWidget("Derrotas", fmt.Sprintf("%d", losses), ColorOpponent),
		buildStatWidget("Winrate", fmt.Sprintf("%d%%", wr), ColorText),
	)
	headerRow := container.NewBorder(nil, nil, titleBlock, statsRow)

	if total == 0 {
		empty := canvas.NewText("Aún no hay sets registrados", ColorMuted)
		empty.TextSize = 16
		empty.Alignment = fyne.TextAlignCenter
		return container.NewStack(
			canvas.NewRectangle(ColorBG),
			container.NewBorder(
				container.NewPadded(container.NewPadded(headerRow)),
				nil, nil, nil,
				container.NewCenter(empty),
			),
		)
	}

	// --- Filtro dinámico ---
	filteredRecords := make([]SetRecord, len(records))
	copy(filteredRecords, records)

	var list *widget.List

	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("Filtrar por rival...")
	filterEntry.OnChanged = func(query string) {
		query = strings.ToLower(strings.TrimSpace(query))
		if query == "" {
			filteredRecords = records
		} else {
			filteredRecords = nil
			for _, r := range records {
				if strings.Contains(strings.ToLower(r.OpponentName), query) {
					filteredRecords = append(filteredRecords, r)
				}
			}
		}
		if list != nil {
			list.Refresh()
		}
	}

	// --- Acciones de borrado ---
	oppNames := []string{}
	seen := map[string]bool{}
	for _, r := range records {
		if !seen[r.OpponentName] {
			oppNames = append(oppNames, r.OpponentName)
			seen[r.OpponentName] = true
		}
	}
	selectedDelOpp := oppNames[0]

	delOppSelect := widget.NewSelect(oppNames, func(s string) {
		selectedDelOpp = s
	})
	delOppSelect.Selected = selectedDelOpp
	delOppSelect.Refresh()

	delOppBtn := widget.NewButton("Borrar rival", func() {
		a.confirmDialog(
			"Borrar rival",
			fmt.Sprintf("¿Eliminar todos los sets contra %s?\nEsta acción no se puede deshacer.", selectedDelOpp),
			func() {
				a.db.DeleteHistoryByOpponent(selectedDelOpp)
				a.showHistory()
			},
		)
	})
	delOppBtn.Importance = widget.DangerImportance

	delAllBtn := widget.NewButton("Borrar todo", func() {
		a.confirmDialog(
			"Borrar historial completo",
			"¿Eliminar TODOS los sets registrados?\nEsta acción no se puede deshacer.",
			func() {
				a.db.DeleteAllHistory()
				a.showHistory()
			},
		)
	})
	delAllBtn.Importance = widget.DangerImportance

	rivalLbl := widget.NewLabel("Rival:")
	actionsBar := container.NewBorder(nil, nil,
		container.NewHBox(rivalLbl, delOppSelect, delOppBtn),
		delAllBtn,
	)

	// --- Lista ---
	list = widget.NewList(
		func() int { return len(filteredRecords) },
		func() fyne.CanvasObject { return newHistoryRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			r := filteredRecords[id]
			updateHistoryRow(obj, r, a.profile.Name, func() {
				a.confirmDialog(
					"Eliminar set",
					fmt.Sprintf("¿Eliminar el set contra %s (%d–%d)?", r.OpponentName, r.UserScore, r.OpponentScore),
					func() {
						a.db.DeleteSet(r.ID)
						a.showHistory()
					},
				)
			})
		},
	)
	list.OnSelected = func(_ widget.ListItemID) { list.UnselectAll() }

	topSection := container.NewVBox(
		container.NewPadded(container.NewPadded(headerRow)),
		widget.NewSeparator(),
		container.NewPadded(filterEntry),
		widget.NewSeparator(),
		container.NewPadded(actionsBar),
		widget.NewSeparator(),
	)

	return container.NewStack(
		canvas.NewRectangle(ColorBG),
		container.NewBorder(
			topSection,
			nil, nil, nil,
			container.NewPadded(list),
		),
	)
}

func buildStatWidget(label, value string, col color.NRGBA) fyne.CanvasObject {
	lbl := canvas.NewText(strings.ToUpper(label), ColorMuted)
	lbl.TextSize = 10
	lbl.Alignment = fyne.TextAlignTrailing
	val := canvas.NewText(value, col)
	val.TextSize = 28
	val.TextStyle = fyne.TextStyle{Bold: true}
	val.Alignment = fyne.TextAlignTrailing
	return container.NewVBox(lbl, val)
}

// historyRowData carries all mutable canvas objects in a reusable list row
type historyRowData struct {
	accentBar    *canvas.Rectangle
	dateTxt      *canvas.Text
	userTxt      *canvas.Text
	oppTxt       *canvas.Text
	typeTxt      *canvas.Text
	userScoreTxt *canvas.Text
	oppScoreTxt  *canvas.Text
	resultTxt    *canvas.Text
	deleteBtn    *widget.Button
}

func newHistoryRow() fyne.CanvasObject {
	d := &historyRowData{}

	d.accentBar = canvas.NewRectangle(ColorDim)
	bg := canvas.NewRectangle(ColorSurface)
	bg.StrokeColor = ColorStroke
	bg.StrokeWidth = 1
	bg.CornerRadius = 2

	d.dateTxt = canvas.NewText("00 Xxx · 00:00", ColorMuted)
	d.dateTxt.TextSize = 11

	d.userTxt = canvas.NewText("USER", ColorUser)
	d.userTxt.TextSize = 18
	d.userTxt.TextStyle = fyne.TextStyle{Bold: true}

	vsTxt := canvas.NewText("VS", ColorDim)
	vsTxt.TextSize = 12

	d.oppTxt = canvas.NewText("OPP", ColorOpponent)
	d.oppTxt.TextSize = 18
	d.oppTxt.TextStyle = fyne.TextStyle{Bold: true}

	d.typeTxt = canvas.NewText("FT5", ColorMuted)
	d.typeTxt.TextSize = 11

	d.userScoreTxt = canvas.NewText("0", ColorUser)
	d.userScoreTxt.TextSize = 30
	d.userScoreTxt.TextStyle = fyne.TextStyle{Bold: true}

	dashTxt := canvas.NewText("—", ColorDim)
	dashTxt.TextSize = 22

	d.oppScoreTxt = canvas.NewText("0", ColorMuted)
	d.oppScoreTxt.TextSize = 30
	d.oppScoreTxt.TextStyle = fyne.TextStyle{Bold: true}

	d.resultTxt = canvas.NewText("▲ Victoria", ColorUser)
	d.resultTxt.TextSize = 12
	d.resultTxt.TextStyle = fyne.TextStyle{Bold: true}
	d.resultTxt.Alignment = fyne.TextAlignTrailing

	// Tacho de basura al extremo derecho
	d.deleteBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
	d.deleteBtn.Importance = widget.DangerImportance

	matchRow := container.NewHBox(d.userTxt, vsTxt, d.oppTxt)
	scoreRow := container.NewHBox(d.userScoreTxt, dashTxt, d.oppScoreTxt)

	cols := container.New(&histRowLayout{},
		d.dateTxt, matchRow, d.typeTxt, scoreRow, d.resultTxt,
	)

	// accent bar izquierda | contenido | tacho derecha
	accentWrapper := container.NewGridWrap(fyne.NewSize(4, 52), d.accentBar)

	row := container.NewStack(
		bg,
		container.NewBorder(nil, nil, accentWrapper, container.NewCenter(d.deleteBtn),
			container.NewPadded(cols),
		),
	)

	marker := widget.NewLabel("")
	marker.Hide()
	row.Add(marker)
	histRowStore[marker] = d

	return row
}

var histRowStore = map[*widget.Label]*historyRowData{}

func updateHistoryRow(obj fyne.CanvasObject, r SetRecord, userName string, onDelete func()) {
	stack := obj.(*fyne.Container)
	var marker *widget.Label
	for _, o := range stack.Objects {
		if lbl, ok := o.(*widget.Label); ok {
			marker = lbl
			break
		}
	}
	if marker == nil {
		return
	}
	d := histRowStore[marker]
	if d == nil {
		return
	}

	won := r.UserWon != nil && *r.UserWon
	if won {
		d.accentBar.FillColor = ColorUser
	} else {
		d.accentBar.FillColor = ColorOpponent
	}
	d.accentBar.Refresh()

	d.dateTxt.Text = r.PlayedAt.Format("02 Jan · 15:04")
	d.dateTxt.Refresh()

	d.userTxt.Text = strings.ToUpper(userName)
	d.userTxt.Refresh()

	oppCol := ParseHexColor(r.OpponentColor)
	d.oppTxt.Text = strings.ToUpper(r.OpponentName)
	d.oppTxt.Color = oppCol
	d.oppTxt.Refresh()

	d.typeTxt.Text = fmt.Sprintf("%s%d", r.SetType, r.SetNumber)
	d.typeTxt.Refresh()

	d.userScoreTxt.Text = fmt.Sprintf("%d", r.UserScore)
	d.userScoreTxt.Refresh()

	d.oppScoreTxt.Text = fmt.Sprintf("%d", r.OpponentScore)
	d.oppScoreTxt.Color = oppCol
	d.oppScoreTxt.Refresh()

	if won {
		d.resultTxt.Text = "▲ Victoria"
		d.resultTxt.Color = ColorUser
	} else {
		d.resultTxt.Text = "▼ Derrota"
		d.resultTxt.Color = ColorOpponent
	}
	d.resultTxt.Refresh()

	d.deleteBtn.OnTapped = onDelete
	if onDelete != nil {
		d.deleteBtn.Show()
	} else {
		d.deleteBtn.Hide()
	}
	d.deleteBtn.Refresh()
}

type histRowLayout struct{}

func (l *histRowLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	if len(objs) < 5 {
		return
	}
	h := size.Height
	col0W := float32(130) // fecha
	col2W := float32(60)  // tipo
	col3W := float32(110) // marcador
	col4W := float32(90)  // resultado
	col1W := size.Width - col0W - col2W - col3W - col4W - 40
	if col1W < 80 {
		col1W = 80
	}
	x := float32(0)
	for i, w := range []float32{col0W, col1W, col2W, col3W, col4W} {
		objs[i].Resize(fyne.NewSize(w, h))
		objs[i].Move(fyne.NewPos(x, 0))
		x += w + 8
	}
}

func (l *histRowLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(600, 52)
}
