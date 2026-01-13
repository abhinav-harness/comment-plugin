.PHONY: build test clean docker

BINARY=comment-plugin
IMAGE=plugins/comment

build:
	CGO_ENABLED=0 go build -o $(BINARY) ./cmd/plugin

test:
	go test -v ./...

clean:
	rm -f $(BINARY)

docker:
	docker build -t $(IMAGE) .
