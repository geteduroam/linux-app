// Package connection implements the NetworkManager connection DBUS interface
package connection

import (
	"golang.org/x/exp/slog"

	"github.com/geteduroam/linux-app/internal/nm/base"
	"github.com/godbus/dbus/v5"
)

const (
	// Interface is the interface for a DBUS connection
	Interface = SettingsInterface + ".Connection"
	// Delete is the interface for the method to delete a DBUS connection
	Delete = Interface + ".Delete"
	// Update is the interface for the method to update a DBUS connection
	Update = Interface + ".Update"
	// GetSettings is the interface for the method to get settings for a DBUS connection
	GetSettings = Interface + ".GetSettings"
)

// Connection is a NetworkManager connection
type Connection struct {
	base.Base
}

// New creates a new NetworkManager DBUS connection
func New(path dbus.ObjectPath) (*Connection, error) {
	c := &Connection{}
	err := c.Init(base.Interface, path)
	if err != nil {
		slog.Debug("Error initiating DBus connection", "error", err)
		return nil, err
	}
	return c, nil
}

// Update updates the connection
func (c *Connection) Update(settings SettingsArgs) error {
	return c.Call(Update, settings)
}

// Delete deletes the connection
func (c *Connection) Delete() error {
	return c.Call(Delete)
}

// GetSettings gets settings for the connection
func (c *Connection) GetSettings() (SettingsArgs, error) {
	var settings map[string]map[string]dbus.Variant

	if err := c.CallReturn(&settings, GetSettings); err != nil {
		return nil, err
	}

	return decodeSettings(settings), nil
}
