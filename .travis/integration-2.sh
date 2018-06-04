#!/usr/bin/env bash

. .travis/integration.sh

# 2.  Create a thinly-provisioned LVM volume named

sudo lvcreate --size 256M --thin test-vg/test-thinpool

sudo docker volume create --driver nickbreen/docker-lvm-plugin \
    --opt size=128M \
    --opt thinpool=test-thinpool \
    --name test-thin-lv

expected_lvs test-thin-lv
expected_vgs

sudo docker volume rm test-thin-lv
