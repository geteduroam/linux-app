// this file implements abstractions over a listview
package main

import (
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type SelectList struct {
	win       *gtk.ScrolledWindow
	list      *gtk.ListView
	activated func(int)
	sorter    func(a, b string) int
	filter    func(a string) bool
	store     *gtk.StringList
	cf        *gtk.CustomFilter
	cs        *gtk.CustomSorter
}

func stringFromPtr(ptr uintptr) string {
	// TODO: Remove this once we have proper callback type signatures
	// The callback should already give a gobject.Binding
	thisl := gobject.BindingNewFromInternalPtr(ptr)
	var thisv gobject.Value
	thisl.GetProperty("string", &thisv)
	// Purego makes a copy of the string when we call .String()
	// So we can safely unset after the return
	defer thisv.Unset()
	return thisv.String()
}

func setupList(item uintptr) {
	iteml := gtk.ListItemNewFromInternalPtr(item)
	label := gtk.NewLabel("")
	defer label.Unref()
	label.Set("xalign", 0)
	iteml.SetChild(label)
	label.SetMarginTop(5)
	label.SetMarginBottom(5)
}

func bindList(item uintptr) {
	iteml := gtk.ListItemNewFromInternalPtr(item)
	var label gtk.Label
	var strobj gtk.StringObject
	iteml.GetChild().Cast(&label)
	defer label.Unref()
	iteml.GetItem().Cast(&strobj)
	defer strobj.Unref()
	label.SetText(strobj.GetString())
}

func NewSelectList(win *gtk.ScrolledWindow, list *gtk.ListView, activated func(int), sorter func(a, b string) int) *SelectList {
	return &SelectList{
		win:       win,
		list:      list,
		sorter:    sorter,
		activated: activated,
		// TODO: unref this when this select list should be destroyed
		store: gtk.NewStringList(0),
	}
}

func (s *SelectList) Add(idx int, label string) {
	s.store.Append(label)
	var strobj gtk.StringObject
	// TODO: this is quite hacky but puregotk doesn't support subclassing yet
	// We have to store the mondel index as the position will not always match 1:1
	// In the beginning it will but after filtering the positions will only show the positions of the current model
	// Whereas we need the positions/index of the original list
	s.store.GetObject(uint(idx)).Cast(&strobj)
	defer strobj.Unref()
	strobj.SetData("model-index", uintptr(idx))
}

func (s *SelectList) Show() {
	s.win.Show()
}

func (s *SelectList) Hide() {
	s.win.Hide()
}

func (s *SelectList) WithFiltering(filter func(a string) bool) *SelectList {
	s.filter = filter
	return s
}

func (s *SelectList) Changed() {
	s.cs.Changed(0)
	if s.cf != nil {
		s.cf.Changed(0)
	}
}

func (s *SelectList) setupFactory() *gtk.ListItemFactory {
	factory := gtk.NewSignalListItemFactory()
	factory.Connect("signal::setup", gobject.NewCallback(func(_ uintptr, item uintptr) {
		setupList(item)
	}), 0)

	factory.Connect("signal::bind", gobject.NewCallback(func(_ uintptr, item uintptr) {
		bindList(item)
	}), 0)

	return factory
}

func (s *SelectList) setupSorter(base gio.ListModel) gio.ListModel {
	s.cs = gtk.NewCustomSorter(func(this uintptr, other uintptr, _ uintptr) int {
		return s.sorter(stringFromPtr(this), stringFromPtr(other))
	}, 0, func(uintptr) {
		// TODO: do something on destroy?
	})
	var sort gtk.Sorter
	s.cs.Cast(&sort)
	sm := gtk.NewSortListModel(base, &sort)
	return sm
}

func (s *SelectList) setupFilter(base gio.ListModel) gio.ListModel {
	s.cf = gtk.NewCustomFilter(func(item uintptr, _ uintptr) bool {
		return s.filter(stringFromPtr(item))
	}, 0, func(uintptr) {
		// TODO: do something on destroy?
	})
	var fil gtk.Filter
	s.cf.Cast(&fil)
	fl := gtk.NewFilterListModel(base, &fil)
	return fl
}

func (s *SelectList) Setup() {
	factory := s.setupFactory()
	defer factory.Unref()
	var model gio.ListModel = s.store
	if s.filter != nil {
		model = s.setupFilter(model)
	}
	// We never want horizontal scrollbars, but want automatically vertical ones
	s.win.SetPolicy(gtk.PolicyExternalValue, gtk.PolicyAutomaticValue)

	// further setup the list by setting the factory and model
	sel := gtk.NewSingleSelection(s.setupSorter(model))
	defer sel.Unref()
	s.list.SetFactory(factory)
	s.list.SetModel(sel)

	// We want to activate on single click always
	s.list.SetSingleClickActivate(true)

	// Call the activated callback
	s.list.ConnectActivate(func(_ gtk.ListView, _ uint) {
		var strobj gtk.StringObject
		sel.GetSelectedItem().Cast(&strobj)
		defer strobj.Unref()
		index := int(strobj.GetData("model-index"))
		s.activated(index)
	})

	// style the widget
	styleWidget(s.list, "list")
}
