all: todo
	@go vet
	@golint .

todo:
	@grep -rn TODO *.go || true
	@grep -rn println *.go || true

run:
	@go run fenster.go utils.go config.go metrics.go tee.go

build: deps
	@go build

package: build
	@tar -cvzf fenster.tar.gz fenster config.ini data/

deps:
	@go get -d -v ./...

test: deps
	@go test ./...
