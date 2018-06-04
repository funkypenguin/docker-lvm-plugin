# LVM Volume Plugin 

[![Build Status](https://travis-ci.org/nickbreen/docker-lvm-plugin.svg?branch=master)](https://travis-ci.org/nickbreen/docker-lvm-plugin)

Docker Volume Driver plugin for LVM volumes

This plugin can be used to create lvm volumes of specified size, which can 
then be bind mounted into the container using `docker run` command.

## Source

Source: [nickbreen/docker-lvm-plugin](https://github.com/nickbreen/docker-lvm-plugin), 
forked from [projectatomic/docker-lvm-plugin](https://github.com/projectatomic/docker-lvm-plugin)'s 
standalone plugin container.

## Building From Source

Consult the `Makefile`. The `Dockerfile` provides a golang build container and also 
generates the plugin rootfs.

Build, create, and enable the plugin locally. The plugin will be automatically installed to 
your local docker daemon and enabled.

```bash
make enable
```

You will need to set the volume group (see [Installation](README.md#Installation))

## Installation

1.  Install the plugin, you should specify a volume group to use during installation.
    
    ```bash
    docker plugin install \
        --alias lvm --grant-all-permissions \
        nickbreen/docker-lvm-plugin VOLUME_GROUP=docker-vg
    ```

2.  You may set/change the volume group later if necessary.
     
    ```bash
    docker plugin set lvm VOLUME_GROUP=docker-vg
    ```

## Usage

1.  It is the responsibility of the administrator to provide a volume group.
    You can choose an existing volume group or create a new volume group 
    using the `vgcreate` command; e.g.
     
    ```bash
    pvcreate /dev/disk/by-uuid/80df0f82-67bc-11e8-b645-3bfd0060be4c
    vgcreate docker-vg /dev/disk/by-uuid/80df0f82-67bc-11e8-b645-3bfd0060be4c
    ```
    
2.  Configure the plugin to use this volume group.  
    ```bash
    docker plugin set lvm VOLUME_GROUP=docker-vg
    ```

3.  The docker-lvm-plugin also supports the creation of thinly-provisioned volumes. 
    To create a thinly-provisioned volume the administrator must first create a thin 
    pool using the `lvcreate` command.
    ```bash
    lvcreate --size 10G --thin docker-vg/docker-thinpool
    ```
    This will create a thinpool named `docker-thinpool` of size 10G under volume group 
    `docker-vg`. *Note:* thinpools are special kind of logical volumes carved out of 
    the volume group. To create this thin pool the volume group must have at least 10G 
    of free space. 

## Volume Creation

The `docker volume create` command supports the creation of regular lvm volumes, thin 
volumes, snapshots of regular and thin volumes.

Usage: `docker volume create [OPTIONS]`

```
-d, --driver    string    Specify volume driver name (default "local")
--label         list      Set metadata for a volume (default [])
--name          string    Specify volume name
-o, --opt       map       Set driver specific options (default map[]) 
```

Following options can be passed using `-o` or `--opt`

```
--opt size
--opt thinpool
--opt snapshot
--opt keyfile
```
 
### Examples

1.  Create a lvm volume named `foobar` of size 208 MB (0.2 GB).
    
    ```bash
    docker volume create --driver lvm \
        --opt size=0.2G \
        --name foobar
    ```

2.  Create a thinly-provisioned lvm volume named `thin_vol` in `docker-thinpool`.

    ```bash
    docker volume create --driver lvm \
        --opt size=0.2G \
        --opt thinpool=docker-thinpool \
        --name thin_vol
    ```
    
3.  Create a snapshot volume of `foobar` named `foobar_snapshot`. For thin 
    snapshots, use the same command above but don't specify a size.
    
    ```bash
    docker volume create --driver lvm \
        --opt snapshot=foobar \
        --opt size=100M \
        --name foobar_snapshot
    ```
    
4.  Create a LUKS encrypted lvm volume named `crypt_vol` with the contents 
    of `/root/key.bin` as a binary passphrase. Snapshots of encrypted volumes 
    use the same key file. The key file must be present when the volume is 
    created, and when it is mounted to a container.
    ```bash
    docker volume create --driver lvm \
        --opt size=0.2G \
        --opt keyfile=/root/key.bin \
        --name crypt_vol
    ```
