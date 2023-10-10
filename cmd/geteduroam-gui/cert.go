package main

import (
	"errors"
	"os"
	"sync"

	"github.com/geteduroam/linux-app/internal/network"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
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
	SignalPool
	win     *gtk.Window
	builder *gtk.Builder

	certPath   string
	upload     gtk.Button
	cert       string
	passphrase string
	pwd        gtk.PasswordEntry
}

func (l *CertificateState) Destroy() {
	l.DisconnectSignals()
	l.upload.Unref()
	l.pwd.Unref()
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
	return l.cert, l.pwd.GetText()
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
	l.builder.GetObject("certificateFileButton").Cast(&l.upload)
	var label gtk.Label
	l.builder.GetObject("certificateFileText").Cast(&label)
	if l.cert != "" {
		label.SetText("A certificate is already provided.\nEnter the passphrase to decrypt")
		l.upload.Hide()
	}

	l.AddSignal(&l.upload, l.upload.ConnectClicked(func(_ gtk.Button) {
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
	}))

	l.builder.GetObject("certificatePassphraseText").Cast(&l.pwd)
	l.pwd.SetText(l.passphrase)
}
