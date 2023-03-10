package connection

import (
	"github.com/godbus/dbus/v5"
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/nm/base"
)

const (
	Interface = SettingsInterface + ".Connection"
	Update    = Interface + ".Update"
)

type Connection struct {
	base.Base
}

func New(path dbus.ObjectPath) (*Connection, error) {
	c := &Connection{}
	err := c.Init(base.Interface, path)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Connection) Update(settings Settings) error {
	return c.Call(Update, settings)
}
