package main

import (
	"fmt"
	"time"

	"github.com/geteduroam/linux-app/internal/notification"
	"github.com/geteduroam/linux-app/internal/utilsx"
	"github.com/geteduroam/linux-app/internal/variant"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type SuccessState struct {
	builder    *gtk.Builder
	parent     *gtk.Window
	stack      *adw.ViewStack
	vBeg       *time.Time
	vEnd       *time.Time
	isredirect bool
}

func NewSuccessState(builder *gtk.Builder, parent *gtk.Window, stack *adw.ViewStack, vBeg *time.Time, vEnd *time.Time, isredirect bool) *SuccessState {
	return &SuccessState{
		builder:    builder,
		parent:     parent,
		stack:      stack,
		vBeg:       vBeg,
		vEnd:       vEnd,
		isredirect: isredirect,
	}
}

func (s *SuccessState) Initialize() {
	var page adw.ViewStackPage
	s.builder.GetObject("successPage").Cast(&page)
	defer page.Unref()

	var title gtk.Label
	s.builder.GetObject("successTitle").Cast(&title)
	defer title.Unref()
	styleWidget(&title, "title")
	if s.isredirect {
		title.SetText("Follow the instructions at the link opened in your browser")
	}

	var logo gtk.Image
	s.builder.GetObject("successLogo").Cast(&logo)
	defer logo.Unref()
	res := MustResource("images/success.png")
	pb, err := bytesPixbuf([]byte(res))
	if err == nil {
		logo.SetFromPixbuf(pb)
		logo.SetSizeRequest(64, 64)
	}

	var sub gtk.Label
	s.builder.GetObject("successSubTitle").Cast(&sub)
	defer sub.Unref()
	sub.SetVisible(!s.isredirect)
	sub.SetText(fmt.Sprintf("Your %s profile has been added", variant.ProfileName))
	styleWidget(&sub, "label")

	var valid gtk.Label
	s.builder.GetObject("validityText").Cast(&valid)
	defer valid.Unref()
	validText := valid.GetText()
	if s.vBeg == nil {
		valid.Hide()
		valid.Unref()
	} else {
		uiTicker(1, func() bool {
			delta := time.Until(*s.vBeg)
			// We do not want to show on 0 seconds
			if delta >= 1*time.Second {
				valid.SetMarkup(fmt.Sprintf("Your profile will be valid in: %s", utilsx.DeltaTime(delta, "<b>", "</b>")))
				valid.Show()
				return true
			}
			if s.vEnd != nil {
				valid.SetMarkup(fmt.Sprintf("%s <b>%d</b> days", validText, utilsx.ValidityDays(*s.vEnd)))
			} else { // not very realistic this happens, but in theory it could
				valid.SetMarkup("Your profile is valid")
			}
			valid.Show()
			valid.Unref()
			return false
		})
	}

	// set the page as current
	setPage(s.stack, &page)

	if s.vEnd == nil {
		return
	}
	if !notification.HasDaemonSupport() {
		return
	}

	dialog := gtk.NewMessageDialog(s.parent, gtk.DialogDestroyWithParentValue, gtk.MessageQuestionValue, gtk.ButtonsYesNoValue, "This connection profile will expire in %i days.\n\nDo you want to enable notifications that warn for imminent expiry using systemd?", utilsx.ValidityDays(*s.vEnd))
	dialog.Present()
	var dialogcb func(gtk.Dialog, int)
	dialogcb = func(_ gtk.Dialog, response int) {
		defer glib.UnrefCallback(&dialogcb) //nolint:errcheck
		notification.ConfigureDaemon(int32(response) == int32(gtk.ResponseYesValue))
		dialog.Destroy()
	}
	dialog.ConnectResponse(&dialogcb)
	dialog.Present()
}
