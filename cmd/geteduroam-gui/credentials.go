package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/geteduroam/linux-app/internal/network"
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
		wg:      &sync.WaitGroup{},
	}
}

type CredentialsState struct {
	builder *gtk.Builder
	cred    network.Credentials

	user gtk.Entry
	pwd  gtk.PasswordEntry
}

func (l *CredentialsState) Prefix() string {
	return "login"
}

func (l *CredentialsState) Get() (string, string) {
	return l.user.Text(), l.pwd.Text()
}

func (l *CredentialsState) Validate() error {
	ut := l.user.Text()
	if ut == "" {
		return errors.New("username cannot be empty")
	}
	if !strings.HasPrefix(ut, l.cred.Prefix) {
		return fmt.Errorf("username must begin with: \"%s\"", l.cred.Prefix)
	}
	if !strings.HasSuffix(ut, l.cred.Suffix) {
		return fmt.Errorf("username must end with: \"%s\"", l.cred.Suffix)
	}
	if l.pwd.Text() == "" {
		return errors.New("password cannot be empty")
	}
	return nil
}

func (l *CredentialsState) Initialize() {
	// TODO: Prefill suffix outside of text entry (so that it cannot be changed)
	// prefill password and username
	l.user = l.builder.GetObject("loginUsernameText").Cast().(gtk.Entry)
	l.user.SetText(l.cred.Prefix + l.cred.Suffix)

	l.pwd = l.builder.GetObject("loginPasswordText").Cast().(gtk.PasswordEntry)
	l.pwd.SetText(l.cred.Password)
}
