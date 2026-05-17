package main

import (
	_ "embed"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

//go:embed bear.png
var bearIconData []byte

func main() {
	a := app.NewWithID("com.beartracker.app")
	a.Settings().SetTheme(&appTheme{})

	bearIcon := fyne.NewStaticResource("bear.png", bearIconData)
	a.SetIcon(bearIcon)

	w := a.NewWindow("Bear Tracker")
	w.SetIcon(bearIcon)
	w.Resize(fyne.NewSize(1280, 800))
	w.SetMaster()
	w.CenterOnScreen()

	db, err := openDB()
	if err != nil {
		log.Fatalf("Error abriendo base de datos: %v", err)
	}

	application := newApp(a, w, db)
	application.start()

	w.ShowAndRun()
}
