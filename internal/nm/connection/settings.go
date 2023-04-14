package connection

import (
	"errors"
	"fmt"

	"github.com/geteduroam/linux-app/internal/nm/base"
	"github.com/godbus/dbus/v5"
)

const (
	SettingsInterface  = base.Interface + ".Settings"
	SettingsObjectPath = base.ObjectPath + "/Settings"

	SettingsAddConnection       = SettingsInterface + ".AddConnection"
	SettingsGetConnectionByUUID = SettingsInterface + ".GetConnectionByUuid"
)

type SettingsArgs map[string]map[string]interface{}

func (s SettingsArgs) UUID() (string, error) {
	c, ok := s["connection"]
	if !ok {
		return "", errors.New("no connection value in connection settings map")
	}
	uuid, ok := c["uuid"]
	if !ok {
		return "", errors.New("no uuid in connection map")
	}
	uuidS, ok := uuid.(string)
	if !ok {
		return "", fmt.Errorf("uuid is not a string: %T", uuid)
	}
	return uuidS, nil
}

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

func (s *Settings) ConnectionByUUID(uuid string) (*Connection, error) {
	var path dbus.ObjectPath
	err := s.CallReturn(&path, SettingsGetConnectionByUUID, uuid)
	if err != nil {
		return nil, err
	}

	return New(path)
}

func decodeSettings(input map[string]map[string]dbus.Variant) (settings SettingsArgs) {
	valueMap := SettingsArgs{}
	for key, data := range input {
		valueMap[key] = decode(data).(map[string]interface{})
	}
	return valueMap
}

func decode(input interface{}) (value interface{}) {
	switch v := input.(type) {
	case dbus.Variant:
		return decode(v.Value())
	case map[string]dbus.Variant:
		return decodeMap(v)
	case []dbus.Variant:
		return decodeArray(v)
	case []map[string]dbus.Variant:
		return decodeMapArray(v)
	default:
		return v
	}
}

func decodeArray(input []dbus.Variant) (value []interface{}) {
	for _, data := range input {
		value = append(value, decode(data))
	}
	return
}

func decodeMapArray(input []map[string]dbus.Variant) (value []map[string]interface{}) {
	for _, data := range input {
		value = append(value, decodeMap(data))
	}
	return
}

func decodeMap(input map[string]dbus.Variant) (value map[string]interface{}) {
	value = map[string]interface{}{}
	for key, data := range input {
		value[key] = decode(data)
	}
	return
}
