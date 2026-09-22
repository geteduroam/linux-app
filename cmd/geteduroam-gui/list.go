// this file implements abstractions over a listview
package main

import (
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func strObjToIdx(v *gtk.StringObject) int {
	gv, err := strconv.Atoi(v.String())
	if err != nil {
		panic(err)
	}
	return gv
}

type SelectList struct {
	win       *gtk.ScrolledWindow
	list      *gtk.ListView
	activated func(int)
	sorter    func(a, b int) int
	filter    func(idx int) bool
	store     *gtk.StringList
	cf        *gtk.CustomFilter
	cs        *gtk.CustomSorter
	items     map[string]string
}

func (s *SelectList) setup(item *gtk.ListItem) {
	label := gtk.NewLabel("")
	label.SetXAlign(0)
	item.SetChild(&label.Widget)
	label.SetMarginTop(5)
	label.SetMarginBottom(5)
}

func (s *SelectList) bind(item *gtk.ListItem) {
	label := item.Child().(*gtk.Label)
	strobj := item.Item().Cast().(*gtk.StringObject)
	label.SetText(s.items[strobj.String()])
}

func NewSelectList(win *gtk.ScrolledWindow, list *gtk.ListView, activated func(int), sorter func(a, b int) int) *SelectList {
	return &SelectList{
		win:       win,
		list:      list,
		sorter:    sorter,
		activated: activated,
		store:     gtk.NewStringList(nil),
		items:     make(map[string]string),
	}
}

func (s *SelectList) Add(idx int, label string) {
	key := strconv.Itoa(idx)
	s.items[key] = label
	s.store.Append(key)
}

func (s *SelectList) Remove(idx int) {
	delete(s.items, string(idx))
	s.store.Remove(uint(idx))
}

func (s *SelectList) Show() {
	s.win.Show()
}

func (s *SelectList) Hide() {
	s.win.Hide()
}

func (s *SelectList) WithFiltering(filter func(idx int) bool) *SelectList {
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
	setupcb := func(obj *glib.Object) {
		listItem := obj.Cast().(*gtk.ListItem)
		s.setup(listItem)
	}
	bindcb := func(obj *glib.Object) {
		listItem := obj.Cast().(*gtk.ListItem)
		s.bind(listItem)
	}
	factory.ConnectSetup(setupcb)
	factory.ConnectBind(bindcb)

	return factory
}

func (s *SelectList) setupSorter(base gio.ListModel) *gtk.SortListModel {
	s.cs = gtk.NewCustomSorter(glib.NewObjectComparer(func(a, b *gtk.StringObject) int {
		return s.sorter(strObjToIdx(a), strObjToIdx(b))
	}))
	sort := s.cs.Cast().(*gtk.CustomSorter)
	sm := gtk.NewSortListModel(&base, &sort.Sorter)
	return sm
}

func (s *SelectList) setupFilter(base gio.ListModel) *gtk.FilterListModel {
	s.cf = gtk.NewCustomFilter(func(obj *glib.Object) bool {
		strobj := obj.Cast().(*gtk.StringObject)
		return s.filter(strObjToIdx(strobj))
	})
	fil := s.cf.Cast().(*gtk.CustomFilter)
	fl := gtk.NewFilterListModel(&base, &fil.Filter)
	return fl
}

func (s *SelectList) Setup() {
	factory := s.setupFactory()
	var model gio.ListModel = s.store.ListModel
	if s.filter != nil {
		model = s.setupFilter(model).ListModel
	}
	// We never want horizontal scrollbars, but want automatically vertical ones
	s.win.SetPolicy(gtk.PolicyExternal, gtk.PolicyAutomatic)

	// further setup the list by setting the factory and model
	sel := gtk.NewSingleSelection(s.setupSorter(model))
	s.list.SetFactory(&factory.ListItemFactory)
	s.list.SetModel(sel)

	// We want to activate on single click always
	s.list.SetSingleClickActivate(true)

	actcb := func(_pos uint) {
		strobj := sel.SelectedItem().Cast().(*gtk.StringObject)
		s.activated(strObjToIdx(strobj))
	}
	s.list.ConnectActivate(actcb)

	// style the widget
	styleWidget(s.list, "list")
}
