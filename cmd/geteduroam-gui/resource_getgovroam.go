//go:build getgovroam

package main

import "embed"

//go:embed resources/label.css resources/list.css resources/title.css resources/window_getgovroam.css resources/main.ui resources/gears.ui resources/images/success.png resources/images/heart.png
var resources embed.FS
