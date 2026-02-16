// Package nm implements the functions needed to configure the connection using NetworkManager
package nm

import (
	"errors"
	"fmt"
	"os/user"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/exp/slog"

	"github.com/geteduroam/linux-app/internal/config"
	"github.com/geteduroam/linux-app/internal/network"
	"github.com/geteduroam/linux-app/internal/network/method"
	"github.com/geteduroam/linux-app/internal/nm/connection"
	"github.com/geteduroam/linux-app/internal/variant"
)

// encodePath encodes a string to a path expected by NetworkManager
// This path is prefixed with file:// and is explicitly NULL terminated
// It returns this path as a byte array
func encodePath(p string) []byte {
	// get the converted path
	// see: https://github.com/NetworkManager/NetworkManager/blob/main/examples/python/dbus/add-wifi-eap-connection.py#L12
	// \x00 is just NUL termination which NM expects
	c := fmt.Sprintf("file://%s\x00", p)
	return []byte(c)
}

// encodeFileBytes creates a file in the config directory with name `name` and contents `contents`
// it ensures that the path is encoded the way NetworkManager expects it to be
func encodeFileBytes(name string, contents []byte) ([]byte, error) {
	p, err := config.WriteFile(name, contents)
	if err != nil {
		slog.Debug("Error writing file", "error", err)
		return nil, err
	}
	return encodePath(p), nil
}

// PreviousCon gets a connection object using the previous UUID
func PreviousCon(pUUID string) (*connection.Connection, error) {
	if pUUID == "" {
		return nil, errors.New("UUID is empty")
	}
	s, err := connection.NewSettings()
	if err != nil {
		slog.Debug("Error creating new settings", "error", err)
		return nil, err
	}
	return s.ConnectionByUUID(pUUID)
}

// createCon creates a new connection using the arguments
// if a previous connection was found with pUUID, it updates that one instead
// it returns the newly created or updated connection object
func createCon(pUUID string, args connection.SettingsArgs) (*connection.Connection, error) {
	prev, err := PreviousCon(pUUID)
	// previous connection found, update it with the new settings args
	if err == nil {
		return prev, prev.Update(args)
	}
	// create a connection settings object
	s, err := connection.NewSettings()
	if err != nil {
		slog.Debug("Error creating new settings", "error", err)
		return nil, err
	}
	// create a new connection
	return s.AddConnection(args)
}

// installBase contains the code for creating a network with NetworkManager
// This contains the shared network settings between TLS and NonTLS
// The specific 8021x settings are given as an argument `specific`
func installBaseSSID(n network.Base, ssid network.SSID, specifics map[string]interface{}, pUUID string) (string, error) {
	fID := fmt.Sprintf("%s (from %s)", ssid.Value, variant.DisplayName)
	cUser, err := user.Current()
	if err != nil {
		return "", err
	}
	sCon := map[string]interface{}{
		// the priority is 1, just above the default 0
		// such that connections for existing eduroam profiles (and default priority)
		// will not be used
		"autoconnect-priority": 1,
		"permissions": []string{
			fmt.Sprintf("user:%s", cUser.Username),
		},
		"type": "802-11-wireless",
		"id":   fID,
	}
	sWifi := map[string]interface{}{
		"ssid":     []byte(ssid.Value),
		"security": "802-11-wireless-security",
	}
	sWsec := map[string]interface{}{
		"key-mgmt": "wpa-eap",
		"proto":    []string{"rsn"},
		"pairwise": []string{strings.ToLower(ssid.MinRSN)},
		"group":    []string{strings.ToLower(ssid.MinRSN)},
	}
	sIP4 := map[string]interface{}{
		"method": "auto",
	}
	sIP6 := map[string]interface{}{
		"method": "auto",
	}
	var sids []string

	for _, sid := range n.ServerIDs {
		v := fmt.Sprintf("DNS:%s", sid)
		sids = append(sids, v)
	}
	caBasePath, err := config.Directory()
	if err != nil {
		return "", err
	}
	err = n.Certs.ToDir(caBasePath)
	if err != nil {
		return "", err
	}
	s8021x := map[string]interface{}{
		"ca-path":            filepath.Join(caBasePath, "ca"),
		"altsubject-matches": sids,
	}
	// add the network specific settings
	for k, v := range specifics {
		s8021x[k] = v
	}

	settings := map[string]map[string]interface{}{
		"connection":               sCon,
		"802-11-wireless":          sWifi,
		"802-11-wireless-security": sWsec,
		"802-1x":                   s8021x,
		"ipv4":                     sIP4,
		"ipv6":                     sIP6,
	}
	con, err := createCon(pUUID, settings)
	if err != nil {
		return "", err
	}
	// get the settings from the added connection
	gs, err := con.GetSettings()
	if err != nil {
		return "", err
	}
	uuid, err := gs.UUID()
	if err != nil {
		return "", err
	}
	return uuid, nil
}

