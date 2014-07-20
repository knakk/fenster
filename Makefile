all: build

build:
	export GOBIN=$(shell pwd)
	go build

package: build
	tar -cvzf fenster.tar.gz fenster config.ini data/

test:
	go get -u -v
	go test ./...
