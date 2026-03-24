BINARY=ytc
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X github.com/dustin/youtrack-cli/cmd.Version=$(VERSION)"

.PHONY: build install clean test

build:
	go build $(LDFLAGS) -o $(BINARY) .

install:
	go install $(LDFLAGS) .

clean:
	rm -f $(BINARY)

test:
	go test ./...
