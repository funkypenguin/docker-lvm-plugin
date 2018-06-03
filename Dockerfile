FROM golang

RUN apt-get update -q \
    && DEBIAN_FRONTEND=noninteractive apt-get install go-md2man lvm2 -yyq \
    && apt-get clean -q
ARG SRC=/go/src/docker-lvm-plugin
COPY Godeps/_workspace/src /go/src
COPY . ${SRC}/
WORKDIR ${SRC}
RUN make GOLANG=go && make install \
    && /usr/libexec/docker/docker-lvm-plugin -version
CMD [ "/usr/libexec/docker/docker-lvm-plugin" ]
