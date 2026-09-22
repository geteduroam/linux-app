package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/geteduroam/linux-app/internal/variant"
)

type StyledWidget interface {
	StyleContext() *gtk.StyleContext
}

func styleWidget(widget StyledWidget, resName string) {
	provider := gtk.NewCSSProvider()
	provider.LoadFromData(MustResource(resName + ".css"))
	sc := widget.StyleContext()
	sc.AddProvider(provider, 800)
}

func setPage(stack *adw.ViewStack, page *adw.ViewStackPage) {
	child := page.Child()
	child.SetObjectProperty("margin-start", 10)
	child.SetObjectProperty("margin-end", 10)
	child.SetObjectProperty("margin-top", 5)
	child.SetObjectProperty("margin-bottom", 5)
	stack.SetVisibleChild(child)
}

func upper(str string) string {
	return strings.ToUpper(str[:1]) + str[1:]
}

func showErrorToast(overlay *adw.ToastOverlay, err error) {
	msg := upper(err.Error())
	toast := adw.NewToast(glib.MarkupEscapeText(msg))
	toast.SetTimeout(5)
	overlay.AddToast(toast)
}

func bytesPixbuf(b []byte) (*gdkpixbuf.Pixbuf, error) {
	// TODO: do this without creating a temp file
	f, err := os.CreateTemp("/tmp", fmt.Sprintf("%s-pixbuf", variant.DisplayName))
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name()) //nolint:errcheck
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
	glib.IdleAdd(cb)
}

func uiTicker(d uint, cb func() bool) {
	glib.TimeoutSecondsAdd(d, cb)
}

func ensureContextError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
		return err
	}
}
