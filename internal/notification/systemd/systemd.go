package systemd

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/exp/slog"
)

func hasSystemd() bool {
	if _, err := os.Stat("/run/systemd/system"); !os.IsNotExist(err) {
		return true
	}
	return false
}

func hasUnit(unit string) bool {
	_, err := exec.Command("systemctl", "--user", "list-unit-files", unit).Output()
	return err == nil
}

const timerName string = "geteduroam-notifs.timer"

func hasUnitFiles() bool {
	if !hasUnit("geteduroam-notifs.service") {
		slog.Error("geteduroam-notifs.service is not installed anywhere")
		return false
	}
	if !hasUnit(timerName) {
		slog.Error(fmt.Sprintf("%s is not installed anywhere", timerName))
		return false
	}
	return true
}

// HasDaemonSupport returns whether or not notifications can be enabled globally
// This depends on if systemd is used and if the unit is ready to be enabled
func HasDaemonSupport() bool {
	if !hasSystemd() {
		return false
	}
	if !hasUnitFiles() {
		return false
	}
	return true
}

// EnableDaemon enables the notification daemon using systemctl commands
func EnableDaemon() error {
	_, err := exec.Command("systemctl", "--user", "enable", "--now", timerName).Output()
	return err
}

// DisableDaemon disables the notification daemon using systemctl commands
func DisableDaemon() error {
	_, err := exec.Command("systemctl", "--user", "is-enabled", timerName).Output()
	// when the timer is not enabled, return nil error and log
	if err != nil {
		slog.Debug("systemd reports timer is not enabled", "err", err)
		return nil
	}
	// timer is enabled
	// disable it
	_, err = exec.Command("systemctl", "--user", "disable", "--now", timerName).Output()
	return err
}
