.PHONY: build lint fmt test clean

build:
	go build -o dist/dotfile ./cmd

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .

test:
	go test ./...

clean:
	rm -rf dist/
