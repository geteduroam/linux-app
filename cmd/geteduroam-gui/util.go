package main

import (
	"context"
	"os"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gdkpixbuf"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
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

func setPage(stack *adw.ViewStack, page *adw.ViewStackPage) {
	child := page.GetChild()
	child.SetMarginStart(10)
	child.SetMarginEnd(10)
	child.SetMarginTop(5)
	child.SetMarginBottom(5)
	stack.SetVisibleChild(child)
}

func upper(str string) string {
	return strings.ToUpper(str[:1]) + str[1:]
}

func showErrorToast(overlay adw.ToastOverlay, err error) {
	msg := upper(err.Error())
	toast := adw.NewToast(glib.MarkupEscapeText(msg, -1))
	toast.SetTimeout(5)
	overlay.AddToast(toast)
}

func bytesPixbuf(b []byte) (*gdkpixbuf.Pixbuf, error) {
	// TODO: do this without creating a temp file
	f, err := os.CreateTemp("/tmp", "geteduroam-pixbuf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(b)
	if err != nil {
		return nil, err
	}
	pb, err := gdkpixbuf.NewPixbufFromFile(f.Name())
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func uiThread(cb func()) {
	var idlecb glib.SourceFunc
	idlecb = func(uintptr) bool {
		// unref so this callback does not take up any slots
		defer glib.UnrefCallback(&idlecb) //nolint:errcheck
		cb()

		// return false here means just run it once, not over and over again
		// see the docs for glib_idle_add
		return false
	}
	glib.IdleAdd(&idlecb, 0)
}

func ensureContextError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	select{
	case <- ctx.Done():
		return context.Canceled
	default:
		return err
	}
}
