# How to enable SR-IOV network with Edge conductor

## Contents

  * [Supported SR-IOV NICs](#supported-sr-iov-nics)
  * [Preparation](#preparation)
  * [Basic configuration for SR-IOV NIC](#basic-configuration-for-sr-iov-nic)  

## Supported SR-IOV NICs

The following NICs are supported with this implementation
* Intel E800 Series
* Intel X700 Series
* Intel 82599ES

The follow NICs have been verified
* X722 NIC - deviceid: 0x37d2 0x37d0
* I350 NIC - deviceid: 0x1521
* X550 NIC - deviceid: 0x1563

## Preparation

Follow [HW Requirements for Edge Conductor Day-0 Host](../../README.md#hw-requirements-for-edge-conductor-day-0-host) and [OS and System Requirements for Edge Conductor Day-0 Host](../../README.md#os-and-system-requirements-for-edge-conductor-day-0-host) to prepare the Day-0 host hardware and software.

Follow [Build and Install Edge Conductor Tool](../../README.md#build-and-install-edge-conductor-tool) to build and install Edge Conductor tool.
Enter `_workspace` folder to run Edge Conductor tool.

Before SR-IOV network deployment, users need to setup the following:

* Hardware
  1. Prepare one or more bare metal servers as worker nodes which have two NICs, one SR-IOV NIC and one provisioning NIC in every SR-IOV worker node.
  2. Make sure VT-d and SR-IOV options in BIOS are enabled in SR-IOV worker nodes.
  3. Set correct time zone and system time in BIOS to avoid time mismatch between the Edge Conductor Day-0 machine and these bare metal servers.
* Operating System
  1. Ubuntu 20.04 is prefered.
&nbsp;

## Basic configuration for SR-IOV NIC

We use Edge Conductor Kit to deploy the cluster with SR-IOV feature enabled in this document.

SR-IOV feature will be enabled by modifying below part in [Example of RKE Kit.yml](../../kit/rke_ubuntu_20.04.yml).

Note: SR-IOV related CNI, device plugin and multus services will be deployed when add "intel-sriov-network" in selector section.

```yaml
Components:
  manifests:
  - "config/manifests/component_manifest.yml"
  selector:
  - name: "intel-sriov-network"
```

SR-IOV extension configuration can be customized through modifying default one in [sriov.yml](../../configs/extensions/sriov.yml).

Note: User needs to set the option "sriov_enabled" to "true" to enable SR-IOV Virtual Funtion creation automatically. If SRIOV PF device ID of worker nodes should be included in the "pfdevices", VF created device ID should be included in the "vfdevices". The "pfNames" can be used to specify which PF will be used to create VF. User needs to make sure the specified PF interface status is down.

```yaml
extension:
- name: sriov
  config:
  - name: sriov_enabled
    value: "false"
  # The SRIOV PF vendors ID
  - name: vendors
    value: "0x8086"
  # The SRIOV PF devices ID
  - name: pfdevices
    value: "0x37d0 0x37d2 0x15e4"
  # The total VFs number that will be created for containers and VMs
  - name: num_vfs
    value: "8"
  # The VFs number that will be created for VMs
  - name: num_vfio_vfs
    value: "1"
- name: nicselector
  config:
  # The SRIOV VF devices ID that is created
  - name: vfdevices
    value: '["1520", "1565", "37cd", "154c", "1889", "15c5"]'
  # The PF interface name
  - name: pfNames
    value: ""
- name: sriov-network-1
  config:
  - name: "type"
    value: "host-local"
  - name: "subnet"
    value: "10.56.217.0/24"
  - name: "rangeStart"
    value: "10.56.217.171"
  - name: "rangeEnd"
    value: "10.56.217.181"
  - name: "routes"
    value: '[{"dst": "0.0.0.0/0"}]'
  - name: "gateway"
    value: "10.56.217.1"
- name: sriov-net-vm
  config:
  - name: "type"
    value: "host-local"
  - name: "subnet"
    value: "10.56.218.0/24"
  - name: "rangeStart"
    value: "10.56.218.171"
  - name: "rangeEnd"
    value: "10.56.218.181"
  - name: "routes"
    value: '[{"dst": "0.0.0.0/0"}]'
  - name: "gateway"
    value: "10.56.218.1"
```

After finishing above configurations, we can start deployment of RKE cluster with SR-IOV feature enabled following [Deploy an RKE Cluster](cluster-deploy-RKE.md).


Copyright (C) 2021 Intel Corporation

SPDX-License-Identifier: Apache-2.0
