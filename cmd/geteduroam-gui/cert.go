package main

import (
	"errors"
	"os"
	"sync"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/geteduroam/linux-app/internal/network"
)

func NewCertificateStateBase(win *gtk.Window, builder *gtk.Builder, stack *adw.ViewStack, cert string, passphrase string, pi network.ProviderInfo) *LoginBase {
	state := CertificateState{
		win:        win,
		builder:    builder,
		cert:       cert,
		passphrase: passphrase,
	}
	return &LoginBase{
		builder: builder,
		stack:   stack,
		state:   &state,
		pi:      pi,
		wg:      &sync.WaitGroup{},
	}
}

type CertificateState struct {
	win     *gtk.Window
	builder *gtk.Builder

	certPath   string
	upload     *gtk.Button
	cert       string
	passphrase string
	pwd        *gtk.PasswordEntry
}

func (l *CertificateState) Prefix() string {
	return "certificate"
}

func (l *CertificateState) File() ([]byte, error) {
	if l.certPath == "" {
		return nil, errors.New("no certificate chosen")
	}
	return os.ReadFile(l.certPath)
}

func (l *CertificateState) Get() (string, string) {
	return l.cert, l.pwd.Text()
}

func (l *CertificateState) Validate() error {
	f, err := l.File()
	// only make sure this error is set if we don't have a valid cert yet
	if err != nil && l.cert == "" {
		return err
	}
	if f != nil {
		l.cert = string(f)
	}
	return nil
}

func (l *CertificateState) Initialize() {
	l.upload = l.builder.GetObject("certificateFileButton").Cast().(*gtk.Button)
	label := l.builder.GetObject("certificateFileText").Cast().(*gtk.Label)
	if l.cert != "" {
		label.SetText("A certificate is already provided.\nEnter the passphrase to decrypt")
		l.upload.Hide()
	}

	clicked := func() {
		// Create a file dialog
		fd, err := NewFileDialog(l.win, "Choose a PKCS12 client certificate")
		if err != nil {
			// TODO: handle error
			panic(err)
		}

		fd.Run(func(p string) {
			l.certPath = p
			label.SetText(l.certPath)
		})
	}

	l.upload.ConnectClicked(clicked)

	l.pwd = l.builder.GetObject("certificatePassphraseText").Cast().(*gtk.PasswordEntry)
	l.pwd.SetText(l.passphrase)
}
