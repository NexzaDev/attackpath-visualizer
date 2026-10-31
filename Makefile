.PHONY: format lint test build check demo

format:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

lint:
	go vet ./...

test:
	go test -race ./...

build:
	go build -o bin/attackpath-visualizer ./cmd/attackpath-visualizer

check: format lint test build

demo:
	go run ./cmd/attackpath-visualizer overview
