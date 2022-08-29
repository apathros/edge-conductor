#
# Copyright (c) 2022 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

# this script is to set up a network bridge 10.10.10.1 for virtual machines 
# and install the dependency

#!/bin/bash

sudo apt update
sudo apt install bridge-utils

sudo ifconfig virbr1 down || true
sudo brctl delbr virbr1 || true
sudo brctl addbr virbr1

sudo ifconfig virbr1 10.10.10.1/24 up

IP_BRIDGE=10.10.10.1
echo "virtual bridge $IP_BRIDGE is up"

if ! sudo iptables -t nat -L | grep -q 10.10.10.0/24; then
    sudo iptables -t nat -A POSTROUTING -s 10.10.10.0/24 -j MASQUERADE
    sudo iptables -I FORWARD -s 10.10.10.0/24 -j ACCEPT
    sudo iptables -I FORWARD -d 10.10.10.0/24  -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
fi


