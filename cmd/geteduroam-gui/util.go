package main

import (
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"strings"
)

type StyledWidget interface {
	GetStyleContext() *gtk.StyleContext
}

func styleWidget(widget StyledWidget, resName string) {
	provider := gtk.NewCssProvider()
	provider.LoadFromData(MustResource(resName+".css"), -1)
	sc := widget.GetStyleContext()
	sc.AddProvider(provider, 800)
}

func upper(str string) string {
	return strings.ToUpper(str[:1])+str[1:]
}

func showErrorToast(overlay adw.ToastOverlay, err error) {
	msg := upper(err.Error())
	toast := adw.NewToast(glib.MarkupEscapeText(msg, -1))
	toast.SetTimeout(5)
	overlay.AddToast(toast)
}

func uiThread(cb func()) {
	glib.IdleAdd(func(uintptr) bool {
		cb()

		// return false here means just run it once, not over and over again
		// see the docs for glib_idle_add
		return false
	}, 0)
}
