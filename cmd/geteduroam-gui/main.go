//go:debug x509negativeserial=1
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/slog"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/geteduroam/linux-app/internal/clientver"
	"github.com/geteduroam/linux-app/internal/discovery"
	"github.com/geteduroam/linux-app/internal/handler"
	"github.com/geteduroam/linux-app/internal/logwrap"
	"github.com/geteduroam/linux-app/internal/network"
	"github.com/geteduroam/linux-app/internal/provider"
	"github.com/geteduroam/linux-app/internal/variant"
)

type serverList struct {
	sync.Mutex
	store     *gtk.StringList
	providers provider.Providers
	list      *SelectList
	custom    bool
}

func (s *serverList) get(idx int, query string) (*provider.Provider, error) {
	if s.custom && idx == len(s.providers) {
		// TODO: add context
		return provider.Custom(context.Background(), query)
	}
	if idx < 0 || idx > len(s.providers) {
		return nil, errors.New("index out of range")
	}
	return &s.providers[idx], nil
}

func (s *serverList) getNames(idx int, query string) (*provider.LocalizedStrings, error) {
	if s.custom && idx == len(s.providers) {
		return &provider.LocalizedStrings{{Display: query}}, nil
	}
	if idx < 0 || idx > len(s.providers) {
		return nil, errors.New("index out of range")
	}
	return &s.providers[idx].Name, nil
}

func (s *serverList) Fill() {
	s.Lock()
	defer s.Unlock()
	for idx, inst := range s.providers {
		s.list.Add(idx, inst.Name.Get())
	}
}

func (s *serverList) AddCustom(label string) {
	if s.custom {
		s.RemoveCustom()
	}
	s.list.Add(len(s.providers), label)
	s.custom = true
}

func (s *serverList) RemoveCustom() {
	if !s.custom {
		return
	}
	s.list.Remove(len(s.providers))
	s.custom = false
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
	m.servers.store = gtk.NewStringList(nil)
}

func (m *mainState) activate() {
	page := m.builder.GetObject("searchPage").Cast().(*adw.ViewStackPage)
	// we do not use setPage here as the margin is already set in the clamp
	// This is so that the background is set on the current main page only but it expands to each side fully
	m.stack.SetVisibleChild(page.Child())
}

func (m *mainState) askCredentials(c network.Credentials, pi network.ProviderInfo) (string, string, error) {
	login := NewCredentialsStateBase(m.builder, m.stack, c, pi)
	login.Initialize()
	user, pass := login.Get()
	return user, pass, nil
}

func (m *mainState) askCertificate(cert string, pwd string, pi network.ProviderInfo) (string, string, error) {
	base := NewCertificateStateBase(m.app.ActiveWindow(), m.builder, m.stack, cert, pwd, pi)
	base.Initialize()
	cert, pass := base.Get()
	return cert, pass, nil
}

func (m *mainState) file(metadata []byte) (*time.Time, *time.Time, error) {
	h := handler.Handlers{
		CredentialsH: m.askCredentials,
		CertificateH: m.askCertificate,
	}
	return h.Configure(metadata)
}

func (m *mainState) direct(p provider.Profile) error {
	config, err := p.EAPDirect()
	if err != nil {
		return err
	}
	_, _, err = m.file(config)
	return err
}

func (m *mainState) local(path string) (*time.Time, *time.Time, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	vBeg, vEnd, err := m.file(b)
	if err != nil {
		return nil, nil, err
	}
	return vBeg, vEnd, nil
}

func (m *mainState) oauth(ctx context.Context, p provider.Profile) (*time.Time, *time.Time, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	config, err := p.EAPOAuth(ctx, func(url string) {
		uiThread(func() {
			l := NewLoadingPage(m.builder, m.stack, "Your browser has been opened to authorize the client", func() {
				cancel()
			})
			l.Initialize()
			// If the browser does not open for some reason the user could grab it with stdout
			// We could also show it in the UI but this might mean too much clutter
			fmt.Println("Browser has been opened with URL:", url)
		})
	})
	if err != nil {
		return nil, nil, err
	}

	return m.file(config)
}

