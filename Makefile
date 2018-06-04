plugin/rootfs create push enable: name := nickbreen/docker-lvm-plugin

.PHONY: create
create: plugin/rootfs
	docker plugin rm --force $(name) || true
	cp config.json plugin
	docker plugin create $(name) plugin

.PHONY: push
push: create
	docker plugin push $(name)

.PHONY: enable
enable: create
	docker plugin enable $(name)
	docker plugin set $(name) VOLUME_GROUP=skull
	docker plugin ls

.PHONY: clean
clean:
	rm -rf plugin

plugin/rootfs: vendor.conf main.go driver.go utils.go
	docker build --tag $(name):rootfs .
	docker rm --force --volumes rootfs || true
	docker create --name rootfs $(name):rootfs
	rm -rf $@
	mkdir -p $@
	docker export rootfs | tar -x -C $@
	docker rm --force --volumes rootfs

