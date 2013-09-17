all: build

build:
	export GOBIN=$(shell pwd)
	go build fenster.go

test:
	go test *.go -v