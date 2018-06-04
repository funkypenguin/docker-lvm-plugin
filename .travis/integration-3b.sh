#!/usr/bin/env bash

. .travis/integration.sh

# 3.  Create a snapshot volume. For thin snapshots don't specify a size.

sudo docker volume create --driver nickbreen/docker-lvm-plugin \
    --opt size=128M \
    --name test-lv

sudo docker volume create --driver nickbreen/docker-lvm-plugin \
    --opt snapshot=test-lv \
    --opt size=192M \
    --name test-thin-snapshot

sudo lvs --no-headings --options lv_name | grep test-snapshot
expectedVgs

sudo docker volume rm test-lv test-thin-snapshot