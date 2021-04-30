SHELL := /bin/bash

test:
	go clean -testcache && CGO_ENABLED=0 TEST_HEADLESS=1 go test -p 1 -count 1 -timeout 30m ./...

build-cmd:
	CGO_ENABLED=0 go build -o news ./cmd/main.go

image-build: build-cmd
	cp -a news docker/
	docker build -t coll-news:latest docker

image-rm:
	docker rm coll-news
	docker rmi coll-news:latest

image-test:
	mkdir -p /home/dormael/coll-news
	docker run -it --name coll-news -v /home/dormael/coll-news:/home/coll coll-news:latest bash

image-run:
	mkdir -p /home/dormael/coll-news
	docker run -it --name coll-news -v /home/dormael/coll-news:/home/coll coll-news:latest news coll -t mobile -n daum -s ./coll_dir