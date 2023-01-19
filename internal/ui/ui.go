package ui

import (
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/glib"
	"log"
	"os"
)

type ui struct {}

func New() *ui {return &ui{}}


func (ui *ui) Run(args []string) int {
	gtk.Init(&args)

	const id = "com.geteduroam.linux"
	app, err := gtk.ApplicationNew(id, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatalln("failed to create application:", err)
	}

	app.Connect("activate", func() {
		builder, err := gtk.BuilderNewFromFile("resources/ui.glade")
		if err != nil {
			log.Fatalln("builder error:", err)
		}

		obj, err := builder.GetObject("mainWindow")
		if err != nil {
			log.Fatalln("failed to get main window error:", err)
		}
		wnd := obj.(*gtk.Window)

		wnd.ShowAll()
		app.AddWindow(wnd)
	})
	return app.Run(os.Args)
}
