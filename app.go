package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	fyneApp   fyne.App
	win       fyne.Window
	db        *DB
	profile   *UserProfile
	activeSet *ActiveSetState

	// header refs updated on nav change
	tabSet  *widget.Button
	tabHist *widget.Button
	tabH2H  *widget.Button

	content *fyne.Container // border container holding header + current screen
	body    *fyne.Container // max container swapped per screen

	anims      []*fyne.Animation
	lastWinner string
}

func newApp(a fyne.App, w fyne.Window, db *DB) *App {
	return &App{fyneApp: a, win: w, db: db}
}

func (a *App) start() {
	hasProfile, _ := a.db.HasProfile()
	if !hasProfile {
		a.showOnboarding()
		return
	}
	p, err := a.db.LoadProfile()
	if err != nil || p == nil {
		a.showOnboarding()
		return
	}
	a.profile = p

	// Check for active set
	active, _ := a.db.LoadActiveSet()
	a.activeSet = active

	a.buildMain()

	if active != nil {
		// Si el set ya tiene ganador (quedó en estado corrupto tras un crash), lo completamos
		if active.IsFinished() {
			userWon := active.Winner() == "user"
			a.db.CompleteSet(active, userWon)
			a.activeSet = nil
			a.showNewSet()
		} else {
			// Renderizamos con recover por si el estado en DB está corrupto
			func() {
				defer func() {
					if r := recover(); r != nil {
						a.db.AbandonActiveSet()
						a.activeSet = nil
						a.showNewSet()
					}
				}()
				a.showActiveSet()
			}()
		}
	} else {
		a.showNewSet()
	}
}

// buildMain creates the persistent layout (header + swappable body)
func (a *App) buildMain() {
	header := a.buildHeader()
	a.body = container.NewMax()
	a.content = container.NewBorder(header, nil, nil, nil, a.body)
	a.win.SetContent(a.content)
}

func (a *App) setBody(obj fyne.CanvasObject) {
	a.stopAnims()
	a.body.Objects = []fyne.CanvasObject{obj}
	a.body.Refresh()
}

func (a *App) trackAnim(anim *fyne.Animation) {
	a.anims = append(a.anims, anim)
}

func (a *App) stopAnims() {
	for _, anim := range a.anims {
		anim.Stop()
	}
	a.anims = nil
}

// buildHeader creates the top navigation bar
func (a *App) buildHeader() fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorSurface)
	border := canvas.NewRectangle(ColorStroke)
	border.FillColor = ColorStroke

	logo := a.buildLogo()

	a.tabSet = widget.NewButton("Set Activo", func() { a.showActiveSetOrNew() })
	a.tabHist = widget.NewButton("Historial", func() { a.showHistory() })
	a.tabH2H = widget.NewButton("Head-to-Head", func() { a.showH2H() })

	a.styleTab(a.tabSet, true)
	a.styleTab(a.tabHist, false)
	a.styleTab(a.tabH2H, false)

	nav := container.NewHBox(logo, a.tabSet, a.tabHist, a.tabH2H)

	profileWidget := a.buildProfileWidget()
	topBar := container.NewBorder(nil, nil, nav, profileWidget)

	headerH := float32(56)
	bg.SetMinSize(fyne.NewSize(0, headerH))

	return container.NewStack(
		bg,
		container.NewPadded(topBar),
	)
}

func (a *App) styleTab(btn *widget.Button, active bool) {
	if active {
		btn.Importance = widget.HighImportance
	} else {
		btn.Importance = widget.LowImportance
	}
}

func (a *App) setActiveTab(active *widget.Button) {
	for _, t := range []*widget.Button{a.tabSet, a.tabHist, a.tabH2H} {
		a.styleTab(t, t == active)
		t.Refresh()
	}
}

func (a *App) buildLogo() fyne.CanvasObject {
	dot := canvas.NewRectangle(ColorUser)
	dot.SetMinSize(fyne.NewSize(10, 10))

	name := canvas.NewText("BEAR·TRACKER", ColorText)
	name.TextSize = 18
	name.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewHBox(dot, name)
}

func (a *App) buildProfileWidget() fyne.CanvasObject {
	col := ColorUser
	if a.profile != nil {
		col = ParseHexColor(a.profile.Color)
	}

	var avatarContent fyne.CanvasObject
	if a.profile != nil && a.profile.PhotoPath != "" {
		img := canvas.NewImageFromFile(a.profile.PhotoPath)
		img.FillMode = canvas.ImageFillContain
		avatarContent = container.NewGridWrap(fyne.NewSize(32, 32), img)
	} else {
		initial := "?"
		if a.profile != nil && len(a.profile.Name) > 0 {
			initial = string([]rune(a.profile.Name)[0])
		}
		avatar := canvas.NewText(initial, ColorText)
		avatar.TextSize = 16
		avatar.TextStyle = fyne.TextStyle{Bold: true}
		avatarBg := canvas.NewCircle(col)
		avatarBg.Resize(fyne.NewSize(32, 32))
		avatarContent = container.NewGridWrap(fyne.NewSize(32, 32),
			container.NewStack(container.NewCenter(avatarBg), container.NewCenter(avatar)),
		)
	}

	name := widget.NewLabel("")
	if a.profile != nil {
		name.SetText(a.profile.Name)
	}
	name.TextStyle = fyne.TextStyle{Bold: true}

	profileBtn := widget.NewButton("", func() {
		a.showProfileDialog()
	})
	profileBtn.Importance = widget.LowImportance

	return container.NewStack(
		profileBtn,
		container.NewHBox(avatarContent, name),
	)
}

func (a *App) refreshProfileWidget() {
	// Rebuild header to reflect profile changes
	a.buildMain()
	if a.activeSet != nil {
		a.showActiveSet()
	} else {
		a.showNewSet()
	}
}

func (a *App) showActiveSetOrNew() {
	a.setActiveTab(a.tabSet)
	if a.activeSet != nil {
		a.showActiveSet()
	} else {
		a.showNewSet()
	}
}

func (a *App) showProfileDialog() {
	var pop *widget.PopUp

	closeIt := func() {
		if pop != nil {
			pop.Hide()
		}
	}

	cardContent := buildProfileContent(a, closeIt)

	// Fondo opaco para que no se vea la pantalla detrás
	bg := canvas.NewRectangle(ColorSurface)
	bg.StrokeColor = ColorStroke
	bg.StrokeWidth = 1
	bg.CornerRadius = 3
	opaqueCard := container.NewStack(bg, container.NewPadded(cardContent))

	// tapAbsorber: evita que clicks dentro del card cierren el PopUp
	absorber := &tapAbsorber{content: opaqueCard}
	absorber.ExtendBaseWidget(absorber)

	pop = widget.NewPopUp(absorber, a.win.Canvas())

	cardSize := fyne.NewSize(560, 500)
	pop.Resize(cardSize)

	cs := a.win.Canvas().Size()
	pop.ShowAtPosition(fyne.NewPos(
		(cs.Width-cardSize.Width)/2,
		(cs.Height-cardSize.Height)/2,
	))
}

// confirmDialog shows a yes/no dialog and calls the callback on yes
func (a *App) confirmDialog(title, message string, onConfirm func()) {
	dialog.ShowConfirm(title, message, func(ok bool) {
		if ok {
			onConfirm()
		}
	}, a.win)
}

func (a *App) showError(msg string) {
	dialog.ShowError(fmt.Errorf("%s", msg), a.win)
}
