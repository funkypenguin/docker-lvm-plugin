# Installation Directories
SYSCONFDIR ?=$(DESTDIR)/etc/docker
SYSTEMDIR ?=$(DESTDIR)/usr/lib/systemd/system
GOLANG ?= /usr/bin/go
BINARY ?= docker-lvm-plugin
MANINSTALLDIR?= ${DESTDIR}/usr/share/man
BINDIR ?=$(DESTDIR)/usr/libexec/docker

export GO15VENDOREXPERIMENT=1

all: man lvm-plugin-build

.PHONY: man
man:
	go-md2man -in man/docker-lvm-plugin.8.md -out docker-lvm-plugin.8

.PHONY: lvm-plugin-build
lvm-plugin-build: main.go driver.go
	$(GOLANG) build -o $(BINARY) .

.PHONY: install
install:
	if test ! -f "$(SYSCONFDIR)/docker-lvm-plugin"; then install -D -m 644 -t $(SYSCONFDIR) etc/docker/docker-lvm-plugin; fi
	install -D -m 644 -t $(SYSTEMDIR) systemd/docker-lvm-plugin.service systemd/docker-lvm-plugin.socket
	install -D -m 755 -t $(BINDIR) $(BINARY) 
	install -D -m 644 -t ${MANINSTALLDIR}/man8 docker-lvm-plugin.8 

.PHONY: clean
clean:
	rm -f $(BINARY)
	rm -f docker-lvm-plugin.8
	rm -rf plugin/rootfs

.PHONY: docker-plugin
docker-plugin: man lvm-plugin-build
	install -D -m 755 -t plugin/rootfs/$(BINDIR) $(BINARY)
	docker plugin rm lvm || true
	docker plugin create lvm plugin
