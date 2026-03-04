LINT_BIN := ./bin/golangci-lint

.PHONY: build-lint
build-lint:
    golangci-lint custom -v