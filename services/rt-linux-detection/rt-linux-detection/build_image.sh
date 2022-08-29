#! /bin/bash
#
# Copyright (c) 2022 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

#Global variables
ROOTDIR=`pwd`
WORKDIR="$ROOTDIR/services/rt-linux-detection"

function checkRequiredPackages () {
   # check required software is installed
   TOOLS="docker bash"
   for i in ${TOOLS[@]}; do
      echo Checking for $i
      $i --version > /dev/null 2>&1
      if [ $? -ne 0 ]; then
         echo "No $i available in path. Please ensure that $i is installed and reachable."
         exit 1
      fi
   done
}

function build_docker_image () {
   #build rt-linux-detection docker image based on alpine
   cd $WORKDIR/rt-linux-detection
   chmod +x rt_linux_detect.sh
   docker build . -t rt-linux-detection:latest
} 

#
# main() is starting here
#

PROXIES="http_proxy https_proxy HTTP_PROXY HTTPS_PROXY"
for p in ${PROXIES[@]}; do
   export p=$p
done

checkRequiredPackages
build_docker_image