func (m *mainState) rowActivated(sel provider.Provider) {
	l := NewLoadingPage(m.builder, m.stack, "Loading organization details...", nil)
	l.Initialize()
	ctx := context.Background()
	chosen := func(p provider.Profile) (err error) {
		defer func() {
			err = ensureContextError(ctx, err)
		}()
		var vBeg *time.Time
		var vEnd *time.Time
		var isredirect bool
		switch p.Flow() {
		case provider.DirectFlow:
			err = m.direct(p)
			if err != nil {
				return err
			}
		case provider.OAuthFlow:
			vBeg, vEnd, err = m.oauth(ctx, p)
			if err != nil {
				return err
			}
		case provider.RedirectFlow:
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
		s := NewSuccessState(m.builder, m.app.ActiveWindow(), m.stack, vBeg, vEnd, isredirect)
		uiThread(func() {
			s.Initialize()
		})
		return nil
	}
	cb := func(p provider.Profile) {
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
	list := m.builder.GetObject("searchList").Cast().(*gtk.ListView)

	cache := discovery.NewCache()
	inst, err := cache.Providers()
	if err != nil {
		m.ShowError(err)
		return
	}
	m.servers.providers = *inst

	search := m.builder.GetObject("searchBox").Cast().(*gtk.SearchEntry)
	activated := func(idx int) {
		l := NewLoadingPage(m.builder, m.stack, "Loading server details...", nil)
		l.Initialize()
		cb := func(inst *provider.Provider, err error) {
			if err != nil {
				l.Hide()
				m.activate()
				m.ShowError(err)
				return
			}
			m.rowActivated(*inst)
		}
		go func() {
			inst, err := m.servers.get(idx, search.Text())
			uiThread(func() {
				cb(inst, err)
			})
		}()
	}

	sorter := func(a, b int) int {
		query := search.Text()
		n1, err := m.servers.getNames(a, query)
		if err != nil {
			return -1
		}
		n2, err := m.servers.getNames(b, query)
		if err != nil {
			return -1
		}
		return provider.SortNames(*n1, *n2, query)
	}

	m.servers.list = NewSelectList(m.scroll, list, activated, sorter).WithFiltering(func(idx int) bool {
		query := search.Text()
		n, err := m.servers.getNames(idx, query)
		if err != nil {
			return false
		}
		return provider.FilterSingle(*n, query)
	})

	// Fill the servers in the select list
	m.servers.Fill()

	// Further set up the list
	m.servers.list.Setup()

	changedcb := func() {
		// TODO len returns length in bytes
		// utf8.RuneCountInString() counts number of characters (runes)
		query := search.Text()
		if len(query) <= 2 {
			m.servers.list.Hide()
			return
		}
		// url entered
		if strings.Count(query, ".") >= 2 {
			m.servers.AddCustom(search.Text())
		} else {
			m.servers.RemoveCustom()
		}
		m.servers.list.Changed()
		m.servers.list.Show()
	}

	// Update the list when searching
	search.ConnectSearchChanged(changedcb)
}

func (m *mainState) localMetadata() {
	fd, err := NewFileDialog(m.app.ActiveWindow(), "Choose an EAP metadata file")
	if err != nil {
		m.ShowError(err)
		return
	}
	fd.Run(func(path string) {
		go func() {
			vBeg, vEnd, err := m.local(path)
			if err != nil {
				uiThread(func() {
					m.activate()
					m.ShowError(err)
				})
				return
			}
			s := NewSuccessState(m.builder, m.app.ActiveWindow(), m.stack, vBeg, vEnd, false)
			s.Initialize()
		}()
	})
}

func (m *mainState) initBurger() {
	gears := m.builder.GetObject("gears").Cast().(*gtk.MenuButton)

	builder := gtk.NewBuilderFromString(MustResource("gears.ui"))
	menu := builder.GetObject("menu").Cast().(*gio.Menu)
	gears.SetMenuModel(menu)

	imp := gio.NewSimpleAction("import-local", nil)
	actcb := func(_ *glib.Variant) {
		m.localMetadata()
	}
	imp.ConnectActivate(actcb)

	aboutcb := func(_ *glib.Variant) {
		awin := gtk.NewAboutDialog()
		awin.SetName(fmt.Sprintf("%s Linux client", variant.DisplayName))
		pb, err := bytesPixbuf([]byte(MustResource("images/heart.png")))
		if err == nil {
			texture := gdk.NewTextureForPixbuf(pb)
			awin.SetLogo(texture)
		}
		lpath, err := logwrap.Location(fmt.Sprintf("%s-gui", variant.DisplayName))
		if err == nil {
			awin.SetSystemInformation("Log location: " + lpath)
		}
		awin.SetProgramName(fmt.Sprintf("%s GUI", variant.DisplayName))
		awin.SetComments(fmt.Sprintf("Client to easily and securely configure %s", variant.ProfileName))
		awin.SetAuthors([]string{"Jeroen Wijenbergh", "Martin van Es", "Alexandru Cacean"})
		awin.SetVersion(clientver.Get())
		awin.SetWebsite("https://github.com/geteduroam/linux-app")
		// SetLicenseType has a scary warning: "comes with absolutely no warranty"
		// While it is true according to the license, I find it unfriendly
		awin.SetLicense("This application has a BSD 3 license.")
		awin.SetTransientFor(m.app.ActiveWindow())
		awin.Show()
	}

	about := gio.NewSimpleAction("about", nil)
	about.ConnectActivate(aboutcb)

	m.app.AddAction(imp)
	m.app.AddAction(about)
}

func (m *mainState) Initialize() {
	m.scroll = m.builder.GetObject("searchScroll").Cast().(*gtk.ScrolledWindow)
	m.stack = m.builder.GetObject("pageStack").Cast().(*adw.ViewStack)
	m.initServers()
	m.initList()
	m.initBurger()
	m.activate()
}

func (m *mainState) ShowError(err error) {
	if errors.Is(err, context.Canceled) {
		return
	}
	slog.Error(err.Error(), "state", "main")
	overlay := m.builder.GetObject("searchToastOverlay").Cast().(*adw.ToastOverlay)
	showErrorToast(overlay, err)
}

type ui struct {
	builder *gtk.Builder
	app     *adw.Application
}

func (ui *ui) initBuilder() {
	// open the builder
	ui.builder = gtk.NewBuilderFromString(MustResource("main.ui"))
}

func (ui *ui) initWindow() {
	// get the window
	win := ui.builder.GetObject("mainWindow").Cast().(*adw.Window)
	win.SetTitle(fmt.Sprintf("%s GUI", variant.DisplayName))
	win.SetDefaultSize(400, 600)
	// style the window using the css
	search := ui.builder.GetObject("searchPage").Cast().(*adw.ViewStackPage)
	child := search.Child().(*adw.ToastOverlay)
	styleWidget(child, fmt.Sprintf("window_%s", variant.DisplayName))
	ui.app.AddWindow(&win.Window)
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
	ui.app = adw.NewApplication(variant.AppID, gio.ApplicationDefaultFlags)
	actcb := func() {
		ui.activate()
	}
	ui.app.ConnectActivate(actcb)

	return ui.app.Run(args)
}

func main() {
	const usage = `Usage of %s:
  -h, --help			Prints this help information
  --version			Prints version information
  -d, --debug			Debug
  --gtk-args                    Arguments to pass to gtk as a string, e.g. "--help". These flags are split on spaces

  This GUI binary is used to add an eduroam connection profile with integration using NetworkManager and Gtk.

  Log file location: %s
`

	var help bool
	var versionf bool
	var debug bool
	var gtkarg string
	program := fmt.Sprintf("%s-gui", variant.DisplayName)
	lpath, err := logwrap.Location(program)
	if err != nil {
		lpath = "N/A"
	}
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&versionf, "version", false, "Show version")
	flag.BoolVar(&debug, "d", false, "Debug")
	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.StringVar(&gtkarg, "gtk-args", "", "Gtk arguments")
	flag.Usage = func() { fmt.Printf(usage, program, lpath) }
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	if versionf {
		fmt.Println(clientver.Get())
		return
	}

	//var handler glib.LogFunc = func(pkg string, level glib.LogLevelFlags, msg string, _ uintptr) {
	//	switch level {
	//	case glib.GLogLevelErrorValue:
	//		slog.Error(msg, "pkg-name", pkg, "level", level)
	//	case glib.GLogLevelCriticalValue:
	//		// Ignore some false positives due to Gtk bug
	//		// Happens when pressing "Import Metadata"
	//		// see https://discourse.gnome.org/t/menu-button-gives-error-messages-with-latest-gtk4/15689/3
	//		ignore := "_gtk_css_corner_value_get_%s: assertion 'corner->class == &GTK_CSS_VALUE_CORNER' failed"
	//		if fmt.Sprintf(ignore, "x") == msg || fmt.Sprintf(ignore, "y") == msg {
	//			return
	//		}
	//		slog.Error("pkg-name", pkg, "level", level)
	//	case glib.GLogLevelWarningValue:
	//		slog.Warn(msg, "pkg-name", pkg, "level", level)
	//	case glib.GLogLevelMessageValue:
	//		slog.Info(msg, "pkg-name", pkg, "level", level)
	//	case glib.GLogLevelInfoValue:
	//		slog.Info(msg, "pkg-name", pkg, "level", level)
	//	case glib.GLogLevelDebugValue:
	//		slog.Debug(msg, "pkg-name", pkg, "level", level)
	//	case glib.GLogLevelMaskValue:
	//		slog.Debug(msg, "pkg-name", pkg, "level", level)
	//	}
	//}

	//glib.LogDefaultHandler(&handler, 0)

	logwrap.Initialize(program, debug)
	ui := ui{}
	args := []string{os.Args[0]}
	if gtkarg != "" {
		args = append(args, strings.Split(gtkarg, " ")...)
	}
	ui.Run(args)
}
