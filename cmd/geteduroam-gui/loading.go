package main

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
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
	page := l.builder.GetObject("loadingPage").Cast().(*adw.ViewStackPage)
	label := l.builder.GetObject("loadingText").Cast().(*gtk.Label)
	label.SetText(l.Message)
	styleWidget(label, "label")
	setPage(l.stack, page)
	spinner := l.builder.GetObject("loadingSpinner").Cast().(*gtk.Spinner)
	l.spinner = spinner

	cancel := l.builder.GetObject("loadingCancel").Cast().(*gtk.Button)
	if l.Cancel != nil {
		cancel.SetVisible(true)
		cb := func() {
			l.Cancel()
		}
		cancel.ConnectClicked(cb)
	} else {
		cancel.SetVisible(false)
	}

	spinner.Start()
}
