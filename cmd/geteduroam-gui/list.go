// this file implements abstractions over a listview
package main

import (
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type SelectList struct {
	SignalPool
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
	// Purego makes a copy of the string when we call .GetString()
	// So we can safely unset after the return
	defer thisv.Unset()
	return thisv.GetString()
}

func setupList(item uintptr) {
	iteml := gtk.ListItemNewFromInternalPtr(item)
	label := gtk.NewLabel("")
	defer label.Unref()
	label.Set("xalign", 0)
	iteml.SetChild(&label.Widget)
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
		store:     gtk.NewStringList(0),
	}
}

func (s *SelectList) Destroy() {
	s.DisconnectSignals()
	s.store.Unref()
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

func (s *SelectList) setupFactory() *gtk.SignalListItemFactory {
	factory := gtk.NewSignalListItemFactory()
	// TODO: Add signal for cleanup
	setupcb := func(_ uintptr, item uintptr) {
		setupList(item)
	}
	bindcb := func(_ uintptr, item uintptr) {
		bindList(item)
	}
	factory.Connect("signal::setup", glib.NewCallback(&setupcb), 0)

	// TODO: Add signal for cleanup
	factory.Connect("signal::bind", glib.NewCallback(&bindcb), 0)

	return factory
}

func (s *SelectList) setupSorter(base gio.ListModel) gio.ListModel {
	sf := (glib.CompareDataFunc)(func(this uintptr, other uintptr, _ uintptr) int {
		return s.sorter(stringFromPtr(this), stringFromPtr(other))
	})

	destroycb := (glib.DestroyNotify)(func(uintptr) {
		// do nothing
	})

	s.cs = gtk.NewCustomSorter(&sf, 0, &destroycb)
	var sort gtk.Sorter
	s.cs.Cast(&sort)
	sm := gtk.NewSortListModel(base, &sort)
	return sm
}

func (s *SelectList) setupFilter(base gio.ListModel) gio.ListModel {
	cf := (gtk.CustomFilterFunc)(func (item uintptr, _ uintptr) bool {
		return s.filter(stringFromPtr(item))
	})
	destroycb := (glib.DestroyNotify)(func(uintptr) {
		// do nothing
	})
	s.cf = gtk.NewCustomFilter(&cf, 0, &destroycb)
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
	s.list.SetFactory(&factory.ListItemFactory)
	s.list.SetModel(sel)

	// We want to activate on single click always
	s.list.SetSingleClickActivate(true)

	actcb := func(_ gtk.ListView, _ uint) {
		var strobj gtk.StringObject
		sel.GetSelectedItem().Cast(&strobj)
		defer strobj.Unref()
		index := int(strobj.GetData("model-index"))
		s.activated(index)
	}

	// Call the activated callback
	s.AddSignal(s.list, s.list.ConnectActivate(&actcb))

	// style the widget
	styleWidget(s.list, "list")
}
