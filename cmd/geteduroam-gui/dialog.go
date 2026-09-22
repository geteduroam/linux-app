package main

import (
	"context"
	"errors"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type FileDialog struct {
	*gtk.FileDialog
	Parent *gtk.Window
}

func NewFileDialog(parent *gtk.Window, label string) (*FileDialog, error) {
	fc := gtk.NewFileDialog()
	if fc == nil {
		return nil, errors.New("file chooser dialog could not be initialized")
	}
	fc.SetAcceptLabel("Select")
	return &FileDialog{
		FileDialog: fc,
		Parent:     parent,
	}, nil
}

func (fd *FileDialog) Run(cb func(path string)) {
	// TODO: what context here
	fd.Open(context.Background(), fd.Parent, func(res gio.AsyncResulter) {
		file, err := fd.OpenFinish(res)
		if err != nil || file == nil {
			return
		}
		path := file.Path()
		if path == "" {
			return
		}
		cb(path)
	})
}
