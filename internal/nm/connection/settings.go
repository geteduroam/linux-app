package connection

import (
	"github.com/godbus/dbus/v5"
	"github.com/jwijenbergh/geteduroam-linux/internal/nm/base"
)

const (
	SettingsInterface  = base.Interface + ".Settings"
	SettingsObjectPath = base.ObjectPath + "/Settings"

	SettingsAddConnection = SettingsInterface + ".AddConnection"
)

type SettingsArgs map[string]map[string]interface{}

type Settings struct {
	base.Base
}

func NewSettings() (*Settings, error) {
	s := &Settings{}
	err := s.Init(base.Interface, SettingsObjectPath)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Settings) AddConnection(settings SettingsArgs) (*Connection, error) {
	var path dbus.ObjectPath
	err := s.CallReturn(&path, SettingsAddConnection, settings)
	if err != nil {
		return nil, err
	}

	return New(path)
}
