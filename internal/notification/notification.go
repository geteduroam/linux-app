package notification

import (
	"os/exec"

	"golang.org/x/exp/slog"

	"github.com/geteduroam/linux-app/internal/notification/systemd"
	"github.com/geteduroam/linux-app/internal/variant"
)

// Send sends a single notification with notify-send
func Send(msg string) error {
	_, err := exec.Command("notify-send", variant.DisplayName, msg).Output()
	return err
}

// HasDaemonSupport returns whether or not notifications can be enabled globally
func HasDaemonSupport() bool {
	// currently we only support systemd
	return systemd.HasDaemonSupport()
}

// enableDaemon enables the notification using systemd's user daemon
func enableDaemon() error {
	// currently we only support systemd
	return systemd.EnableDaemon()
}

// disableDaemon disables notifications when they were enabled
func disableDaemon() error {
	return systemd.DisableDaemon()
}

// ConfigureDaemon configures the notification daemon
// on if enable is true
// else off
// it logs if an error occurs
func ConfigureDaemon(enable bool) {
	var err error
	if enable {
		err = enableDaemon()
	} else {
		err = disableDaemon()
	}
	if err != nil {
		slog.Error("failed to disable/enable notification support", "state", enable, "err", err)
	}
}
