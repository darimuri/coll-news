SHELL := /bin/bash

test:
	go clean -testcache && CGO_ENABLED=0 TEST_HEADLESS=1 go test -p 1 -count 1 -timeout 30m ./...

build-cmd:
	CGO_ENABLED=0 go build -o news ./cmd/main.go

run-cmd:
	CGO_ENABLED=0 go run ./cmd/main.go coll -t mobile -s daum -d ./coll_dir -e -l 10

build-image: build-cmd
	cp -a news docker/
	docker build -t coll-news:latest docker

tag-image: build-image
ifdef TAG
	docker tag coll-news:latest dormael/coll-news:${TAG}
else
	@echo "TAG is required"
endif

push-image: tag-image
ifdef TAG
	docker push dormael/coll-news:${TAG}
else
	@echo "TAG is required"
endif

rm-image:
	docker rm coll-news
	docker rmi coll-news:latest

launch-image:
	mkdir -p /home/dormael/coll-news
	docker run -it --name coll-news -v /home/dormael/coll-news:/home/coll coll-news:latest bash

run-image: build-image
	mkdir -p /home/dormael/coll-news
	docker run -it --name coll-news -v /home/dormael/coll-news:/home/coll coll-news:latest news coll -t mobile -s daum -d ./coll_dir -e -l 3

