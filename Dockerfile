FROM golang:1.10-alpine AS builder

COPY . /go/src/github.com/nickbreen/docker-lvm-plugin
WORKDIR /go/src/github.com/nickbreen/docker-lvm-plugin
RUN set -ex \
    && apk add --no-cache --virtual .build-deps gcc libc-dev \
    && go install --ldflags '-extldflags "-static"' \
    && apk del .build-deps
CMD [ "/go/bin/docker-lvm-plugin" ]

FROM alpine
RUN apk update && apk add lvm2
RUN mkdir -p /run/docker/plugins
COPY --from=builder /go/bin/docker-lvm-plugin .
CMD ["docker-lvm-plugin"]
