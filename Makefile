all: todo
	@go vet
	@golint .

todo:
	@grep -rn TODO *.go || true
	@grep -rn println *.go || true

run:
	@go run fenster.go utils.go config.go

build: deps
	@export GOBIN=$(shell pwd)
	@go build

package: build
	@tar -cvzf fenster.tar.gz fenster config.ini data/

deps:
	@go get -d -v ./...

test: deps
	@go test ./...
