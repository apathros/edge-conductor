# Edge Conductor Tool: How to Deploy KIND Cluster

This document is about how to config and run Edge Conductor tool to deploy a
KIND cluster.

> *NOTE:*  KIND is not intended for production use.

## Preparation

Follow [HW Requirements for Edge Conductor Day-0
Host](../../README.md#hw-requirements-for-edge-conductor-day-0-host) and [OS and
System Requirements for Edge Conductor Day-0
Host](../../README.md#os-and-system-requirements-for-edge-conductor-day-0-host)
to prepare the Day-0 host hardware and software.

> *NOTE:*  For each KIND node, 2 CPU cores and 2 gigabytes (GB) memory are
> needed additionally.

Follow
[Build-and-Install-Edge-Conductor-Tool](../../README.md#build-and-install-edge-conductor-tool)
to build and install Edge Conductor tool.
Enter `_workspace` folder to run Edge Conductor tool.

## Edge Conductor Kit for KIND

An example of Edge Conductor Kit for KIND is under:

```
kit/
└── kind.yml
```

We will use this Edge Conductor Kit to deploy the KIND cluster in this document.

For more details of the Edge Conductor Kit, check the [Example of KIND
Kit.yml](../../kit/kind.yml)

## Custom Config

Modify the Edge Conductor Kit config file(kit/kind.yml) following
[Edge Conductor Configurations | Edge Conductor Kit
Introduction](ec-configurations.md#edge-conductor-kit-introduction), which is a
mandatory parameter for "conductor init".

## Init Edge Conductor Environment

Run the following commands to initialize the Edge Conductor environment:

```bash
./conductor init -c kit/kind.yml
```

## Build and Deploy KIND Cluster

Run the following commands to build and deploy KIND cluster:

```bash
./conductor cluster build
./conductor cluster deploy
```

The kubeconfig will be copied to the default path `~/.kube/config`.

## Check the KIND Cluster

Install the [kubectl tool (v1.20.0)](https://kubernetes.io/docs/tasks/tools/) to
interact with the target cluster.

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

## Remove the KIND cluster

To remove the KIND cluster, enter the command:

```bash
./conductor cluster remove
```

Copyright (c) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
