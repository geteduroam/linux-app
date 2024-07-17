# geteduroam Linux client

This repository contains the source code for the geteduroam Linux client. Currently WIP.

# Dependencies
- Go >= 1.18
- Make
- GTK >= 4.06 (for the GUI)
- Libadwaita >= 1.1 (for the GUI)
- libnotify (for notifications)

# Building
## CLI
```bash
make build-cli
```

## GUI
```bash
make build-gui
```

# Running
## CLI
```bash
make run-cli
```

## GUI
```bash
make run-gui
```

# Testing
```bash
make test
```

# Notifications
For eduroam profiles that use TLS client certificates, the client can
warn for imminent expiry. As the geteduroam client is not always open,
we provide Systemd user files that check daily for imminent
expiry. These systemd user files run the `cmd/geteduroam-notifcheck/` binary.

To _manually_ set this up with Systemd, make sure that the
`systemd/user` service and timer files are in a location that the
systemd user daemon can find it. E.g. move them to `/etc/systemd/user`
or `~/.config/systemd/user`. The DEB and RPM packages do this
automatically. Note that when moving these files, make sure to reload systemd with `systemctl --user daemon-reload`.

Contributions are welcome to support other daemons.
# License
BSD 3
