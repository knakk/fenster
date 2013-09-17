all: build

build:
	export GOBIN=$(shell pwd)
	go build fenster.go

test:
	go test *.go -v
	go test rdf/*.go -v
	go test sparql/*.go -v