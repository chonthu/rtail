COMMIT = $(shell git describe --tags --dirty --always)

default: test

test:
	go test -v

build:
	go build -ldflags "-X main.Version=$(COMMIT)"

zip:
	gzip rtail

package: test build zip
	go run release/release.go $(COMMIT) rtail.gz > rtail.rb
	