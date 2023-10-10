package main

import "embed"

//go:embed resources/*.css resources/geteduroam.ui resources/gears.ui resources/images/success.png resources/images/geteduroam.png
var resources embed.FS

func MustResource(name string) string {
	b, err := resources.ReadFile("resources/" + name)
	if err != nil {
		panic(err)
	}
	return string(b)
}
