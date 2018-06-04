#!/usr/bin/env bash

set -e -x -o pipefail

. .travis/integration.sh


# 1.  Create an LVM volume

sudo docker volume create --driver nickbreen/docker-lvm-plugin \
    --opt size=128M \
    --name test-lv

sudo lvs --no-headings --options lv_name | grep test-lv
expectedVgs

sudo docker volume rm test-lv