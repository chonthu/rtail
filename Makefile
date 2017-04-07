COMMIT = $(shell git describe --tags --dirty --always)

default: test

test:
	go test -v

build:
	go build -ldflags "-X main.Version=$(COMMIT)"

package: build
	@sh -c "'$(CURDIR)/bin/package.sh'"