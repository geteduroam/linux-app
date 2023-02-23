.build:
	go build cmd/geteduroam/main.go

build: .build
	@echo "Done building, run 'make run' to run the client"

test:
	go test ./...

run: .build
	./main

clean:
	go clean
	rm -rf main
