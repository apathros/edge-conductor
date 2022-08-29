# Edge Conductor Tool: How to Deploy RKE Cluster

This document is about how to config and run Edge Conductor tool to deploy a RKE cluster.

## Preparation

Follow [HW Requirements for Edge Conductor Day-0 Host](../../README.md#hw-requirements-for-edge-conductor-day-0-host) and [OS and System Requirements for Edge Conductor Day-0 Host](../../README.md#os-and-system-requirements-for-edge-conductor-day-0-host) to prepare the Day-0 host hardware and software.

Follow [Build-and-Install-Edge-Conductor-Tool](../../README.md#build-and-install-edge-conductor-tool) to build and install Edge Conductor tool.
Enter `_workspace` folder to run Edge Conductor tool.

Before the RKE deployment, users need to:
1. Make sure the nodes meet hardware and software requirements for RKE cluster deployment: [RKE installation requirements](https://rancher.com/docs/rancher/v2.x/en/installation/requirements)
1. Provision nodes according to the requirements in [RKE official website](https://rancher.com/docs/rke/latest/en/os/).
1. The ntp (Network Time Protocol) package should be installed and start the service to sync time for all nodes. This prevents errors with certificate validation that can occur when the time is not synchronized between the client and server.
1. Some distributions of Linux may have default firewall rules that block communication with Helm. We recommend disabling firewalld. For Kubernetes 1.19 and 1.20, firewalld must be turned off.

## Edge Conductor Kit for RKE

Examples of Edge Conductor Kit for RKE are under:
```
kit/
├── rke_preinstalled.yml
└── rke_ubuntu_20.04.yml
```

We will use this Edge Conductor Kit to deploy the RKE cluster in this document.

For more details of the Edge Conductor Kit, check the [Example of RKE Edge Conductor Kit](../../kit/rke_preinstalled.yml)

## Custom Config

Modify the Edge Conductor Kit config file (kit/rke_preinstalled.yml) following [Edge Conductor Configurations | Edge Conductor Kit Introduction](ec-configurations.md#edge-conductor-kit-introduction), which is a mandatory parameter for "conductor init".

Follow the instruction in each of the Kit config files to config the node list and finish the preconditions.

## Init Edge Conductor Environment

Run the "init" commands with any RKE Kit config file to initialize the Edge Conductor environment.

For example, for preinstalled systems:

```shell
./conductor init -c kit/rke_preinstalled.yml
```

For bare metal setup to start from OS provisioning (in offline mode):

```shell
./conductor init -c kit/rke_ubuntu_20.04.yml
```

The RKE cluster configurations will be generated automatically. The "rke_cluster.yml" and ".rkestate" files will be exported to "~/.ec/rke/cluster" folder by default, users can also specify an export folder in RKE Kit config file with "Cluster" - "export_config_folder".


## OS Provisioning

If an Kit config to start from OS provisioning is selected, follow the [OS Provisioning](./os-provider-esp.md) guide to finish OS installation before cluster deployment.

## SSH Access

To deploy RKE cluster, make sure it is able to access all hosts in the cluster from Day-0 host.

```
ssh-copy-id -i < your ssh key name on Day-0 > < user >@< host >
```

## Build and Deploy RKE Cluster

Run the following commands to build and deploy RKE cluster.

```
./conductor cluster build
./conductor cluster deploy
```

The kubeconfig will be copied to the default path `~/.kube/config`.

## Check the RKE Cluster

Install the [kubectl tool (v1.20.0)](https://kubernetes.io/docs/tasks/tools/) to interact with the target cluster.

```bash
kubectl get nodes
```

## Continue to Deploy Services

To build and deploy the services, enter the commands:

```bash
./conductor service build
./conductor service deploy
```

> Use `--kubeconfig` to specify the kubeconfig if you don't want to use the default config file from `~/.kube/config`.

To deploy service rook-ceph and rook-ceph-cluster, please ensure these is at least one additional disk with more than 1GB capacity left for ceph osd deployment on any one worker node.

Copyright (c) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
