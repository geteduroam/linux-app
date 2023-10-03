package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/geteduroam/linux-app/internal/network"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gdkpixbuf"
	"github.com/jwijenbergh/puregotk/v4/gtk"
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

func (l *LoginState) ShowError(err error) {
	toast := adw.NewToast(err.Error())
	toast.SetTimeout(5)
	var overlay adw.ToastOverlay
	l.builder.GetObject("loginToastOverlay").Cast(&overlay)
	defer overlay.Unref()
	overlay.AddToast(toast)
}

func (l *LoginState) Get() (string, string) {
	defer l.user.Unref()
	defer l.pwd.Unref()
	defer l.btn.Unref()
	l.wg.Wait()

	return l.user.GetText(), l.pwd.GetText()
}

func (l *LoginState) Validate() bool {
	ut := l.user.GetText()
	if ut == "" {
		l.ShowError(errors.New("username cannot be empty"))
		return false
	}
	if !strings.HasPrefix(ut, l.cred.Prefix) {
		l.ShowError(fmt.Errorf("username must begin with: \"%s\"", l.cred.Prefix))
		return false
	}
	if !strings.HasSuffix(ut, l.cred.Suffix) {
		l.ShowError(fmt.Errorf("username must end with: \"%s\"", l.cred.Suffix))
		return false
	}
	if l.pwd.GetText() == "" {
		l.ShowError(errors.New("password cannot be empty"))
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

func (l *LoginState) fillLogo(logo *gtk.Image) error {
	d, err := base64.StdEncoding.DecodeString(l.pi.Logo)
	if err != nil {
		return err
	}
	// TODO: do this without creating a temp file
	f, err := os.CreateTemp("/tmp", "geteduroam-linux-instance-logo")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(d)
	if err != nil {
		return err
	}
	pb, err := gdkpixbuf.NewFromFilePixbuf(f.Name())
	if err != nil {
		return err
	}
	uiThread(func() {
		defer logo.Unref()
		logo.SetFromPixbuf(pb)
		logo.SetSizeRequest(100, 100)
	})
	return nil
}

func (l *LoginState) Initialize() {
	l.wg.Add(1)
	var page adw.ViewStackPage
	l.builder.GetObject("loginPage").Cast(&page)
	defer page.Unref()

	// set the title
	var title gtk.Label
	l.builder.GetObject("instanceTitle").Cast(&title)
	defer title.Unref()
	styleWidget(&title, "label")
	title.SetText(l.pi.Name)

	// set logo
	var logo gtk.Image
	l.builder.GetObject("instanceLogo").Cast(&logo)
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
	l.builder.GetObject("instanceEmail").Cast(&email)
	defer email.Unref()
	if l.pi.Helpdesk.Email != "" {
		email.SetText(fmt.Sprintf("E-mail: %s", l.pi.Helpdesk.Email))
	} else {
		email.Hide()
	}
	var tel gtk.Label
	l.builder.GetObject("instanceTel").Cast(&tel)
	defer tel.Unref()
	if l.pi.Helpdesk.Phone != "" {
		tel.SetText(fmt.Sprintf("Tel.: %s", l.pi.Helpdesk.Phone))
	} else {
		tel.Hide()
	}
	var web gtk.Label
	l.builder.GetObject("instanceWeb").Cast(&web)
	defer web.Unref()
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
}
