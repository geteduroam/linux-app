package main

func MustResource(name string) string {
	b, err := resources.ReadFile("resources/" + name)
	if err != nil {
		panic(err)
	}
	return string(b)
}
