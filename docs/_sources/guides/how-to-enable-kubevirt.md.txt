# How to enable Kubevirt with Edge conductor

## Contents

  * [Preparation](#preparation)
  * [Basic configuration for kubevirt](#basic-configuration-for-kubevirt)

## Preparation

Follow [HW Requirements for Edge Conductor Day-0 Host](../../README.md#hw-requirements-for-edge-conductor-day-0-host) and [OS and System Requirements for Edge Conductor Day-0 Host](../../README.md#os-and-system-requirements-for-edge-conductor-day-0-host) to prepare the Day-0 host hardware and software.

Follow [Build and Install Edge Conductor Tool](../../README.md#build-and-install-edge-conductor-tool) to build and install Edge Conductor tool.
Enter `_workspace` folder to run Edge Conductor tool.

Before Kubevirt deployment, users need to setup the following:

* Validate Hardware Virtualization Support
  1. Follow Kubevirt installation guidelines to make sure hardware virtualization is enabled in worker nodes. refer to: [Kubevirt User Guide](https://kubevirt.io/user-guide/operations/installation/).
* Operating System
  1. Ubuntu 20.04 is prefered.
&nbsp;

## Basic configuration for kubevirt

We use Edge Conductor Kit to deploy the cluster with Kubevirt enabled in this document.

Kubevirt will be enabled by modifying below part in [Example of RKE Kit.yml](../../kit/rke_ubuntu_20.04.yml).


```yaml
Components:
  manifests:
  - "config/manifests/component_manifest.yml"
  selector:
  - name: kubevirt-operator
  - name: kubevirt-cr


```
To make sure kubevirt can access QEMU, add the follow lines into the file "/etc/apparmor.d/usr.sbin.libvirtd" of each worker nodes, then reload appArmor through executing "systemctl reload apparmor.service".

```
    /usr/libexec/virtiofsd PUx,
    /usr/libexec/qemu-kvm PUx,
```


After finishing above configuration, we can start deployment of KIND, RKE, or CAPI cluster with kubevirt enabled following [Deploy a Target Cluster](../../README.md#edge-conductor-documentation).


Copyright (C) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
