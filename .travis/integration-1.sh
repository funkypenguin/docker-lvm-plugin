#!/usr/bin/env bash

. .travis/integration.sh
set -e -o pipefail

# 1.  Create an LVM volume

sudo docker volume create --driver nickbreen/docker-lvm-plugin \
    --opt size=32M \
    --name test-lv

expected_manifest test-lv
expected_lvs test-lv
expected_vgs

sudo docker volume rm test-lv