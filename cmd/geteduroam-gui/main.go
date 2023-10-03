package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/gtk"

	"github.com/geteduroam/linux-app/internal/discovery"
	"github.com/geteduroam/linux-app/internal/handler"
	"github.com/geteduroam/linux-app/internal/instance"
	"github.com/geteduroam/linux-app/internal/network"
)

type serverList struct {
	sync.Mutex
	store     *gtk.StringList
	instances instance.Instances
	list      *SelectList
}

func (s *serverList) get(idx int) (*instance.Instance, error) {
	if idx < 0 || idx > len(s.instances) {
		return nil, errors.New("index out of range")
	}
	return &s.instances[idx], nil
}

func (s *serverList) Fill() {
	s.Lock()
	defer s.Unlock()
	for idx, inst := range s.instances {
		s.list.Add(idx, inst.Name)
	}
}

type mainState struct {
	builder *gtk.Builder
	servers *serverList
	scroll  *gtk.ScrolledWindow
}

func (m *mainState) initServers() {
	m.servers = &serverList{}
	m.servers.store = gtk.NewStringList(0)
}

func (m *mainState) activate() {
	var stack adw.ViewStack
	m.builder.GetObject("pageStack").Cast(&stack)
	defer stack.Unref()
	var page adw.ViewStackPage
	m.builder.GetObject("searchPage").Cast(&page)
	defer page.Unref()
	stack.SetVisibleChild(page.GetChild())
}

func (m *mainState) askCredentials(c network.Credentials, pi network.ProviderInfo) (string, string, error) {
	var stack adw.ViewStack
	m.builder.GetObject("pageStack").Cast(&stack)
	defer stack.Unref()
	login := NewLoginState(m.builder, &stack, c, pi)
	login.Initialize()
	user, pass := login.Get()
	return user, pass, nil
}

func (m *mainState) file(metadata []byte) (*time.Time, error) {
	h := handler.Handlers{
		CredentialsH: m.askCredentials,
		//CertificateH: askCertficiate,
	}
	return h.Configure(metadata)
}

func (m *mainState) direct(p instance.Profile) error {
	config, err := p.EAPDirect()
	if err != nil {
		return err
	}
	_, err = m.file(config)
	return err
}

