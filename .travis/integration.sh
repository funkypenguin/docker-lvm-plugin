#!/usr/bin/env bash

set -e -x -o pipefail

setup() {
    dd if=/dev/zero of=/tmp/loop0.img bs=$1 count=1M
    sudo losetup /dev/loop0 /tmp/loop0.img
    sudo sfdisk /dev/loop0 <<< ",,8e,,"
    sudo pvcreate /dev/loop0 -f
    sudo vgcreate test-vg /dev/loop0
}

teardown() {
    sudo lvremove test-vg -f || true
    sudo vgremove test-vg -f || true
    sudo losetup --detach /dev/loop0 || true
    rm -f /tmp/loop0.img
}

expectedVgs() {
    sudo vgs -S vg_name=test-vg
}

plugin() {
    # make and enable the plugin
    make create
    # configure the plugin
    sudo docker plugin set nickbreen/docker-lvm-plugin VOLUME_GROUP=test-vg
    sudo docker plugin enable nickbreen/docker-lvm-plugin
    # list enabled plugins
    sudo docker plugin ls --filter enabled=true | grep nickbreen/docker-lvm-plugin
}

trap "teardown" EXIT

setup 1024
plugin
