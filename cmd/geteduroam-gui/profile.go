package main

import (
	"golang.org/x/exp/slog"

	"github.com/geteduroam/linux-app/internal/provider"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type ProfileState struct {
	builder  *gtk.Builder
	stack    *adw.ViewStack
	profiles []provider.Profile
	success  func(provider.Profile)
	sl       *SelectList
}

func NewProfileState(builder *gtk.Builder, stack *adw.ViewStack, profiles []provider.Profile, success func(provider.Profile)) *ProfileState {
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

	sorter := func(a, b int) int {
		// Here we have no search query
		return provider.SortNames(p.profiles[a].Name, p.profiles[b].Name, "")
	}
	activated := func(idx int) {
		go func() {
			p.success(p.profiles[idx])
			uiThread(func() {
				p.Destroy()
			})
		}()
	}

	p.sl = NewSelectList(&scroll, &list, activated, sorter)

	for idx, prof := range p.profiles {
		p.sl.Add(idx, prof.Name.Get())
	}

	p.sl.Setup()
	setPage(p.stack, &page)
}
