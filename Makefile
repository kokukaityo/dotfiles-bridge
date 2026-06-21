.PHONY: build lint fmt test bats clean

build:
	go build -o dist/dotfile ./cmd

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .

test:
	go test ./...

bats: build
	npx bats test/bats/

clean:
	rm -rf dist/
