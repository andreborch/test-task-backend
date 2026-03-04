LINT_BIN := ./loglint
CUSTOM_PLUGIN := ./custom-gcl

build:
 go build -o loglint ./cmd/loglint

build-lint: build
 golangci-lint custom -v

lint: build-lint
 $(CUSTOM_PLUGIN) run ./...

lint-only:
 $(CUSTOM_PLUGIN) run --enable-only loglint ./...

test:
 go test -v ./...

check: test lint

clean:
 rm -f $(LINT_BIN) loglint.exe loglint

.PHONY: clean check test lint-only lint build-lint