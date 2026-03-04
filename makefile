LINT_BIN := ./bin/golangci-lint

.PHONY: build-lint
build-lint:
    golangci-lint custom -v

.PHONY: lint
lint: build-lint
    $(LINT_BIN) run ./...

.PHONY: lint-only
lint-only:
    $(LINT_BIN) run --enable-only loglint ./...

.PHONY: test
test:
    go test -v ./...

.PHONY: check
check: test lint

.PHONY: clean
clean:
    rm -f $(LINT_BIN) loglint.exe loglint