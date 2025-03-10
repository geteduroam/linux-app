package main

import (
	"flag"
	"fmt"
	"time"

	"golang.org/x/exp/slog"

	"github.com/geteduroam/linux-app/internal/config"
	"github.com/geteduroam/linux-app/internal/log"
	"github.com/geteduroam/linux-app/internal/nm"
	"github.com/geteduroam/linux-app/internal/notification"
)

const usage = `Usage of %s:
  -h, --help			Prints this help information

  This CLI binary is needed for periodically checking for validity and giving notifications when the eduroam connection profile added by geteduroam is about to expire.
  It gives a warning 10 days before expiry, and then every day. You can schedule to start this binary daily yourself or rely on the built-in systemd user timer.
  You also need notify-send installed to send the actual notifications.

  Log file location: %s
`

func hasValidProfile(uuids []string) bool {
	for _, uuid := range uuids {
		con, err := nm.PreviousCon(uuid)
		if err != nil {
			slog.Error("no connection with uuid", "uuid", uuid, "error", err)
			continue
		}
		if con == nil {
			slog.Error("connection is nil")
			continue
		}
		return true
	}
	return false
}

func main() {
	program := "geteduroam-notifcheck"
	lpath, err := log.Location(program)
	if err != nil {
		lpath = "N/A"
	}
	flag.Usage = func() { fmt.Printf(usage, program, lpath) }
	flag.Parse()
	log.Initialize("geteduroam-notifcheck", false)
	cfg, err := config.Load()
	if err != nil {
		slog.Error("no previous state", "error", err)
		return
	}
	if cfg.Validity == nil {
		slog.Info("validity is nil")
		return
	}
	if !hasValidProfile(cfg.UUIDs) {
		slog.Info("no valid profiles found")
		return

	}

	valid := *cfg.Validity
	now := time.Now()
	diff := valid.Sub(now)
	days := int(diff.Hours() / 24)

	var text string
	if days > 10 {
		slog.Info("the profile is still valid for more than 10 days", "days", days)
		return
	}
	if days < 0 {
		text = "profile is expired"
	}
	if days == 0 {
		text = "profile expires today"
	}
	if days > 0 {
		text = fmt.Sprintf("profile expires in %d days", days)
	}
	msg := fmt.Sprintf("Your eduroam %s. Re-run geteduroam to renew the profile", text)
	err = notification.Send(msg)
	if err != nil {
		slog.Error("failed to send notification", "error", err)
		return
	}
}
