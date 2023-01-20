package main

import (
	"os"
	//"github.com/geteduroam/linux/internal/discovery"
	"github.com/geteduroam/linux/internal/ui"
)

func main() {
	gui := ui.New()
	os.Exit(gui.Run(os.Args))
}
