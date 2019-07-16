# Docker LVM Plugin [![Build Status](https://travis-ci.org/nickbreen/docker-lvm-plugin.svg?branch=master)](https://travis-ci.org/nickbreen/docker-lvm-plugin)

Docker Volume Driver for LVM volumes packaged as a Docker Plugin: see https://github.com/containers/docker-lvm-plugin.

This plugin can be used to create lvm volumes of specified size, which can
then be bind mounted into the container using `docker run` command.

## Setup

Edit `/etc/docker/docker-lvm-plugin` and set the volume group to create volumes in.

Install (and enable) the plugin

    docker plugin install nickbreen/docker-lvm-plugin

The docker-lvm-plugin also supports the creation of thinly-provisioned volumes. To create a thinly-provisioned volume, a user (administrator) must first create a thin pool using the `lvcreate` command.
```bash
lvcreate -L 10G -T vg1/mythinpool
```
This will create a thinpool named `mythinpool` of size 10G under volume group `vg1`.
NOTE: thinpools are special kind of logical volumes carved out of the volume group.
Hence in the above example, to create the thinpool `mythinpool` you must have atleast 10G of freespace in volume group `vg1`.

## Volume Creation
`docker volume create` command supports the creation of regular lvm volumes, thin volumes, snapshots of regular and thin volumes.

Usage: docker volume create [OPTIONS]
```bash
-d, --driver    string    Specify volume driver name (default "local")
--label         list      Set metadata for a volume (default [])
--name          string    Specify volume name
-o, --opt       map       Set driver specific options (default map[])
```
Following options can be passed using `-o` or `--opt`
```bash
--opt size
--opt thinpool
--opt snapshot
--opt keyfile
```
Please see examples below on how to use these options.

## Examples
```bash
$ docker volume create -d lvm --opt size=0.2G --name foobar
```
This will create a lvm volume named `foobar` of size 208 MB (0.2 GB).
```bash
docker volume create -d lvm --opt size=0.2G --opt thinpool=mythinpool --name thin_vol
```
This will create a thinly-provisioned lvm volume named `thin_vol` in mythinpool.
```bash
docker volume create -d lvm --opt snapshot=foobar --opt size=100M --name foobar_snapshot
```
This will create a snapshot volume of `foobar` named `foobar_snapshot`. For thin snapshots, use the same command above but don't specify a size.
```bash
docker volume create -d lvm --opt size=0.2G --opt keyfile=/root/key.bin --name crypt_vol
```

This will create a LUKS encrypted lvm volume named `crypt_vol` with the contents of `/root/key.bin` as a binary passphrase. Snapshots of encrypted volumes use the same key file. The key file must be present when the volume is created, and when it is mounted to a container.
