package base

import (
	"github.com/godbus/dbus/v5"
)

const (
	Interface  = "org.freedesktop.NetworkManager"
	ObjectPath = "/org/freedesktop/NetworkManager"
)

type Base struct {
	conn   *dbus.Conn
	object dbus.BusObject
}

func (b *Base) Init(iface string, objectPath dbus.ObjectPath) error {
	var err error

	b.conn, err = dbus.SystemBus()
	if err != nil {
		return err
	}

	b.object = b.conn.Object(iface, objectPath)

	return nil
}

func (b *Base) Call(method string, args ...interface{}) error {
	return b.object.Call(method, 0, args...).Err
}

func (b *Base) CallReturn(ret interface{}, method string, args ...interface{}) error {
	return b.object.Call(method, 0, args...).Store(ret)
}

func (b *Base) Path() dbus.ObjectPath {
	return b.object.Path()
}
