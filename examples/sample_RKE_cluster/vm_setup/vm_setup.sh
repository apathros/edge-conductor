#
# Copyright (c) 2022 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

# this script is to create two VMs with QEMU hypervisor.
# ESP, as OS provider, will boot these two nodes to build the RKE cluster.

#!/bin/bash -e

sudo apt update
sudo apt install qemu-kvm libvirt-clients libvirt-daemon-system virt-manager

VMLOC=$1

if [ -z $VMLOC ]; then echo "ERROR: image location not specified." && exit 1; fi

mydir=$(pwd)
vmsloc=$(cd ${VMLOC};pwd)/vms
imgloc=$vmsloc/images
mkdir -p $imgloc

function create_vm()
{
  name=$1
  cpus=$2
  mem_gb=$3
  disk0_gb=$4
  disk1_gb=$5
  mac=$6
  disk0_loc=${imgloc}/${name}-sda.img
  disk1_loc=${imgloc}/${name}-sdb.img
 
  echo "Createing VM name=$name,cpus=$cpus,mem_gb=$mem_gb,disk=$disk0_loc,$disk1_loc"

  sudo virsh destroy  $name 2>/dev/null || true
  sudo virsh undefine $name 2>/dev/null || true

  rm -f ${imgloc}/${name}-sda.img
  sudo qemu-img create -f qcow2 ${disk0_loc} ${disk0_gb}G
  rm -f ${imgloc}/${name}-sdb.img
  sudo qemu-img create -f qcow2 ${disk1_loc} ${disk1_gb}G
  template="${mydir}/template.xml"
  cp -f ${template} ${vmsloc}/${name}.xml
  sed -i -e "s,__MODIFY___NAME,$name," \
    -e "s,__MODIFY___MEM_IN_GIB,$mem_gb," \
    -e "s,__MODIFY___CPU_NUM,$cpus," \
    -e "s,__MODIFY___DISK0,$disk0_loc," \
    -e "s,__MODIFY___DISK1,$disk1_loc," \
    -e "s,__MODIFY___MAC,$mac," \
    $vmsloc/$name.xml
  sudo virsh define $vmsloc/$name.xml

}

create_vm rke1 2 4 20 20 52:54:00:c3:b1:cb
create_vm rke2 2 4 20 20 52:54:00:c3:b1:cc
