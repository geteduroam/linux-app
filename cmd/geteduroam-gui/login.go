package main

import (
	"encoding/base64"
	"fmt"
	"sync"

	"golang.org/x/exp/slog"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/geteduroam/linux-app/internal/network"
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
}

type LoginBase struct {
	builder *gtk.Builder
	stack   *adw.ViewStack
	state   LoginState
	pi      network.ProviderInfo
	wg      *sync.WaitGroup

	btn *gtk.Button
}

func (l *LoginBase) GetObject(id string, obj any) {
	fID := l.state.Prefix() + id
	g := l.builder.GetObject(fID)
	if g == nil {
		panic("no such object with id: " + fID)
	}
	obj = g
}

func (l *LoginBase) ShowError(err error) {
	slog.Error(err.Error(), "state", "login")
	var overlay *adw.ToastOverlay
	l.GetObject("ToastOverlay", overlay)
	showErrorToast(overlay, err)
}

func (l *LoginBase) Get() (string, string) {
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

	// set the title
	var title gtk.Label
	l.GetObject("InstanceTitle", &title)
	styleWidget(&title, "label")
	title.SetText(l.pi.Name)

	if l.pi.Description != "" {
		var descr gtk.Label
		l.GetObject("InstanceDescription", &descr)
		descr.SetText("Description: " + l.pi.Description)
	}

	// set logo
	var logo gtk.Image
	l.GetObject("InstanceLogo", &logo)

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
	if l.pi.Helpdesk.Email != "" {
		email.SetText(fmt.Sprintf("E-mail: %s", l.pi.Helpdesk.Email))
	} else {
		email.Hide()
	}
	var tel gtk.Label
	l.GetObject("InstanceTel", &tel)
	if l.pi.Helpdesk.Phone != "" {
		tel.SetText(fmt.Sprintf("Tel.: %s", l.pi.Helpdesk.Phone))
	} else {
		tel.Hide()
	}
	var web gtk.Label
	l.GetObject("InstanceWeb", &web)
	if l.pi.Helpdesk.Web != "" {
		web.SetText(fmt.Sprintf("Website: %s", l.pi.Helpdesk.Web))
	} else {
		web.Hide()
	}
	l.state.Initialize()
	l.btn = &gtk.Button{}
	l.GetObject("Submit", l.btn)
	l.btn.SetSensitive(true)

	submit := func() {
		l.Submit()
	}
	l.btn.ConnectClicked(submit)

	// set the page as current
	setPage(l.stack, &page)
}
