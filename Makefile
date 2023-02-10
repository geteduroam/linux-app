build:
	go build cmd/geteduroam/main.go
	@echo "Done building, run 'make run' to run the client"

run: build
	./main

clean:
	go clean
	rm -rf main

all: build
