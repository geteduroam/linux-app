package main

import (
	"strings"

	"github.com/geteduroam/linux-app/internal/instance"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type ProfileState struct {
	builder *gtk.Builder
	stack *adw.ViewStack
	profiles []instance.Profile
	success func(instance.Profile)
}

func NewProfileState(builder *gtk.Builder, stack *adw.ViewStack, profiles []instance.Profile, success func(instance.Profile)) *ProfileState {
	return &ProfileState{
		builder: builder,
		stack: stack,
		profiles: profiles,
		success: success,
	}
}

func (p *ProfileState) Initialize() error {
	var page adw.ViewStackPage
	p.builder.GetObject("profilePage").Cast(&page)
	var scroll gtk.ScrolledWindow
	p.builder.GetObject("profileScroll").Cast(&scroll)
	var list gtk.ListView
	p.builder.GetObject("profileList").Cast(&list)

	var label gtk.Label
	p.builder.GetObject("profileLabel").Cast(&label)
	styleWidget(&label, "label")

	sorter := strings.Compare
	activated := func(idx int) {
		p.success(p.profiles[idx])
	}

	sl := NewSelectList(&scroll, &list, activated, sorter)

	for idx, p := range p.profiles {
		sl.Add(idx, p.Name)
	}

	sl.Setup()
	p.stack.SetVisibleChild(page.GetChild())
	return nil
}
