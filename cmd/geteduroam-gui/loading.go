package main

import (
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type LoadingState struct {
	builder *gtk.Builder
	stack   *adw.ViewStack
	Message string
}

func NewLoadingPage(builder *gtk.Builder, stack *adw.ViewStack, message string) *LoadingState {
	return &LoadingState{
		builder: builder,
		stack:   stack,
		Message: message,
	}
}

func (l *LoadingState) Show() {
	var page adw.ViewStackPage
	l.builder.GetObject("loadingPage").Cast(&page)
	var label gtk.Label
	l.builder.GetObject("loadingText").Cast(&label)
	label.SetText(l.Message)
	styleWidget(&label, "label")
	l.stack.SetVisibleChild(page.GetChild().GetLayoutManager().GetWidget())

	var spinner gtk.Spinner
	l.builder.GetObject("loadingSpinner").Cast(&spinner)
	spinner.Start()
}
