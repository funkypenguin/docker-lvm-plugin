#!/usr/bin/env bash

set -e -x -o pipefail

. .travis/integration.sh


# 1.  Create an LVM volume

sudo docker volume create --driver nickbreen/docker-lvm-plugin \
    --opt size=128M \
    --name test-lv

expected_lvs test-lv
expected_vgs

sudo docker volume rm test-lv