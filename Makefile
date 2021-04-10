SHELL := /bin/bash

test:
	go clean -testcache && CGO_ENABLED=0 TEST_HEADLESS=1 go test -p 1 -count 1 -timeout 30m ./...

build-cmd:
	CGO_ENABLED=0 go build -o news ./cmd/main.go