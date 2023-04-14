package nm

import (
	"errors"
	"fmt"
	"os/user"
	"strings"

	"github.com/geteduroam/linux-app/internal/config"
	"github.com/geteduroam/linux-app/internal/network"
	"github.com/geteduroam/linux-app/internal/network/method"
	"github.com/geteduroam/linux-app/internal/nm/connection"
)

// encodePath encodes a string to a path expected by NetworkManager
// This path is prefixed with file:// and is expicitly NULL terminated
// It returns this path as a byte array
func encodePath(p string) []byte {
	// get the converted path
	// see: https://github.com/NetworkManager/NetworkManager/blob/main/examples/python/dbus/add-wifi-eap-connection.py#L12
	// \x00 is just NUL termination which NM expects
	c := fmt.Sprintf("file://%s\x00", p)
	return []byte(c)
}

// buildCertFile creates a certificate file to be used by NetworkManager
// It gets the name of the certificate that is used for the filename and ends it with .pem
// Cert is the array of certificates that need to be inputted between the BEGIN and END certificate strings
func buildCertFile(name string, cert []string) ([]byte, error) {
	filename := fmt.Sprintf("%s.pem", name)
	content := ""
	for i, c := range cert {
		if i != 0 {
			content += "\n"
		}
		content += fmt.Sprintf(
			`-----BEGIN CERTIFICATE-----
%s
-----END CERTIFICATE-----`,
			c)
	}
	p, err := config.WriteFile(filename, []byte(content))
	if err != nil {
		return nil, err
	}
	return encodePath(p), nil
}

// previousCon gets a connection object using the previous UUID
func previousCon(pUUID string) (*connection.Connection, error) {
	if pUUID == "" {
		return nil, errors.New("UUID is empty")
	}
	s, err := connection.NewSettings()
	if err != nil {
		return nil, err
	}
	return s.ConnectionByUUID(pUUID)
}

// createCon creates a new connection using the arguments
// if a previous connection was found with pUUID, it updates that one instead
// it returns the newly created or updated connection object
func createCon(pUUID string, args connection.SettingsArgs) (*connection.Connection, error) {
	prev, err := previousCon(pUUID)
	// previous connection found, update it with the new settings args
	if err == nil {
		return prev, prev.Update(args)
	}
	// create a connection settings object
	s, err := connection.NewSettings()
	if err != nil {
		return nil, err
	}
	// create a new connection
	return s.AddConnection(args)
}

// Install installs a non TLS network and returns an error if it cannot configure it
// Right now it adds a new profile that is not automatically added
// It returns the uuid if the connection was added successfully
func Install(n network.NonTLS, pUUID string) (string, error) {
	fID := fmt.Sprintf("%s (from Geteduroam)", n.SSID)
	cUser, err := user.Current()
	if err != nil {
		return "", err
	}
	cert, err := buildCertFile("ca-cert", n.Cert)
	if err != nil {
		return "", err
	}
	sCon := map[string]interface{}{
		"permissions": []string{
			fmt.Sprintf("user:%s", cUser.Username),
		},
		"type": "802-11-wireless",
		"id":   fID,
	}
	sWifi := map[string]interface{}{
		"ssid":     []byte(n.SSID),
		"security": "802-11-wireless-security",
	}
	sWsec := map[string]interface{}{
		"key-mgmt": "wpa-eap",
		"proto":    []string{"rsn"},
		"pairwise": []string{strings.ToLower(n.MinRSN)},
		"group":    []string{strings.ToLower(n.MinRSN)},
	}
	var sids []string

	for _, sid := range n.ServerIDs {
		v := fmt.Sprintf("DNS:%s", sid)
		sids = append(sids, v)
	}
	s8021x := map[string]interface{}{
		"eap": []string{
			n.Method().String(),
		},
		"identity":           n.Credentials.Username,
		"ca-cert":            cert,
		"anonymous-identity": n.AnonIdentity,
		"password":           n.Credentials.Password,
		"password-flags":     0,
		"altsubject-matches": sids,
	}
	if n.InnerAuth.EAP() && n.MethodType == method.TTLS {
		s8021x["phase2-autheap"] = n.InnerAuth.String()
	} else {
		s8021x["phase2-auth"] = n.InnerAuth.String()
	}
	sIP4 := map[string]interface{}{
		"method": "auto",
	}
	sIP6 := map[string]interface{}{
		"method": "auto",
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
