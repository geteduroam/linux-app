# Do not show gtk warnings for CGO building
export CGO_CPPFLAGS="-Wno-deprecated-declarations"

build:
	@echo "Building... This can take a while for the first run"
	go build cmd/geteduroam/main.go
	@echo "Done building, run make run to run the client"

run: build
	./main

clean:
	go clean
	rm -rf main

all: build
