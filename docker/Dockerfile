# use alpine to build the image
FROM alpine:latest

ARG user=coll
ARG group=coll
ARG uid=1000
ARG gid=1000
ARG USER_HOME=/home/coll
ARG alpine_mirror="dl-cdn.alpinelinux.org"

# to take advantage of the cache, don't put everything in one line

RUN sed -i "s/dl-cdn.alpinelinux.org/$alpine_mirror/g" /etc/apk/repositories
RUN cat /etc/apk/repositories
RUN apk update && apk add --no-cache bash \
        alsa-lib \
        at-spi2-atk \
        atk \
        cairo \
        cups-libs \
        dbus-libs \
        eudev-libs \
        expat \
        flac \
        gdk-pixbuf \
        glib \
        libgcc \
        libjpeg-turbo \
        libpng \
        libwebp \
        libx11 \
        libxcomposite \
        libxdamage \
        libxext \
        libxfixes \
        tzdata \
        libexif \
        udev \
        xvfb \
        zlib-dev \
        font-noto \
        font-noto-cjk \
        font-noto-arabic \
        font-noto-thai \
        font-noto-devanagari \
        chromium \
        chromium-chromedriver

RUN apk add font-noto-emoji --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community

COPY fonts-local.conf /etc/fonts/local.conf
COPY news /usr/bin/news

# timezone support
RUN apk add tzdata

RUN mkdir -p $USER_HOME \
  && chown ${uid}:${gid} $USER_HOME \
  && addgroup -g ${gid} ${group} \
  && adduser -h "$USER_HOME" -u ${uid} -G ${group} -s /bin/bash -D ${user}

VOLUME $USER_HOME
WORKDIR $USER_HOME

USER ${user}
