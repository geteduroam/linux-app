package main

import (
	"fmt"
	"time"

	"github.com/geteduroam/linux-app/internal/utils"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type SuccessState struct {
	builder *gtk.Builder
	stack   *adw.ViewStack
	expiry  *time.Time
}

func NewSuccessState(builder *gtk.Builder, stack *adw.ViewStack, expiry *time.Time) *SuccessState {
	return &SuccessState{
		builder: builder,
		stack:   stack,
		expiry:  expiry,
	}
}

func (s *SuccessState) Initialize() error {
	var page adw.ViewStackPage
	defer page.Unref()
	s.builder.GetObject("successPage").Cast(&page)

	var title gtk.Label
	defer title.Unref()
	s.builder.GetObject("successTitle").Cast(&title)
	styleWidget(&title, "label")

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
	s.stack.SetVisibleChild(page.GetChild())
	return nil
}
