package main

import (
	"golang.org/x/exp/slog"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/geteduroam/linux-app/internal/provider"
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

func (p *ProfileState) ShowError(err error) {
	slog.Error(err.Error(), "state", "profile")
	overlay := p.builder.GetObject("profileToastOverlay").Cast().(*adw.ToastOverlay)
	showErrorToast(overlay, err)
}

func (p *ProfileState) Initialize() {
	page := p.builder.GetObject("profilePage").Cast().(*adw.ViewStackPage)
	scroll := p.builder.GetObject("profileScroll").Cast().(*gtk.ScrolledWindow)
	list := p.builder.GetObject("profileList").Cast().(*gtk.ListView)

	label := p.builder.GetObject("profileLabel").Cast().(*gtk.Label)
	styleWidget(label, "label")

	sorter := func(a, b int) int {
		// Here we have no search query
		return provider.SortNames(p.profiles[a].Name, p.profiles[b].Name, "")
	}
	activated := func(idx int) {
		go func() {
			p.success(p.profiles[idx])
		}()
	}

	p.sl = NewSelectList(scroll, list, activated, sorter)

	for idx, prof := range p.profiles {
		p.sl.Add(idx, prof.Name.Get())
	}

	p.sl.Setup()
	setPage(p.stack, page)
}
