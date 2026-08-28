.PHONY: build release test clean install

# Binary output names
BINARY_NAME=hop
VERSION=1.1.0
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")

# High-performance release flags (stripped DWARF and debug symbols)
LDFLAGS=-s -w -X 'hoppr/internal/version.Version=$(VERSION)' -X 'hoppr/internal/version.Commit=$(COMMIT)' -X 'hoppr/internal/version.BuildDate=$(BUILD_DATE)'

build:
	go build -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY_NAME) .

release:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -o $(BINARY_NAME) ./cmd/hop

test:
	go test -v -race ./...

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

install:
	go install -ldflags="$(LDFLAGS)" -trimpath ./cmd/hop
