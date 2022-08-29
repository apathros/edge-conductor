#! /bin/sh
#
# Copyright (c) 2022 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#
function check_linux_version () {
  # check if RT Kernel is detected 
  if uname -r | grep -q '\-rt'; then  
     echo "RT Kernel detected"
     label_nodes "yes"
  else
     echo "Non RT Kernel detected"
     label_nodes "no"
  fi
}

function label_nodes () {
  # check required software is installed
  RT_VERSION=$1
  echo "Labelling $NODE_NAME"
  kubectl label nodes $NODE_NAME RT_kernel_present=$RT_VERSION --overwrite=true
  if [ $? -ne 0 ];
  then
    echo "Labelling $NODE_NAME failed"
    kubectl label nodes $NODE_NAME RT_kernel_present=$RT_VERSION --overwrite=true
  fi
}

# main() is starting here
#
check_linux_version
while true; do sleep 100000; done
