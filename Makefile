.PHONY: build lint fmt test bats clean

GOEXE := $(shell go env GOEXE)

build:
	go build -o dist/dotfiles$(GOEXE) ./cmd

lint:
	golangci-lint run ./...

fmt:
	gofumpt -w .

test:
	GOTMPDIR=$(CURDIR)/dist go test ./...

bats: build
	npx bats test/bats/

clean:
	rm -rf dist/

ifneq (,$(filter exe-%,$(MAKECMDGOALS)))
  EXE_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  $(eval $(EXE_ARGS):;@:)
endif

exe-%:
	DOTFILES_DIR=$(or $(DOTFILES_DIR),$$HOME/dotfiles-test) GOTMPDIR=$(CURDIR)/dist go run ./cmd $* $(EXE_ARGS)
