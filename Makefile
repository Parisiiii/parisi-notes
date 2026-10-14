GO ?= go
INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: build install run test fmt vet check

build:
	$(GO) build -o bin/parisi-notes ./cmd/parisi-notes

install:
	mkdir -p $(INSTALL_DIR)
	GOBIN=$(INSTALL_DIR) $(GO) install ./cmd/parisi-notes

run:
	$(GO) run ./cmd/parisi-notes

test:
	$(GO) test ./...

fmt:
	gofmt -w cmd internal

vet:
	$(GO) vet ./...

check:
	@test -z "$$(gofmt -l cmd internal)" || (echo 'Run make fmt'; exit 1)
	$(GO) vet ./...
	$(GO) test ./...
	$(GO) build -o bin/parisi-notes ./cmd/parisi-notes
