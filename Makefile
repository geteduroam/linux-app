.build-cli:
	go build -o geteduroam-cli ./cmd/geteduroam-cli

build-cli: .build-cli
	@echo "Done building, run 'make run-cli' to run the CLI"


test:
	go test ./...

run-cli: .build-cli
	./geteduroam-cli

clean:
	go clean
	rm -rf geteduroam-cli
