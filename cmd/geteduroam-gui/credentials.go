package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/geteduroam/linux-app/internal/network"
	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

func NewCredentialsStateBase(builder *gtk.Builder, stack *adw.ViewStack, cred network.Credentials, pi network.ProviderInfo) *LoginBase {
	state := CredentialsState{
		builder: builder,
		cred:    cred,
	}
	return &LoginBase{
		builder: builder,
		stack:   stack,
		state:   &state,
		pi:      pi,
	}
}

type CredentialsState struct {
	builder *gtk.Builder
	cred    network.Credentials

	user gtk.Entry
	pwd  gtk.PasswordEntry
}

func (l *CredentialsState) Destroy() {
	l.user.Unref()
	l.pwd.Unref()
}

func (l *CredentialsState) Prefix() string {
	return "login"
}

func (l *CredentialsState) Get() (string, string) {
	return l.user.GetText(), l.pwd.GetText()
}

func (l *CredentialsState) Validate() error {
	ut := l.user.GetText()
	if ut == "" {
		return errors.New("username cannot be empty")
	}
	if !strings.HasPrefix(ut, l.cred.Prefix) {
		return fmt.Errorf("username must begin with: \"%s\"", l.cred.Prefix)
	}
	if !strings.HasSuffix(ut, l.cred.Suffix) {
		return fmt.Errorf("username must end with: \"%s\"", l.cred.Suffix)
	}
	if l.pwd.GetText() == "" {
		return errors.New("password cannot be empty")
	}
	return nil
}

func (l *CredentialsState) Initialize() {
	// TODO: Prefill suffix outside of text entry (so that it cannot be changed)
	// prefill password and username
	l.builder.GetObject("loginUsernameText").Cast(&l.user)
	l.user.SetText(l.cred.Prefix + l.cred.Suffix)

	l.builder.GetObject("loginPasswordText").Cast(&l.pwd)
	l.pwd.SetText(l.cred.Password)
}
