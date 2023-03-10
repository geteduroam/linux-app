package nm

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/jwijenbergh/geteduroam-linux/internal/config"
	"github.com/jwijenbergh/geteduroam-linux/internal/network"
	"github.com/jwijenbergh/geteduroam-linux/internal/network/method"
	"github.com/jwijenbergh/geteduroam-linux/internal/nm/connection"
)

func encodePath(p string) []byte {
	// get the converted path
	// see: https://github.com/NetworkManager/NetworkManager/blob/main/examples/python/dbus/add-wifi-eap-connection.py#L12
	// \x00 is just NUL termination which NM expects
	c := fmt.Sprintf("file://%s\x00", p)
	return []byte(c)
}

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
	// TODO: or XDG_DATA_HOME?
	home := os.Getenv("XDG_CONFIG_HOME")
	if home == "" {
		// TODO: expand with $HOME instead?
		home = "~/.config"
	}
	c := &config.Config{Directory: path.Join(home, "geteduroam")}
	p, err := c.Write(filename, content)
	if err != nil {
		return nil, err
	}
	return encodePath(p), nil
}

func Install(n network.NonTLS) error {
	fID := fmt.Sprintf("%s (from Geteduroam)", n.SSID)
	s, err := connection.NewSettings()
	if err != nil {
		return nil
	}
	cUser, err := user.Current()
	if err != nil {
		return err
	}
	cert, err := buildCertFile("ca-cert", n.Cert)
	if err != nil {
		return err
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
		"identity":           n.Username,
		"ca-cert":            cert,
		"anonymous-identity": n.AnonIdentity,
		"password":           n.Password,
		"password-flags":     0,
		"altsubject-matches": sids,
	}
	if n.InnerAuth.EAP() && n.MethodType == method.TTLS {
		s8021x["phase2-autheap"] = n.InnerAuth.String()
	} else {
		s8021x["phase2-auth"] = n.InnerAuth.String()
	}
	sIp4 := map[string]interface{}{
		"method": "auto",
	}
	sIp6 := map[string]interface{}{
		"method": "auto",
	}
	settings := map[string]map[string]interface{}{
		"connection":               sCon,
		"802-11-wireless":          sWifi,
		"802-11-wireless-security": sWsec,
		"802-1x":                   s8021x,
		"ipv4":                     sIp4,
		"ipv6":                     sIp6,
	}
	_, err = s.AddConnection(settings)
	if err != nil {
		panic(err)
	}
	return nil
}
