package main

import (
	"fmt"
	"time"

	"github.com/geteduroam/linux-app/internal/notification"
	"github.com/geteduroam/linux-app/internal/utils"
	"github.com/geteduroam/linux-app/internal/variant"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type SuccessState struct {
	builder    *gtk.Builder
	parent     *gtk.Window
	stack      *adw.ViewStack
	expiry     *time.Time
	isredirect bool
}

func NewSuccessState(builder *gtk.Builder, parent *gtk.Window, stack *adw.ViewStack, expiry *time.Time, isredirect bool) *SuccessState {
	return &SuccessState{
		builder:    builder,
		parent:     parent,
		stack:      stack,
		expiry:     expiry,
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

	var expiry gtk.Label
	s.builder.GetObject("expiryText").Cast(&expiry)
	defer expiry.Unref()
	if s.expiry != nil {
		expiry.SetMarkup(fmt.Sprintf("%s <b>%d</b> days", expiry.GetText(), utils.ValidityDays(*s.expiry)))
		expiry.Show()
	} else {
		expiry.Hide()
	}
	// set the page as current
	setPage(s.stack, &page)

	if s.expiry == nil {
		return
	}
	if !notification.HasDaemonSupport() {
		return
	}

	dialog := adw.NewMessageDialog(s.parent, "Enable notifications?", fmt.Sprintf("This connection profile will expire in %d days.\n\nDo you want to enable notifications that warn for imminent expiry using systemd?", utils.ValidityDays(*s.expiry)))
	dialog.AddResponse("disable", "Disable")
	dialog.AddResponse("enable", "Enable")
	dialog.SetResponseAppearance("enable", adw.ResponseSuggestedValue)
	dialog.SetDefaultResponse("disable")
	dialog.SetCloseResponse("disale")

	var dialogcb func(adw.MessageDialog, string)
	dialogcb = func(_ adw.MessageDialog, response string) {
		defer glib.UnrefCallback(&dialogcb) //nolint:errcheck
		notification.ConfigureDaemon(response == "enable")
	}
	dialog.ConnectResponse(&dialogcb)
	dialog.Present()
}
