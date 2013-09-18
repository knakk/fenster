all: build

build:
	export GOBIN=$(shell pwd)
	go build fenster.go config.go

package: build
	tar -cvzf fenster.tar.gz fenster config.ini data/

test:
	go test *.go -v
	go test rdf/*.go -v
	go test sparql/*.go -v