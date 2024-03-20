package main

import (
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type LoadingState struct {
	builder *gtk.Builder
	stack   *adw.ViewStack
	spinner *gtk.Spinner
	Message string
	Cancel  func()
}

func NewLoadingPage(builder *gtk.Builder, stack *adw.ViewStack, message string, cancel func()) *LoadingState {
	return &LoadingState{
		builder: builder,
		stack:   stack,
		Message: message,
		Cancel:  cancel,
	}
}

func (l *LoadingState) Hide() {
	if l.spinner != nil {
		l.spinner.Stop()
	}
}

func (l *LoadingState) Initialize() {
	var page adw.ViewStackPage
	l.builder.GetObject("loadingPage").Cast(&page)
	defer page.Unref()
	var label gtk.Label
	l.builder.GetObject("loadingText").Cast(&label)
	defer label.Unref()
	label.SetText(l.Message)
	styleWidget(&label, "label")
	setPage(l.stack, &page)
	var spinner gtk.Spinner
	l.builder.GetObject("loadingSpinner").Cast(&spinner)
	defer spinner.Unref()
	l.spinner = &spinner

	var cancel gtk.Button
	l.builder.GetObject("loadingCancel").Cast(&cancel)
	defer cancel.Unref()
	if l.Cancel != nil {
		cancel.SetVisible(true)
		cb := func(_ gtk.Button) {
			l.Cancel()
		}
		cancel.ConnectClicked(&cb)
	} else {
		cancel.SetVisible(false)
	}

	spinner.Start()
}
