GO ?= go

.PHONY: build run test fmt vet check

build:
	$(GO) build -o bin/parisi ./cmd/parisi

run:
	$(GO) run ./cmd/parisi

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
	$(GO) build -o bin/parisi ./cmd/parisi
