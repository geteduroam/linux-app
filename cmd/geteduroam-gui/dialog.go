package main

import (
	"errors"

	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type FileDialog struct {
	*gtk.FileChooserDialog
	win *gtk.Window
}

func NewFileDialog(parent *gtk.Window, label string) (*FileDialog, error) {
	fwidg := gtk.NewFileChooserDialog(
		label,
		parent,
		gtk.FileChooserActionOpenValue,
		"Cancel",
		gtk.ResponseCancelValue,
		"Select",
		gtk.ResponseAcceptValue,
		0,
	)
	if fwidg == nil {
		return nil, errors.New("file chooser dialog could not be initialized")
	}
	var fc gtk.FileChooserDialog
	fwidg.Cast(&fc)
	var fwin gtk.Window
	fwidg.Cast(&fwin)
	return &FileDialog{
		FileChooserDialog: &fc,
		win:               &fwin,
	}, nil
}

func (fd *FileDialog) Run(cb func(path string)) {
	fd.ConnectResponse(func(_ gtk.Dialog, res int) {
		// TODO: int32 casting is a puregotk bug? gint should be int32 but I think it someties is a normal int
		if int32(res) == int32(gtk.ResponseAcceptValue) {
			f := fd.GetFile()
			cb(f.GetPath())
		}
		fd.win.Destroy()
	})
	fd.Present()
}
