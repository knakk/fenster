all: build

build:
	@export GOBIN=$(shell pwd)
	@go build

package: build
	@tar -cvzf fenster.tar.gz fenster config.ini data/

deps:
	@go get -d -v ./...

test: deps
	@go test ./...
