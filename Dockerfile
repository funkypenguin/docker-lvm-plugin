FROM golang:1.10-alpine AS builder

RUN apk add --no-cache gcc libc-dev
COPY . /go/src/github.com/nickbreen/docker-lvm-plugin
WORKDIR /go/src/github.com/nickbreen/docker-lvm-plugin
RUN go install --ldflags '-extldflags "-static"'

FROM alpine
RUN apk update && apk add lvm2 xfsprogs cryptsetup thin-provisioning-tools
RUN mkdir -p /run/docker/plugins
COPY --from=builder /go/bin/docker-lvm-plugin /
