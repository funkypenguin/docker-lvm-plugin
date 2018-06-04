plugin/rootfs create push enable install: name := nickbreen/docker-lvm-plugin

.PHONY: create
create: config.json plugin/rootfs
	docker plugin rm --force $(name) || true
	cp config.json plugin
	docker plugin create $(name) plugin

.PHONY: install
install: push
	docker plugin rm --force $(name) || true
	docker plugin install $(name) --grant-all-permissions

.PHONY: push
push: create
	docker plugin push $(name)

.PHONY: enable
enable: create
	docker plugin ls | grep $(name)
	docker plugin set $(name) VOLUME_GROUP=skull
	docker plugin enable $(name)

.PHONY: clean
clean:
	rm -rf plugin

vendor: vendor.conf
	go get github.com/LK4D4/vndr
	vndr -whitelist github.com -whitelist golang.org
	ln -sf . vendor/src # Hack for Idea GoLand and Go Vendoring

plugin/rootfs: .dockerignore Dockerfile main.go driver.go utils.go vendor
	docker build --tag $(name):rootfs .
	docker rm --force --volumes rootfs || true
	docker create --name rootfs $(name):rootfs
	rm -rf $@
	mkdir -p $@
	docker export rootfs | tar -x -C $@
	docker rm --force --volumes rootfs

