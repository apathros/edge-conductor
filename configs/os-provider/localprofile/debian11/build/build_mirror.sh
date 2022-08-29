#!/bin/bash
#
# Copyright (c) 2022 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

set -x

# Output locatoin
# ${WEB_FILES}/${profile_name}/build/
# Example
# esp/data/usr/share/nginx/html/files/Ubuntu_20.04/build/
MIRROR_DIR=/opt/output
DEBIAN_VERSION=bullseye
DEBIAN_MIRROR_URL=http://ftp.us.debian.org/debian

# Output file
# debian_rootfs.tgz

# This script will build a debian roofs tarball debian_rootfs.img
# pre-request
#   minial debian os at least

# step 0 : prepare
apt update -y
apt install -y debootstrap

# step 1 : create rootfs image and mount on fs 2G
umount rootfs/dev || true
umount rootfs/proc || true
umount rootfs/sys || true
sync

# step 2 : mirror and make /tmpfile/debian as a rootfs
## warning !
## /etc/sudoer
## Defaults env_keep += "http_proxy https_proxy ftp_proxy"
mkdir -p rootfs
debootstrap --arch amd64 --variant=minbase $DEBIAN_VERSION rootfs $DEBIAN_MIRROR_URL

# step 3 : bind system fs
mount --bind /dev rootfs/dev
mount -t proc proc rootfs/proc
mount -t sysfs sysfs rootfs/sys

# step 4 : !! Can not be updated !! update package and install extra packages
#cp /etc/apt/apt.conf  /tmpfile/debian/etc/apt/apt.conf
#LANG=C.UTF-8 chroot /tmpfile/debian/ sh -c \
#        "apt update -y && \
#         apt install -y wget openssh-server && \
#         apt clean"

rm rootfs/etc/apt/sources.list
echo "deb $DEBIAN_MIRROR_URL $DEBIAN_VERSION main" >> rootfs/etc/apt/sources.list
echo "deb $DEBIAN_MIRROR_URL $DEBIAN_VERSION-updates main" >> rootfs/etc/apt/sources.list

# step 5: change root and install packages
LANG=C.UTF-8 chroot rootfs sh -c "
    export TERM=xterm-color && \
    export DEBIAN_FRONTEND=noninteractive && \
    export https_proxy=$https_proxy && \
    export no_proxy=$no_proxy && \
    export HTTPS_PROXY=$HTTPS_PROXY && \
    export NO_PROXY=$NO_PROXY && \
    apt update && \
    apt install -y sudo wget vim && \
    apt --download-only --assume-yes install systemd && \
    apt --download-only --assume-yes install locales && \
    apt --download-only --assume-yes install debconf && \
    apt --download-only --assume-yes install grub-efi&& \
    apt --download-only --assume-yes install shim-unsigned shim-helpers-amd64-signed && \
    apt --download-only --assume-yes install grub-pc && \
    apt --download-only --assume-yes install linux-image-generic && \
    apt --download-only --assume-yes install docker.io && \
    apt --download-only --assume-yes install pciutils && \
    apt --download-only --assume-yes install wget && \
    apt --download-only --assume-yes install openssh-server && \
    apt --download-only --assume-yes install socat  && \
    apt --download-only --assume-yes install ebtables  && \
    apt --download-only --assume-yes install ethtool  && \
    apt --download-only --assume-yes install conntrack  && \
    apt --download-only --assume-yes install ufw  && \
    apt --download-only --assume-yes install cloud-init  && \
    apt --download-only --assume-yes install pciutils  && \
    apt --download-only --assume-yes install net-tools  && \
    apt --download-only --assume-yes install nano  && \
    apt --download-only --assume-yes install init && \
    apt --download-only --assume-yes install vim && \
    apt --download-only --assume-yes install qemu-guest-agent && \
    apt --download-only --assume-yes install chrony && \
    apt --download-only --assume-yes install sudo && \
    apt --download-only --assume-yes install dkms && \
    apt --download-only --assume-yes install libgstreamer1.0-0 && \
    apt --download-only --assume-yes install gstreamer1.0-tools && \
    apt --download-only --assume-yes install gstreamer1.0-plugins-base && \
    apt --download-only --assume-yes install gstreamer1.0-plugins-good && \
    apt --download-only --assume-yes install gstreamer1.0-libav && \
    apt --download-only --assume-yes install mtools && \
    echo \"chroot install done !!!\""


# step 7 : unmount and put rootfs.tgz under tftp output path
umount rootfs/dev || true
umount rootfs/proc || true
umount rootfs/sys || true
sync

rm -f rootfs.tgz
pushd rootfs
tar czvf ../rootfs.tgz *
popd
mv rootfs.tgz  $MIRROR_DIR
