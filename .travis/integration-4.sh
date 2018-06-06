#!/usr/bin/env bash

. .travis/integration.sh

# 4.  Create a LUKS encrypted lvm volume named `crypt_vol` with the contents
#     of `/root/key.bin` as a binary passphrase. Snapshots of encrypted volumes
#     use the same key file. The key file must be present when the volume is
#     created, and when it is mounted to a container.

D=$(sudo find /var/lib/docker/plugins -type d -name docker-lvm-plugin)
sudo dd if=/dev/urandom of=$D/key.bin bs=512 count=4
sudo docker volume create --driver nickbreen/docker-lvm-plugin \
    --opt size=192M \
    --opt keyfile=/var/lib/docker-lvm-plugin/key.bin \
    --name test-crypt

expected_manifest test-crypt
expected_lvs test-crypt
expected_vgs

sudo docker volume rm test-crypt
