#!/bin/bash
#
# Copyright (c) 2022 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

set -a

#this is provided while using Utility OS
source /opt/bootstrap/functions



# --- Add Packages
ubuntu_bundles="openssh-server"
ubuntu_packages="wget socat ebtables ethtool conntrack ufw cloud-init pciutils net-tools nano vim qemu-guest-agent chrony mtools"

# --- List out any docker images you want pre-installed separated by spaces. ---
pull_sysdockerimagelist=""

# --- List out any docker tar images you want pre-installed separated by spaces.  We be pulled by wget. ---
wget_sysdockerimagelist="" 



# --- Install Extra Packages ---
run "Installing Extra Packages on Ubuntu ${param_ubuntuversion}" \
    "docker run -i --rm --privileged --name ubuntu-installer ${DOCKER_PROXY_ENV} -v /dev:/dev -v /sys/:/sys/ -v $ROOTFS:/target/root ubuntu:${param_ubuntuversion} sh -c \
    'mount --bind dev /target/root/dev && \
    mount -t proc proc /target/root/proc && \
    mount -t sysfs sysfs /target/root/sys && \
    LANG=C.UTF-8 chroot /target/root sh -c \
        \"$(echo ${INLINE_PROXY} | sed "s#'#\\\\\"#g") export TERM=xterm-color && \
        export DEBIAN_FRONTEND=noninteractive && \
        ${MOUNT_DURING_INSTALL} && \
	apt install -y ${ubuntu_bundles} && \
        apt install -y ${ubuntu_packages}\"'" \
    ${PROVISION_LOG}

run "DNS resolved-system Permanent" \
    "touch $ROOTFS/etc/systemd/resolved.conf && \
     sed -i \"s/#DNS=/DNS=$GATEWAY 127.0.0.53/g\" $ROOTFS/etc/systemd/resolved.conf && \
     ln -sf /run/systemd/resolve/resolv.conf $ROOTFS/etc/resolv.conf" \
    "$TMP/provisioning.log"

# --- Pull any and load any system images ---
#for image in $pull_sysdockerimagelist; do
#	run "Installing system-docker image $image" "docker exec -i system-docker docker pull $image" "$TMP/provisioning.log"
#done
#for image in $wget_sysdockerimagelist; do
#	run "Installing system-docker image $image" "wget -O- $image 2>> $TMP/provisioning.log | docker exec -i system-docker docker load" "$TMP/provisioning.log"
#done
