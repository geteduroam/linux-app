package main

import (
	//"github.com/geteduroam/linux/internal/discovery"
	"github.com/geteduroam/linux/internal/ui"

    "os"
)

func main() {
	gui := ui.New()
	os.Exit(gui.Run(os.Args))
}
