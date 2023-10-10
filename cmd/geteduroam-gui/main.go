package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/slog"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"

	"github.com/geteduroam/linux-app/internal/discovery"
	"github.com/geteduroam/linux-app/internal/handler"
	"github.com/geteduroam/linux-app/internal/instance"
	"github.com/geteduroam/linux-app/internal/log"
	"github.com/geteduroam/linux-app/internal/network"
	"github.com/geteduroam/linux-app/internal/version"
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
	app     *adw.Application
	builder *gtk.Builder
	servers *serverList
	scroll  *gtk.ScrolledWindow
	stack   *adw.ViewStack
}

func (m *mainState) initServers() {
	m.servers = &serverList{}
	m.servers.store = gtk.NewStringList(0)
}

func (m *mainState) activate() {
	var page adw.ViewStackPage
	m.builder.GetObject("searchPage").Cast(&page)
	defer page.Unref()
	// we do not use setPage here as the margin is already set in the clamp
	// This is so that the background is set on the current main page only but it expands to each side fully
	m.stack.SetVisibleChild(page.GetChild())
}

func (m *mainState) askCredentials(c network.Credentials, pi network.ProviderInfo) (string, string, error) {
	login := NewCredentialsStateBase(m.builder, m.stack, c, pi)
	login.Initialize()
	user, pass := login.Get()
	return user, pass, nil
}

func (m *mainState) askCertificate(cert string, pwd string, pi network.ProviderInfo) (string, string, error) {
	base := NewCertificateStateBase(m.app.GetActiveWindow(), m.builder, m.stack, cert, pwd, pi)
	base.Initialize()
	cert, pass := base.Get()
	return cert, pass, nil
}

