package main

import (
	"fmt"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/geteduroam/linux-app/internal/notification"
	"github.com/geteduroam/linux-app/internal/utilsx"
	"github.com/geteduroam/linux-app/internal/variant"
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
	page := s.builder.GetObject("successPage").Cast().(*adw.ViewStackPage)
	title := s.builder.GetObject("successTitle").Cast().(*gtk.Label)
	styleWidget(title, "title")
	if s.isredirect {
		title.SetText("Follow the instructions at the link opened in your browser")
	}

	logo := s.builder.GetObject("successLogo").Cast().(*gtk.Image)
	res := MustResource("images/success.png")
	pb, err := bytesPixbuf([]byte(res))
	if err == nil {
		logo.SetFromPixbuf(pb)
		logo.SetSizeRequest(64, 64)
	}

	sub := s.builder.GetObject("successSubTitle").Cast().(*gtk.Label)
	sub.SetVisible(!s.isredirect)
	sub.SetText(fmt.Sprintf("Your %s profile has been added", variant.ProfileName))
	styleWidget(sub, "label")

	valid := s.builder.GetObject("validityText").Cast().(*gtk.Label)
	validText := valid.Text()
	if s.vBeg == nil {
		valid.Hide()
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
			return false
		})
	}

	// set the page as current
	setPage(s.stack, page)

	if s.vEnd == nil {
		return
	}
	if !notification.HasDaemonSupport() {
		return
	}

	dialog := gtk.NewMessageDialog(s.parent, gtk.DialogDestroyWithParent, gtk.MessageQuestion, gtk.ButtonsYesNo)
	dialog.SetTitle(fmt.Sprintf("This connection profile will expire in %i days.\n\nDo you want to enable notifications that warn for imminent expiry using systemd?", utilsx.ValidityDays(*s.vEnd)))
	dialog.Present()
	dialogcb := func(response int) {
		notification.ConfigureDaemon(int32(response) == int32(gtk.ResponseYes))
		dialog.Destroy()
	}
	dialog.ConnectResponse(dialogcb)
	dialog.Present()
}
