.build-cli:
	go build -o geteduroam-cli ./cmd/geteduroam-cli

.build-gui:
	CGO_ENABLED=0 go build -o geteduroam-gui ./cmd/geteduroam-gui

build-cli: .build-cli
	@echo "Done building, run 'make run-cli' to run the CLI"

build-gui: .build-gui
	@echo "Done building, run 'make run-gui' to run the GUI"

test:
	go test ./...

run-cli: .build-cli
	./geteduroam-cli

run-gui: .build-gui
	./geteduroam-gui

clean:
	go clean
	rm -rf geteduroam-cli
	rm -rf geteduroam-gui
