FROM golang:1.13.8-alpine3.11 AS build

RUN CGO_ENABLED=0 go build -o news ./cmd/main.go

# use alpine to build the image
FROM zenika/alpine-chrome:89-with-chromedriver

ARG user=chrome
ARG group=chrome
ARG uid=1000
ARG gid=1000
ARG USER_HOME=/home/coll
ARG alpine_mirror="dl-cdn.alpinelinux.org"

# to take advantage of the cache, don't put everything in one line

USER root

RUN sed -i "s/dl-cdn.alpinelinux.org/$alpine_mirror/g" /etc/apk/repositories
RUN cat /etc/apk/repositories
RUN apk update && apk add --no-cache bash \
        tzdata \
        font-noto \
        font-noto-cjk \
        font-noto-arabic \
        font-noto-thai \
        font-noto-devanagari

RUN apk add font-noto-emoji --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community

COPY fonts-local.conf /etc/fonts/local.conf
COPY --from=build news /usr/bin/news

# timezone support
RUN cp /usr/share/zoneinfo/Asia/Seoul /etc/localtime \
  && echo "Asia/Seoul" > /etc/timezone

RUN mkdir -p $USER_HOME \
  && chown ${uid}:${gid} $USER_HOME

VOLUME $USER_HOME
WORKDIR $USER_HOME

USER ${user}

EXPOSE 3000
ENTRYPOINT []

