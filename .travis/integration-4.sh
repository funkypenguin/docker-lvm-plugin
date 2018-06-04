#!/usr/bin/env bash

. .travis/integration.sh

# 4.  Create a LUKS encrypted lvm volume named `crypt_vol` with the contents
#     of `/root/key.bin` as a binary passphrase. Snapshots of encrypted volumes
#     use the same key file. The key file must be present when the volume is
#     created, and when it is mounted to a container.

dd if=/dev/urandom of=/tmp/key.bin bs=1024 count=4
sudo docker volume create --driver nickbreen/docker-lvm-plugin \
    --opt size=192M \
    --opt keyfile=/tmp/key.bin \
    --name test-crypt

sudo docker volume rm test-crypt
