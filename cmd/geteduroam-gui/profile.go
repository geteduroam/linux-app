package main

import (
	"golang.org/x/exp/slog"

	"github.com/geteduroam/linux-app/internal/instance"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type ProfileState struct {
	builder  *gtk.Builder
	stack    *adw.ViewStack
	profiles []instance.Profile
	success  func(instance.Profile)
	sl *SelectList
}

func NewProfileState(builder *gtk.Builder, stack *adw.ViewStack, profiles []instance.Profile, success func(instance.Profile)) *ProfileState {
	return &ProfileState{
		builder:  builder,
		stack:    stack,
		profiles: profiles,
		success:  success,
	}
}

func (p *ProfileState) Destroy() {
	p.sl.Destroy()
}

func (p *ProfileState) ShowError(err error) {
	slog.Error(err.Error(), "state", "profile")
	var overlay adw.ToastOverlay
	p.builder.GetObject("profileToastOverlay").Cast(&overlay)
	defer overlay.Unref()
	showErrorToast(overlay, err)
}

func (p *ProfileState) Initialize() {
	var page adw.ViewStackPage
	p.builder.GetObject("profilePage").Cast(&page)
	defer page.Unref()
	var scroll gtk.ScrolledWindow
	p.builder.GetObject("profileScroll").Cast(&scroll)
	defer scroll.Unref()
	var list gtk.ListView
	p.builder.GetObject("profileList").Cast(&list)
	defer list.Unref()

	var label gtk.Label
	p.builder.GetObject("profileLabel").Cast(&label)
	defer label.Unref()
	styleWidget(&label, "label")

	sorter := func(a, b string) int {
		// Here we have no search query
		return instance.SortNames(a, b, "")
	}
	activated := func(idx int) {
		go p.success(p.profiles[idx])
		p.Destroy()
	}

	p.sl = NewSelectList(&scroll, &list, activated, sorter)

	for idx, prof := range p.profiles {
		p.sl.Add(idx, prof.Name)
	}

	p.sl.Setup()
	p.stack.SetVisibleChild(page.GetChild())
}
