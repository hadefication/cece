BINARY=cc
VERSION?=dev
LDFLAGS=-ldflags "-X github.com/hadefication/cece/cmd.version=$(VERSION)"

.PHONY: build install test clean

build:
	go build $(LDFLAGS) -o $(BINARY) .

install: build
	cp $(BINARY) ~/.local/bin/$(BINARY)

test:
	go test ./... -v

clean:
	rm -f $(BINARY)
