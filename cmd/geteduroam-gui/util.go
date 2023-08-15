package main

import (
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/jwijenbergh/puregotk/v4/glib"
)

type StyledWidget interface{
	GetStyleContext() *gtk.StyleContext
}

func styleWidget(widget StyledWidget, css string) {
	provider := gtk.NewCssProvider()
	provider.LoadFromPath(css)
	sc := widget.GetStyleContext()
	// TODO: This casts the css style context to the underlying pointer
	// we should have to do this and this has to be fixed in puregotk
	sc.AddProvider(provider, 800)
}

func uiThread(cb func()) {
	glib.IdleAdd(func(uintptr) bool {
		cb()

		// return false here means just run it once, not over and over again
		// see the docs for glib_idle_add
		return false
	}, 0)
}
