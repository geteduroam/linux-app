package main

import (
	"encoding/base64"
	"fmt"
	"github.com/geteduroam/linux-app/internal/network"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gdkpixbuf"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"os"
	"strings"
	"sync"
)

type LoginState struct {
	builder *gtk.Builder
	stack   *adw.ViewStack
	cred    network.Credentials
	pi      network.ProviderInfo
	wg      sync.WaitGroup

	user gtk.Entry
	pwd  gtk.PasswordEntry
	btn  gtk.Button
}

func NewLoginState(builder *gtk.Builder, stack *adw.ViewStack, cred network.Credentials, pi network.ProviderInfo) *LoginState {
	return &LoginState{
		builder: builder,
		stack:   stack,
		cred:    cred,
		pi:      pi,
	}
}

func (l *LoginState) ShowError(msg string) {
	toast := adw.NewToast(msg)
	toast.SetTimeout(5)
	var overlay adw.ToastOverlay
	l.builder.GetObject("loginToastOverlay").Cast(&overlay)
	overlay.AddToast(toast)
}

func (l *LoginState) Get() (string, string) {
	l.wg.Wait()

	return l.user.GetText(), l.pwd.GetText()
}

func (l *LoginState) Validate() bool {
	ut := l.user.GetText()
	if ut == "" {
		l.ShowError("Username cannot be empty")
		return false
	}
	if !strings.HasPrefix(ut, l.cred.Prefix) {
		l.ShowError(fmt.Sprintf("Username must begin with: \"%s\"", l.cred.Prefix))
		return false
	}
	if !strings.HasSuffix(ut, l.cred.Suffix) {
		l.ShowError(fmt.Sprintf("Username must end with: \"%s\"", l.cred.Suffix))
		return false
	}
	if l.pwd.GetText() == "" {
		l.ShowError("Password cannot be empty")
		return false
	}
	return true
}

func (l *LoginState) Submit() {
	if !l.Validate() {
		return
	}
	l.btn.SetSensitive(false)
	l.wg.Done()
}

func (l *LoginState) Initialize() error {
	l.wg.Add(1)
	var page adw.ViewStackPage
	l.builder.GetObject("loginPage").Cast(&page)

	// set the title
	var title gtk.Label
	l.builder.GetObject("instanceTitle").Cast(&title)
	styleWidget(&title, "label")
	title.SetText(l.pi.Name)

	// set logo
	var logo gtk.Image
	l.builder.GetObject("instanceLogo").Cast(&logo)

	if l.pi.Logo != "" {
		d, err := base64.StdEncoding.DecodeString(l.pi.Logo)
		if err != nil {
			panic(err)
		}
		if err == nil {
			// TODO: do this without creating a temp file
			f, err := os.CreateTemp("/tmp", "geteduroam-linux-instance-logo")
			if err != nil {
				panic(err)
			}
			if err == nil {
				defer os.Remove(f.Name())
				f.Write(d)
				pb := gdkpixbuf.NewFromFilePixbuf(f.Name())
				uiThread(func() {
					logo.SetFromPixbuf(pb)
					logo.SetSizeRequest(100, 100)
				})
			}
		}
	} else {
		logo.Hide()
	}
	// set the contact
	var email gtk.Label
	l.builder.GetObject("instanceEmail").Cast(&email)
	if l.pi.Helpdesk.Email != "" {
		email.SetText(fmt.Sprintf("E-mail: %s", l.pi.Helpdesk.Email))
	} else {
		email.Hide()
	}
	var tel gtk.Label
	l.builder.GetObject("instanceTel").Cast(&tel)
	if l.pi.Helpdesk.Phone != "" {
		tel.SetText(fmt.Sprintf("Tel.: %s", l.pi.Helpdesk.Phone))
	} else {
		tel.Hide()
	}
	var web gtk.Label
	l.builder.GetObject("instanceWeb").Cast(&web)
	if l.pi.Helpdesk.Web != "" {
		web.SetText(fmt.Sprintf("Website: %s", l.pi.Helpdesk.Web))
	} else {
		web.Hide()
	}
	// prefill password and username
	l.builder.GetObject("usernameText").Cast(&l.user)
	l.user.SetText(l.cred.Prefix + l.cred.Suffix)

	l.builder.GetObject("passwordText").Cast(&l.pwd)
	l.pwd.SetText(l.cred.Password)

	l.builder.GetObject("submitLogin").Cast(&l.btn)
	l.btn.SetSensitive(true)
	l.btn.ConnectClicked(func(_ gtk.Button) {
		l.Submit()
	})

	// set the page as current
	l.stack.SetVisibleChild(page.GetChild())
	return nil
}
