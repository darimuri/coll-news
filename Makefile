SHELL := /bin/bash

COMMIT=$(shell git log -1 --pretty=%h)
DATE=$(shell date +'%Y-%m-%d')
VERSION="v0.1"

test:
	go clean -testcache && CGO_ENABLED=0 TEST_HEADLESS=1 go test -p 1 -count 1 -timeout 30m ./...

build-cmd:
	CGO_ENABLED=0 go build -ldflags "-X 'github.com/darimuri/coll-news/cmd/version.commit=$(COMMIT)' -X 'github.com/darimuri/coll-news/cmd/version.date=$(DATE)' -X 'github.com/darimuri/coll-news/cmd/version.version=$(VERSION)'" -o news ./cmd/main.go

run-cmd: build-cmd
	mkdir -p `pwd`/coll_dir
	./news coll -t mobile -s daum -d ./coll_dir -e -l 3 -b /usr/bin/chromium-browser --stop-after-collect

build-image:
	docker rmi coll-news:$(COMMIT) 2> /dev/null || exit 0
	docker build --build-arg COMMIT=$(COMMIT) --build-arg VERSION=$(VERSION) --build-arg DATE=$(DATE) -t coll-news:$(COMMIT) -f docker/Dockerfile ./

tag-image: build-image
ifdef TAG
	docker tag coll-news:$(COMMIT) darimuri/coll-news:${TAG}
else
	@echo "TAG is required"
endif

push-image: tag-image
ifdef TAG
	docker push darimuri/coll-news:${TAG}
else
	@echo "TAG is required"
endif

rm-image:
	docker rm coll-news 2> /dev/null || exit 0
	docker rmi coll-news:$(COMMIT) 2> /dev/null || exit 0

launch-image: build-image
	mkdir -p `pwd`/coll_dir
	docker rm coll-news-$(COMMIT) 2> /dev/null || exit 0
	docker run -it --name coll-news-$(COMMIT) -v `pwd`/coll_dir:/home/coll/coll_dir coll-news:$(COMMIT) bash

run-image-mobile: build-image
	mkdir -p `pwd`/coll_dir
	docker run -it --name coll-news-mobile-$(COMMIT) -v `pwd`/coll_dir:/home/coll/coll_dir coll-news:$(COMMIT) news coll -t mobile -s daum -d ./coll_dir -e -l 3 -b /usr/bin/chromium-browser

run-image-pc: build-image
	mkdir -p `pwd`/coll_dir
	docker run -it --name coll-news-pc-$(COMMIT) -v `pwd`/coll_dir:/home/coll/coll_dir coll-news:$(COMMIT) news coll -t pc -s daum -d ./coll_dir -e -l 3 -b /usr/bin/chromium-browser

