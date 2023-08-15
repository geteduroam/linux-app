package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"

	"sync"

	"github.com/geteduroam/linux-app/internal/discovery"

	"github.com/geteduroam/linux-app/internal/instance"
)

type serverList struct {
	sync.Mutex
	iter *gtk.TreeIter
	store *gtk.ListStore
	instances instance.Instances
}

func (s *serverList) get(idx int) (*instance.Instance, error) {
	if idx < 0 || idx > len(s.instances) {
		return nil, errors.New("index out of range")
	}
	return &s.instances[idx], nil
}

func (s *serverList) GetSelected(sel *gtk.TreeSelection) (*instance.Instance, error) {
	s.Lock()
	defer s.Unlock()
	sel.GetSelected(nil, s.iter)

	// get the value index
	var val gobject.Value
	s.store.GetValue(s.iter, 1, &val)
	idx := val.Int()
	entry, err := s.get(idx)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *serverList) Clear() {
	s.store.Clear()
}

func (s *serverList) Add(idx int, srv instance.Instance) {
	s.store.Append(s.iter)
	s.store.Set(s.iter, 0, srv.Name, -1)
	s.store.Set(s.iter, 1, idx, -1)
}

func (s *serverList) Model() gtk.TreeModel {
	s.Lock()
	defer s.Unlock()
	return s.store
}

type mainState struct{
	builder *gtk.Builder
	servers *serverList
	scroll *gtk.ScrolledWindow
}

func (m *mainState) initServers() {
	store := gtk.NewListStore(2, gobject.TypeStringVal, gobject.TypeIntVal)
	m.servers = &serverList{}
	m.servers.store = store
	m.servers.iter = &gtk.TreeIter{}
}

func (m *mainState) rowActived(tree gtk.TreeView) {
	s, err := m.servers.GetSelected(tree.GetSelection())
	if err != nil {
		return
	}
	var stack adw.ViewStack
	m.builder.GetObject("pageStack").Cast(&stack)
	var page gtk.Box
	m.builder.GetObject("searchPage").Cast(&page)
	l := NewLoadingPage(m.builder, &stack, "Loading organization details...")
	l.Show()

	if len(s.Profiles) > 1 {
		panic("A profile selection screen is not yet implemented")
	}
	go func() {
		p := s.Profiles[0]
		switch p.Flow() {
		case instance.DirectFlow:
			fmt.Println("DIRECT FLOW")
		case instance.OAuthFlow:
			fmt.Println("OAUTH FLOW")
		case instance.RedirectFlow:
			fmt.Println("REDIRECT FLOW")
		}
	}()
}

func (m *mainState) initTree() {
	// style the treeview
	var tree gtk.TreeView
	m.builder.GetObject("searchTree").Cast(&tree)
	styleWidget(&tree, "tree")
	tree.SetHeadersVisible(false)
	column := gtk.NewTreeViewColumn()
	tree.AppendColumn(column)

	renderer := gtk.NewCellRendererText()
	renderer.Set("ypad", 10)
	// We never want horizontal scrollbars, but want automatically vertical ones
	m.scroll.SetPolicy(gtk.PolicyNeverValue, gtk.PolicyAutomaticValue)
	// TODO: The height here is hacky as it could depend on the system fonts (?)
	// We have to set a fixed width because we don't want a scrolledwindow scrollbar (see the policy set above)
	// The height has to always be set because when I pass -1 (automatic), some entries are larger than others
	renderer.SetFixedSize(400, 50)
	column.PackStart(renderer, true)
	column.AddAttribute(renderer, "text", 0)
	tree.SetActivateOnSingleClick(true)

	// when an entry is clicked we want to get the selection
	// and do the operation
	tree.ConnectRowActivated(func(gtk.TreeView, uintptr, uintptr) {
		m.rowActived(tree)
	})

	tree.SetModel(m.servers.Model())
}

func (m *mainState) initSearch() {
	var search gtk.SearchEntry
	m.builder.GetObject("searchBox").Cast(&search)
	c := discovery.NewCache()
	search.ConnectSearchChanged(func(gtk.SearchEntry) {
		// get the query 
		var val gobject.Value
		search.GetProperty("text", &val)
		q := val.String()
		// update the search with the query
		go m.fillSearch(c, q)
	})
}

func (m *mainState) Initialize() error {
	m.scroll = &gtk.ScrolledWindow{}
	m.builder.GetObject("searchScroll").Cast(m.scroll)
	m.initServers()
	m.initTree()
	m.initSearch()
	return nil
}

func (m *mainState) State() StateType {
	return MainState
}

func (m *mainState) fillSearch(cache *discovery.Cache, search string) {
	m.servers.Lock()
	defer m.servers.Unlock()
	if search == "" {
		uiThread(func () {
			m.scroll.Hide()
			m.servers.Clear()
		})
		return
	}
	inst, err := cache.Instances()
	if err != nil {
		panic(err)
	}
	m.servers.instances = *inst.Filter(search)

	var wg sync.WaitGroup
	wg.Add(1)

	// update the list in the gtk thread
	uiThread( func() {
		m.servers.Clear()
		for idx, ins := range m.servers.instances {
			m.servers.Add(idx, ins)
		}
		m.scroll.Show()
		wg.Done()
	})
	wg.Wait()
}


type ui struct {
	builder *gtk.Builder
	app *adw.Application
	state State
}

func (ui *ui) initBuilder() {
	// open the builder
	ui.builder = gtk.NewFromStringBuilder(MustResource("geteduroam.ui"), -1)
}

func (ui *ui) initWindow() {
	// get the window
	var win gtk.Window
	ui.builder.GetObject("mainWindow").Cast(&win)
	win.SetDefaultSize(400, 600)


	// style the window using the css
	var search adw.ViewStackPage
	ui.builder.GetObject("searchPage").Cast(&search)
	widg := search.GetChild().GetLayoutManager().GetWidget()
	styleWidget(widg, "window")
	ui.app.AddWindow(&win)
	win.Show()
}

// Go transitions the UI to a new state
// In the future we might want to do this with a FSM
// So for now this is a really dumb setter
func (ui *ui) Go(state State) {
	ui.state = state
	if err := state.Initialize(); err != nil {
		ui.app.GetActiveWindow().Close()
	}
}

func (ui *ui) activate() {
	// Initialize the builder
	// The builder essentially just creates the bulk of the UI by loading the XML specification
	ui.initBuilder()

	// Initialize the rest of the window
	ui.initWindow()

	// Go to the main state
	ui.Go(&mainState{builder: ui.builder})
}

func (ui *ui) Run() int {
	const id = "com.geteduroam.linux"
	ui.app = adw.NewApplication(id, gio.GApplicationFlagsNoneValue)
	defer ui.app.Unref()
	ui.app.ConnectActivate(func (o gio.Application) {
		ui.activate()
	})
	return ui.app.Run(len(os.Args), os.Args)
}

func main() {
	ui := ui{}
	ui.Run()
}
