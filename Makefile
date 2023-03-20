.build:
	go build -o geteduroam-cli cmd/geteduroam/main.go

build: .build
	@echo "Done building, run 'make run' to run the client"

test:
	go test ./...

run: .build
	./geteduroam-cli

clean:
	go clean
	rm -rf geteduroam-cli
