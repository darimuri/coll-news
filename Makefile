SHELL := /bin/bash

test:
	go clean -testcache && CGO_ENABLED=0 TEST_HEADLESS=1 go test -p 1 -count 1 -timeout 30m ./...

run-cmd: build-cmd
	mkdir -p `pwd`/coll_dir
	./news coll -t mobile -s daum -d ./coll_dir -e -l 3 -b /usr/bin/chromium-browser

build-image:
	docker build -t coll-news:latest -f docker/Dockerfile ./

tag-image: build-image
ifdef TAG
	docker tag coll-news:latest darimuri/coll-news:${TAG}
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
	docker rmi coll-news:latest 2> /dev/null || exit 0

launch-image:
	mkdir -p `pwd`/coll_dir
	docker run -it --name coll-news -v `pwd`/coll_dir:/home/coll/coll_dir coll-news:latest bash

run-image: build-image
	mkdir -p `pwd`/coll_dir
	docker run -it --name coll-news -v `pwd`/coll_dir:/home/coll/coll_dir coll-news:latest news coll -t mobile -s daum -d ./coll_dir -e -l 3 -b /usr/bin/chromium-browser

