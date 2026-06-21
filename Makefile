.PHONY: build lint fmt test bats clean

GOEXE := $(shell go env GOEXE)

build:
	go build -o dist/dotfile$(GOEXE) ./cmd

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