// installBase contains the code for creating a network with NetworkManager
// This contains the shared network settings between TLS and NonTLS
// The specific 8021x settings are given as an argument `specific`
// It loops through all SSIDs and creates different networks for each
func installBase(n network.Base, specifics map[string]interface{}, pUUIDs []string) ([]string, error) {
	// get  a mapping from ssids to the accompanying uuid
	ssidMap := make(map[string]string)
	for _, puuid := range pUUIDs {
		con, err := PreviousCon(puuid)
		if err != nil {
			slog.Debug("failed getting previous con for UUID map", "error", err)
			continue
		}
		settings, err := con.GetSettings()
		if err != nil {
			slog.Debug("failed getting settings con for UUID map", "error", err)
			continue
		}
		ssid, err := settings.SSID()
		if err != nil {
			slog.Debug("failed getting ssid from settings con for UUID map", "error", err)
			continue
		}
		if ssid != "" {
			ssidMap[ssid] = puuid
		}
	}

	var added []string

	// remove connections no longer needed
	defer func() {
		for ssid, puuid := range ssidMap {
			if slices.Contains(added, puuid) {
				continue
			}
			slog.Debug("connection does not contain previous UUID, removing the connection", "ssid", ssid, "uuid", puuid, "added", added)
			con, err := PreviousCon(puuid)
			if err == nil {
				err := con.Delete()
				if err != nil {
					slog.Debug("failed to delete connection", "error", err)
				}
			} else {
				slog.Debug("previous connection does not exist, not removing", "error", err)
			}
		}
	}()

	// add new connections
	for _, ssid := range n.SSIDs {
		uuid := ssidMap[ssid.Value]
		guuid, err := installBaseSSID(n, ssid, specifics, uuid)
		if err != nil {
			return added, err
		}
		added = append(added, guuid)
	}

	return added, nil
}

// Install installs a non TLS network and returns an error if it cannot configure it
// Right now it adds a new profile that is not automatically added
// It returns the uuid if the connection was added successfully
func Install(n network.NonTLS, pUUIDs []string) ([]string, error) {
	s8021x := map[string]interface{}{
		"eap": []string{
			n.Method().String(),
		},
		"anonymous-identity": n.AnonIdentity,
		"identity":           n.Credentials.Username,
		"password":           n.Credentials.Password,
		"password-flags":     0,
	}
	if n.InnerAuth.EAP() && n.MethodType == method.TTLS {
		s8021x["phase2-autheap"] = n.InnerAuth.String()
	} else {
		s8021x["phase2-auth"] = n.InnerAuth.String()
	}
	return installBase(n.Base, s8021x, pUUIDs)
}

// InstallTLS installs a TLS network and returns an error if it cannot configure it
// Right now it adds a new profile that is not automatically added
// It returns the uuid if the connection was added successfully
func InstallTLS(n network.TLS, pUUIDs []string) ([]string, error) {
	ccFile, err := encodeFileBytes("client-cert.pem", n.ClientCert.ToPEM())
	if err != nil {
		return nil, err
	}
	pkp, pwd, err := n.ClientCert.PrivateKeyPEMEnc()
	if err != nil {
		return nil, err
	}
	pkFile, err := encodeFileBytes("private-key.pem", pkp)
	if err != nil {
		return nil, err
	}
	s8021x := map[string]interface{}{
		"eap": []string{
			"tls",
		},
		"identity":                   n.AnonIdentity,
		"client-cert":                ccFile,
		"private-key":                pkFile,
		"private-key-password":       pwd,
		"private-key-password-flags": 0,
	}
	return installBase(n.Base, s8021x, pUUIDs)
}
