all: build

build:
	export GOBIN=$(shell pwd)
	go build fenster.go config.go

package: build
	tar -cvzf fenster.tar.gz fenster config.ini data/

test:
	go get -u -v
	go test ./...

integration: build
	./integration-tests.sh
