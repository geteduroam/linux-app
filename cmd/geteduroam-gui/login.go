package main

import (
	"encoding/base64"
	"fmt"
	"sync"

	"golang.org/x/exp/slog"

	"github.com/geteduroam/linux-app/internal/network"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"

)

type LoginState interface {
	// Get returns the information needed
	// In case of credentials: username + password
	// In case of certificates: pkcs12 + passphrase
	Get() (string, string)

	// Initialize initializes the builder components
	Initialize()

	// Validate validates the user input
	// This is called before pressing the submit button
	Validate() error

	// Prefix returns the GTK builder prefix that is used to get each item
	Prefix() string

	// Destroy destroys any resources that should be called when the view is hidden
	Destroy()
}

type LoginBase struct {
	SignalPool
	builder *gtk.Builder
	stack   *adw.ViewStack
	state   LoginState
	pi      network.ProviderInfo
	wg      *sync.WaitGroup

	btn *gtk.Button
}

func (l *LoginBase) GetObject(id string, obj gobject.Ptr) {
	fID := l.state.Prefix() + id
	g := l.builder.GetObject(fID)
	if g == nil {
		panic("no such object with id: " + fID)
	}
	g.Cast(obj)
}

func (l *LoginBase) ShowError(err error) {
	slog.Error(err.Error(), "state", "login")
	var overlay adw.ToastOverlay
	l.GetObject("ToastOverlay", &overlay)
	defer overlay.Unref()
	showErrorToast(overlay, err)
}

func (l *LoginBase) Destroy() {
	l.DisconnectSignals()
	l.btn.Unref()
}

func (l *LoginBase) Get() (string, string) {
	defer l.state.Destroy()
	defer l.Destroy()
	l.wg.Wait()
	return l.state.Get()
}

func (l *LoginBase) Submit() {
	if err := l.state.Validate(); err != nil {
		l.ShowError(err)
		return
	}
	defer l.wg.Done()
	l.btn.SetSensitive(false)
}

func (l *LoginBase) fillLogo(logo *gtk.Image) error {
	d, err := base64.StdEncoding.DecodeString(l.pi.Logo)
	if err != nil {
		return err
	}
	pb, err := bytesPixbuf(d)
	if err == nil {
		uiThread(func() {
			defer logo.Unref()
			logo.SetFromPixbuf(pb)
			logo.SetSizeRequest(100, 100)
		})
	}
	return nil
}

func (l *LoginBase) Initialize() {
	l.wg.Add(1)
	var page adw.ViewStackPage
	l.GetObject("Page", &page)
	defer page.Unref()

	// set the title
	var title gtk.Label
	l.GetObject("InstanceTitle", &title)
	defer title.Unref()
	styleWidget(&title, "label")
	title.SetText(l.pi.Name)

	if l.pi.Description != "" {
		var descr gtk.Label
		l.GetObject("InstanceDescription", &descr)
		defer descr.Unref()
		descr.SetText("Description: " + l.pi.Description)
	}

	// set logo
	var logo gtk.Image
	l.GetObject("InstanceLogo", &logo)
	defer logo.Unref()

	if l.pi.Logo != "" {
		err := l.fillLogo(&logo)
		// TODO: do not panic here but just log
		if err != nil {
			panic(err)
		}
	} else {
		logo.Hide()
	}
	// set the contact
	var email gtk.Label
	l.GetObject("InstanceEmail", &email)
	defer email.Unref()
	if l.pi.Helpdesk.Email != "" {
		email.SetText(fmt.Sprintf("E-mail: %s", l.pi.Helpdesk.Email))
	} else {
		email.Hide()
	}
	var tel gtk.Label
	l.GetObject("InstanceTel", &tel)
	defer tel.Unref()
	if l.pi.Helpdesk.Phone != "" {
		tel.SetText(fmt.Sprintf("Tel.: %s", l.pi.Helpdesk.Phone))
	} else {
		tel.Hide()
	}
	var web gtk.Label
	l.GetObject("InstanceWeb", &web)
	defer web.Unref()
	if l.pi.Helpdesk.Web != "" {
		web.SetText(fmt.Sprintf("Website: %s", l.pi.Helpdesk.Web))
	} else {
		web.Hide()
	}
	l.state.Initialize()
	l.btn = &gtk.Button{}
	l.GetObject("Submit", l.btn)
	l.btn.SetSensitive(true)
	l.AddSignal(l.btn, l.btn.ConnectSignal("clicked", func() {
		l.Submit()
	}))

	// set the page as current
	l.stack.SetVisibleChild(page.GetChild())
}