func (m *mainState) local(path string) (*time.Time, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	v, err := m.file(b)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (m *mainState) oauth(p instance.Profile) (*time.Time, error) {
	config, err := p.EAPOAuth(func(url string) {
		uiThread(func() {
			var stack adw.ViewStack
			m.builder.GetObject("pageStack").Cast(&stack)
			defer stack.Unref()
			l := NewLoadingPage(m.builder, &stack, "Your browser has been opened to authorize the client")
			l.Initialize()
			// If the browser does not open for some reason the user could grab it with stdout
			// We could also show it in the UI but this might mean too much clutter
			fmt.Println("Browser has been opened with URL:", url)
		})
	})
	if err != nil {
		return nil, err
	}

	return m.file(config)
}

func (m *mainState) rowActived(sel instance.Instance) {
	var stack adw.ViewStack
	m.builder.GetObject("pageStack").Cast(&stack)
	defer stack.Unref()
	var page gtk.Box
	m.builder.GetObject("searchPage").Cast(&page)
	defer page.Unref()
	l := NewLoadingPage(m.builder, &stack, "Loading organization details...")
	l.Initialize()
	chosen := func(p instance.Profile) error {
		var valid *time.Time
		var err error
		switch p.Flow() {
		case instance.DirectFlow:
			err = m.direct(p)
			if err != nil {
				return err
			}
		case instance.OAuthFlow:
			valid, err = m.oauth(p)
			if err != nil {
				return err
			}
		case instance.RedirectFlow:
			return errors.New("not implemented yet")
		}
		s := NewSuccessState(m.builder, &stack, valid)
		uiThread(func() {
			s.Initialize()
		})
		return nil
	}
	if len(sel.Profiles) > 1 {
		profiles := NewProfileState(m.builder, &stack, sel.Profiles, chosen)
		profiles.Initialize()
	} else {
		go func() {
			err := chosen(sel.Profiles[0])
			if err != nil {
				l.Hide()
				m.activate()
				m.ShowError(err)
			}
		}()
	}
}

func (m *mainState) initList() {
	// style the treeview
	var list gtk.ListView
	m.builder.GetObject("searchList").Cast(&list)
	defer list.Unref()

	cache := discovery.NewCache()
	inst, err := cache.Instances()
	if err != nil {
		m.ShowError(err)
		return
	}
	m.servers.instances = *inst

	var search gtk.SearchEntry
	m.builder.GetObject("searchBox").Cast(&search)
	defer search.Unref()

	activated := func(idx int) {
		inst, err := m.servers.get(idx)
		// TODO: handle error
		if err != nil {
			m.ShowError(err)
			return
		}
		m.rowActived(*inst)
	}

	sorter := func(a, b string) int {
		return instance.SortNames(a, b, search.GetText())
	}

	m.servers.list = NewSelectList(m.scroll, &list, activated, sorter).WithFiltering(func(a string) bool {
		return instance.FilterSingle(a, search.GetText())
	})

	// Fill the servers in the select list
	m.servers.Fill()

	// Further set up the list
	m.servers.list.Setup()

	// Update the list when searching
	search.ConnectSearchChanged(func(_ gtk.SearchEntry) {
		// TODO len returns length in bytes
		// utf8.RuneCountInString() counts number of characters (runes)
		if len(search.GetText()) <= 2 {
			m.servers.list.Hide()
			return
		}
		m.servers.list.Changed()
		m.servers.list.Show()
	})
}

func (m *mainState) initBurger(app *adw.Application) {
	var gears gtk.MenuButton
	m.builder.GetObject("gears").Cast(&gears)
	defer gears.Unref()

	var menu gio.MenuModel
	builder := gtk.NewFromStringBuilder(MustResource("gears.ui"), -1)
	defer builder.Unref()
	builder.GetObject("menu").Cast(&menu)
	gears.SetMenuModel(&menu)

	imp := gio.NewSimpleAction("import-local", nil)
	var stack adw.ViewStack
	m.builder.GetObject("pageStack").Cast(&stack)
	defer stack.Unref()
	imp.ConnectActivate(func(_ gio.SimpleAction, _ uintptr) {
		fd, err := NewFileDialog(app.GetActiveWindow(), "Choose an EAP metadata file")
		if err != nil {
			m.ShowError(err)
			return
		}
		fd.Run(func(path string) {
			go func() {
				v, err := m.local(path)
				if err != nil {
					m.ShowError(err)
					return
				}
				s := NewSuccessState(m.builder, &stack, v)
				uiThread(func() {
					s.Initialize()
				})
			}()
		})
	})

	about := gio.NewSimpleAction("about", nil)
	about.ConnectActivate(func(_ gio.SimpleAction, _ uintptr) {
		var awin adw.AboutWindow
		adw.NewAboutWindow().Cast(&awin)
		// TODO: Make the version a global var somewhere
		awin.SetVersion("0.1-dev")
		awin.SetApplicationName("Geteduroam Linux")
		awin.SetWebsite("https://github.com/geteduroam/linux-app")
		// SetLicenseType has a scary warning: "comes with absolutely no warranty"
		// While it is true according to the license, I find it unfriendly
		awin.SetLicense("This application has a BSD 3 license.")
		awin.SetIssueUrl("https://github.com/geteduroam/linux-app/issues/new")
		awin.SetDevelopers([]string{"Jeroen Wijenbergh", "Martin van Es", "Alexandru Cacean"})
		awin.SetTransientFor(app.GetActiveWindow())
		awin.Show()
	})

	app.AddAction(imp)
	app.AddAction(about)
}

func (m *mainState) Initialize(app *adw.Application) {
	m.scroll = &gtk.ScrolledWindow{}
	m.builder.GetObject("searchScroll").Cast(m.scroll)
	m.initServers()
	m.initList()
	m.initBurger(app)
}

func (m *mainState) ShowError(err error) {
	toast := adw.NewToast(err.Error())
	toast.SetTimeout(5)
	var overlay adw.ToastOverlay
	m.builder.GetObject("searchToastOverlay").Cast(&overlay)
	defer overlay.Unref()
	overlay.AddToast(toast)
}

type ui struct {
	builder *gtk.Builder
	app     *adw.Application
}

func (ui *ui) initBuilder() {
	// open the builder
	ui.builder = gtk.NewFromStringBuilder(MustResource("geteduroam.ui"), -1)
}

func (ui *ui) initWindow() {
	// get the window
	var win gtk.Window
	ui.builder.GetObject("mainWindow").Cast(&win)
	defer win.Unref()
	win.SetDefaultSize(400, 600)
	// style the window using the css
	var search adw.ViewStackPage
	ui.builder.GetObject("searchPage").Cast(&search)
	defer search.Unref()
	widg := search.GetChild()
	defer widg.Unref()
	styleWidget(widg, "window")
	ui.app.AddWindow(&win)
	win.Show()
}

func (ui *ui) activate() {
	// Initialize the builder
	// The builder essentially just creates the bulk of the UI by loading the XML specification
	ui.initBuilder()

	// Initialize the rest of the window
	ui.initWindow()

	// Go to the main state
	m := &mainState{builder: ui.builder}
	m.Initialize(ui.app)
}

func (ui *ui) Run() int {
	const id = "com.geteduroam.linux"
	ui.app = adw.NewApplication(id, gio.GApplicationFlagsNoneValue)
	defer ui.app.Unref()
	ui.app.ConnectActivate(func(o gio.Application) {
		ui.activate()
	})
	return ui.app.Run(len(os.Args), os.Args)
}

func main() {
	ui := ui{}
	ui.Run()
}