func (m *mainState) file(metadata []byte) (*time.Time, error) {
	h := handler.Handlers{
		CredentialsH: m.askCredentials,
		CertificateH: m.askCertificate,
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
			l := NewLoadingPage(m.builder, m.stack, "Your browser has been opened to authorize the client")
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
	var page gtk.Box
	m.builder.GetObject("searchPage").Cast(&page)
	defer page.Unref()
	l := NewLoadingPage(m.builder, m.stack, "Loading organization details...")
	l.Initialize()
	chosen := func(p instance.Profile) error {
		var valid *time.Time
		var err error
		var isredirect bool
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
			isredirect = true
			url, err := p.RedirectURI()
			if err != nil {
				return err
			}
			err = exec.Command("xdg-open", url).Start()
			if err != nil {
				return err
			}
			fmt.Println("Browser has been opened with URL:", url)
		}
		s := NewSuccessState(m.builder, m.stack, valid, isredirect)
		uiThread(func() {
			s.Initialize()
		})
		return nil
	}
	cb := func(p instance.Profile) {
		err := chosen(p)
		if err != nil {
			l.Hide()
			m.activate()
			m.ShowError(err)
		}
	}
	if len(sel.Profiles) > 1 {
		profiles := NewProfileState(m.builder, m.stack, sel.Profiles, cb)
		profiles.Initialize()
	} else {
		go cb(sel.Profiles[0])
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

func (m *mainState) localMetadata() {
	fd, err := NewFileDialog(m.app.GetActiveWindow(), "Choose an EAP metadata file")
	if err != nil {
		m.ShowError(err)
		return
	}
	fd.Run(func(path string) {
		go func() {
			v, err := m.local(path)
			if err != nil {
				uiThread(func() {
					m.activate()
					m.ShowError(err)
				})
				return
			}
			s := NewSuccessState(m.builder, m.stack, v, false)
			s.Initialize()
		}()
	})
}

func (m *mainState) initBurger() {
	var gears gtk.MenuButton
	m.builder.GetObject("gears").Cast(&gears)
	defer gears.Unref()

	var menu gio.MenuModel
	builder := gtk.NewFromStringBuilder(MustResource("gears.ui"), -1)
	defer builder.Unref()
	builder.GetObject("menu").Cast(&menu)
	gears.SetMenuModel(&menu)

	imp := gio.NewSimpleAction("import-local", nil)
	imp.ConnectActivate(func(_ gio.SimpleAction, _ uintptr) {
		m.localMetadata()
	})

	about := gio.NewSimpleAction("about", nil)
	about.ConnectActivate(func(_ gio.SimpleAction, _ uintptr) {
		var awin gtk.AboutDialog
		gtk.NewAboutDialog().Cast(&awin)
		awin.SetName("geteduroam Linux client")
		pb, err := bytesPixbuf([]byte(MustResource("images/geteduroam.png")))
		if err == nil {
			texture := gdk.NewForPixbufTexture(pb)
			defer pb.Unref()
			awin.SetLogo(texture)
			defer texture.Unref()
		}
		lpath, err := log.Location("geteduroam-gui")
		if err == nil {
			awin.SetSystemInformation("Log location: " + lpath)
		}
		awin.SetProgramName("geteduroam GUI")
		awin.SetComments("Client to easily and securely configure eduroam")
		awin.SetAuthors([]string{"Jeroen Wijenbergh", "Martin van Es", "Alexandru Cacean"})
		awin.SetVersion(version.Get())
		awin.SetWebsite("https://github.com/geteduroam/linux-app")
		// SetLicenseType has a scary warning: "comes with absolutely no warranty"
		// While it is true according to the license, I find it unfriendly
		awin.SetLicense("This application has a BSD 3 license.")
		awin.SetTransientFor(m.app.GetActiveWindow())
		awin.Show()
	})

	m.app.AddAction(imp)
	m.app.AddAction(about)
}

func (m *mainState) Initialize() {
	m.scroll = &gtk.ScrolledWindow{}
	m.builder.GetObject("searchScroll").Cast(m.scroll)
	m.stack = &adw.ViewStack{}
	m.builder.GetObject("pageStack").Cast(m.stack)
	m.initServers()
	m.initList()
	m.initBurger()
	m.activate()
}

func (m *mainState) ShowError(err error) {
	slog.Error(err.Error(), "state", "main")
	var overlay adw.ToastOverlay
	m.builder.GetObject("searchToastOverlay").Cast(&overlay)
	defer overlay.Unref()
	showErrorToast(overlay, err)
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
	m := &mainState{app: ui.app, builder: ui.builder}
	m.Initialize()
}

func (ui *ui) Run(args []string) int {
	const id = "com.geteduroam.linux"
	ui.app = adw.NewApplication(id, gio.GApplicationFlagsNoneValue)
	defer ui.app.Unref()
	ui.app.ConnectActivate(func(o gio.Application) {
		ui.activate()
	})

	return ui.app.Run(len(args), args)
}

func main() {
	const usage = `Usage of %s:
  -h, --help			Prints this help information
  --version			Prints version information
  -d, --debug			Debug
  --gtk-args                    Arguments to pass to gtk as a string, e.g. "--help". These flags are splitted on spaces
`

	var help bool
	var versionf bool
	var debug bool
	var gtkarg string
	program := "geteduroam-gui"
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&versionf, "version", false, "Show version")
	flag.BoolVar(&debug, "d", false, "Debug")
	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.StringVar(&gtkarg, "gtk-args", "", "Gtk arguments")
	flag.Usage = func() { fmt.Printf(usage, "geteduroam-gui") }
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	if versionf {
		fmt.Println(version.Get())
		return
	}

	var handler glib.LogFunc = func(pkg string, level glib.LogLevelFlags, msg string, _ uintptr) {
		switch level {
		case glib.GLogLevelErrorValue:
			slog.Error(msg, "pkg-name", pkg, "level", level)
		case glib.GLogLevelCriticalValue:
			// Ignore some false positives due to Gtk bug
			// Happens when pressing "Import Metadata"
			// see https://discourse.gnome.org/t/menu-button-gives-error-messages-with-latest-gtk4/15689/3
			ignore := "_gtk_css_corner_value_get_%s: assertion 'corner->class == &GTK_CSS_VALUE_CORNER' failed"
			if fmt.Sprintf(ignore, "x") == msg || fmt.Sprintf(ignore, "y") == msg {
				return
			}
			slog.Error("pkg-name", pkg, "level", level)
		case glib.GLogLevelWarningValue:
			slog.Warn(msg, "pkg-name", pkg, "level", level)
		case glib.GLogLevelMessageValue:
			slog.Info(msg, "pkg-name", pkg, "level", level)
		case glib.GLogLevelInfoValue:
			slog.Info(msg, "pkg-name", pkg, "level", level)
		case glib.GLogLevelDebugValue:
			slog.Debug(msg, "pkg-name", pkg, "level", level)
		case glib.GLogLevelMaskValue:
			slog.Debug(msg, "pkg-name", pkg, "level", level)
		}
	}

	glib.LogSetDefaultHandler(handler, 0)

	log.Initialize(program, debug)
	ui := ui{}
	args := []string{os.Args[0]}
	if gtkarg != "" {
		args = append(args, strings.Split(gtkarg, " ")...)
	}
	ui.Run(args)
}
