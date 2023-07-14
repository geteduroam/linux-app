package connection

import (
	"golang.org/x/exp/slog"

	"github.com/geteduroam/linux-app/internal/nm/base"
	"github.com/godbus/dbus/v5"
)

const (
	Interface   = SettingsInterface + ".Connection"
	Update      = Interface + ".Update"
	GetSettings = Interface + ".GetSettings"
)

type Connection struct {
	base.Base
}

func New(path dbus.ObjectPath) (*Connection, error) {
	c := &Connection{}
	err := c.Init(base.Interface, path)
	if err != nil {
		slog.Debug("Error initiating DBus connection", "error", err)
		return nil, err
	}
	return c, nil
}

func (c *Connection) Update(settings SettingsArgs) error {
	return c.Call(Update, settings)
}

func (c *Connection) GetSettings() (SettingsArgs, error) {
	var settings map[string]map[string]dbus.Variant

	if err := c.CallReturn(&settings, GetSettings); err != nil {
		return nil, err
	}

	return decodeSettings(settings), nil
}
