// Package base implements the base of the DBUS connection with NetworkManager
package base

import (
	"github.com/godbus/dbus/v5"
)

const (
	// Interface is the DBUS interface for NetworkManager
	Interface = "org.freedesktop.NetworkManager"
	// ObjectPath is the DBUS Object Path for NetworkManager
	ObjectPath = "/org/freedesktop/NetworkManager"
)

// Base is the base DBUS connection for NetworkManager
type Base struct {
	conn   *dbus.Conn
	object dbus.BusObject
}

// Init initializes NetworkManager dbus connection
func (b *Base) Init(iface string, objectPath dbus.ObjectPath) error {
	var err error

	b.conn, err = dbus.SystemBus()
	if err != nil {
		return err
	}

	b.object = b.conn.Object(iface, objectPath)

	return nil
}

// Call calls a DBUS method
func (b *Base) Call(method string, args ...interface{}) error {
	return b.object.Call(method, 0, args...).Err
}

// CallReturn calls a DBUS method and returns data
func (b *Base) CallReturn(ret interface{}, method string, args ...interface{}) error {
	return b.object.Call(method, 0, args...).Store(ret)
}

// Path returns the DBUS object path
func (b *Base) Path() dbus.ObjectPath {
	return b.object.Path()
}
