package main

import "embed"

//go:embed resources/*.css resources/geteduroam.ui
var resources embed.FS

func MustResource(name string) string {
	b, err := resources.ReadFile("resources/" + name)
	if err != nil {
		panic(err)
	}
	return string(b)
}

