# geteduroam Linux client

This repository contains the source code for the geteduroam Linux client. Currently WIP.

[![Get it on Flathub](https://flathub.org/api/badge?locale=en)](https://flathub.org/apps/app.eduroam.geteduroam)

Note that currently it only works with NetworkManager. But support for e.g. wpa-supplicant and iwd is planned.

# Install through DEB/RPM
To install the client using official packages, go to the [GitHub
releases page](https://github.com/geteduroam/linux-app/releases) and
pick DEB or RPM. These files can be saved and double clicked in your
file manager to install.

# Manual install
This section, explains the steps needed to manually build the client. We go over the CLI client and GUI client.
We also have a small binary that is used for sending of notifications, which we will explain too.

## Dependencies
- Go >= 1.18
- Make
- GTK >= 4.06 (for the GUI)
- Libadwaita >= 1.1 (for the GUI)
- libnotify (for notifications)
- NetworkManager

## CLI
To build the CLI client run:
```bash
make build-cli
```

This outputs the CLI to `./geteduroam-cli`, move this to somewhere in your `$PATH`, e.g. `/usr/bin`.

During development, the CLI can be build and run with the command:
```bash
make run-cli
```

## GUI
To build the GUI client run:
```bash
make build-gui 
```

This outputs the GUI to `./geteduroam-gui`, move this to somewhere in your `$PATH`, e.g. `/usr/bin`.

During development, the GUI can be build and run with the command:
```bash
make run-gui
```

## Notifications
For eduroam profiles that use TLS client certificates, the client can
warn for imminent expiry. As the geteduroam client is not always open,
we provide Systemd user files that check daily for imminent
expiry. These systemd user files run the `./cmd/geteduroam-notifcheck/` binary.

To build this binary, run:
```bash
make build-notifcheck
```

This outputs this binary to `./geteduroam-notifcheck`, move this somewhere in your `$PATH`, e.g. `/usr/bin/`.

During development, the notifcheck binary build and run with the command:
```bash
make run-notifcheck
```

To then set this up with Systemd, make sure that the
`systemd/user` service and timer files are in a location that the
systemd user daemon can find it. E.g. move them to `/etc/systemd/user`
or `~/.config/systemd/user`. The DEB and RPM packages do this
automatically. Note that when moving these files, make sure to reload systemd with `systemctl --user daemon-reload`.
Also note that these files hard-code the path to your `geteduroam-notifcheck` binary, in the service files we assume it is in `/usr/bin/`.

Contributions are welcome to support other daemons.
# License
[BSD 3](./LICENSE)
