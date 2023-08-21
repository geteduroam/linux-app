package main

import (
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type SuccessState struct {
	builder *gtk.Builder
	stack   *adw.ViewStack
}

func NewSuccessState(builder *gtk.Builder, stack *adw.ViewStack) *SuccessState {
	return &SuccessState{
		builder: builder,
		stack:   stack,
	}
}

func (s *SuccessState) Initialize() error {
	var page adw.ViewStackPage
	s.builder.GetObject("successPage").Cast(&page)

	var title gtk.Label
	s.builder.GetObject("successTitle").Cast(&title)
	styleWidget(&title, "label")
	// set the page as current
	s.stack.SetVisibleChild(page.GetChild())
	return nil
}
