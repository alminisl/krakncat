# Makefile for krakncat

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build parameters
BINARY_NAME=krakn
BINARY_UNIX=$(BINARY_NAME)_unix

# Build targets
.PHONY: all build clean test deps tidy install

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v .

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

test:
	$(GOTEST) -v ./...

deps:
	$(GOMOD) download

tidy:
	$(GOMOD) tidy

# Cross compilation for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v .

# Install to system
install: build
	sudo cp $(BINARY_NAME) /usr/local/bin/

# Uninstall from system
uninstall:
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# Development commands
dev: deps tidy build

# Check for issues
check:
	$(GOCMD) vet ./...
	$(GOCMD) fmt ./...
