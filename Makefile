# Minimal build helpers for lightweight delivery.

BIN ?= bin/znp
GOFILES ?= ./cmd/znp
GOFLAGS ?=
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS ?= -s -w \
	-X github.com/zero-net-panel/zero-net-panel/cmd/znp/cli.Version=$(VERSION) \
	-X github.com/zero-net-panel/zero-net-panel/cmd/znp/cli.Commit=$(COMMIT) \
	-X github.com/zero-net-panel/zero-net-panel/cmd/znp/cli.BuildDate=$(DATE)

.PHONY: build
build:
	@echo "Building $(BIN) (version $(VERSION))"
	@mkdir -p $(dir $(BIN))
	GOFLAGS="$(GOFLAGS)" go build -ldflags "$(LDFLAGS)" -o $(BIN) $(GOFILES)

.PHONY: clean
clean:
	@rm -rf $(BIN)
