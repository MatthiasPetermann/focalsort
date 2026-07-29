APP := focalsort
DIST := dist
GOARCH ?= amd64

.PHONY: all build linux test clean

all: build

build:
	go build -o $(APP) .

# Creates a statically linked Linux binary that can be copied to a server.
linux:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -trimpath -ldflags="-s -w" -o $(DIST)/$(APP)-linux-$(GOARCH) .

test:
	go test ./...

clean:
	rm -rf $(DIST)
	rm -f $(APP)
