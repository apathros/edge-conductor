[Edge Conductor]: https://github.com/intel/edge-conductor
[Tutorials]: ../index.md
[Deploy a KIND cluster]: ./deploy_a_kind_cluster.md

[Edge Conductor] / [Tutorials] / [Deploy a KIND cluster]

# Deploy a KIND cluster

[KIND](https://github.com/kubernetes-sigs/kind) is a tool for testing kubernetes, it runs a local kubernetes cluster using Docker itself.

Follow these steps below to deploy a Kubernetes in Docker (KIND) deployment with Edge Conductor tool.

## Contents

* [Prerequisites](#prerequisites)
* [Configure KIND deployment](#configure-kind-deployment)
* [Initialize Edge Conductor](#initialize-edge-conductor)
* [Build and Deploy KIND Cluster](#build-and-deploy-kind-cluster)
* [Check the KIND Cluster](#check-the-kind-cluster)
* [What's Next](#whats-next)

## Prerequisites

First, make sure your host meets the following requirements to run this tutorial:

1. [Hardware requirements for the Edge Conductor Day-0 Host](../../../README.md#hw-requirements-for-edge-conductor-day-0-host). This tutorial will deploy a 2-node KIND cluster (1 control plane and 1 worker). For each KIND node, you need to add 2 CPU cores and 2 gigabytes (GB) memory.

2. Follow [Build and Install the Tool](./how_to_install_tools.md#build-and-install-the-tool) to build and install Edge Conductor tool.

3. Enter the `_workspace` folder to run the Edge Conductor tool.

## Edge Conductor kit for KIND

An example of Edge Conductor kit for KIND is under:

```shell
kit/
└── kind.yml
```

We will use [this Edge Conductor kit](../../kit/kind.yml) to deploy the KIND cluster in this document.

## Custom Config

Open the Edge Conductor kit config file (kit/kind.yml) and modify the password of the local registry. Edge Conductor tool will launch a local registry as the storage of binary files, container images and configuration files. This registry password will be used to create user authentication.


```yaml
Parameters:
  customconfig:
    registry:
      password: "<passWord_nnnn>"
```

> *Note:* The password must be surrounded by double quotes (`"`). We recommend that it contains at least 8 characters with 1 lowercase letter, 1 uppercase letter and 1 numeric character.

Check more details of the [Edge Conductor kit configuration](../../guides/ec-configurations.md) here.

## Init Edge Conductor Environment

Run the following commands to initialize the Edge Conductor environment:

```bash
./conductor init -c kit/kind.yml
```

You will see output similar to:

```shell
INFO[0000] Init Edge Conductor
INFO[0000] ==
INFO[0000] Current workflow: init
...
INFO[0005] workflow finished
INFO[0005] ==
INFO[0005] Done
```

## Build and Deploy KIND Cluster

Run the following command to build KIND Cluster:

```bash
./conductor cluster build
```

You will see output similar to:

```bash
INFO[0000] Init Edge Conductor
INFO[0000] ==
INFO[0000] Top Config File: kit/kind.yml
...
INFO[0008] workflow finished
INFO[0008] Connecting Plugin docker-run
INFO[0008] ==
INFO[0008] Done
```

Run the following command to deploy KIND Cluster:

```bash
./conductor cluster deploy
```

You will see output similar to:

```bash
INFO[0000] Edge Conductor - Deploy Cluster
INFO[0000] ==
...
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.21.1) 🖼
 ✓ Preparing nodes 📦 📦 📦 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
 ✓ Joining worker nodes 🚜
Set kubectl context to "kind-kind"
...
INFO[0094] workflow finished
INFO[0094] ==
INFO[0094] Done
```

The kubeconfig will be copied to the default path `~/.kube/config`.

> *NOTE:*  If you export KUBECONFIG, you need to unset the KUBECONFIG or copy
`~/.kube/config` to your KUBECONFIG directory.

## Check the KIND Cluster

Run the following commands to check the nodes that are available and the
services deployed to the KIND cluster:

```bash
kubectl get pods,svc,nodes -A
```

You will see output similar to:

```shell
NAMESPACE            NAME                                             READY   STATUS    RESTARTS   AGE
kube-system          pod/coredns-558bd4d5db-4qjmq                     1/1     Running   0          11m
kube-system          pod/coredns-558bd4d5db-fckgb                     1/1     Running   0          11m
kube-system          pod/etcd-kind-control-plane                      1/1     Running   0          11m
kube-system          pod/kindnet-djzbj                                1/1     Running   0          11m
kube-system          pod/kindnet-l5x8b                                1/1     Running   0          11m
kube-system          pod/kube-apiserver-kind-control-plane            1/1     Running   0          11m
kube-system          pod/kube-controller-manager-kind-control-plane   1/1     Running   0          11m
kube-system          pod/kube-proxy-7vrf8                             1/1     Running   0          11m
kube-system          pod/kube-proxy-dvgw9                             1/1     Running   0          11m
kube-system          pod/kube-scheduler-kind-control-plane            1/1     Running   0          11m
local-path-storage   pod/local-path-provisioner-547f784dff-28jlh      1/1     Running   0          11m

NAMESPACE     NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                  AGE
default       service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP                  12m
kube-system   service/kube-dns     ClusterIP   10.96.0.10   <none>        53/UDP,53/TCP,9153/TCP   12m

NAMESPACE   NAME                      STATUS   ROLES                  AGE   VERSION
            node/kind-control-plane   Ready    control-plane,master   12m   v1.21.1
            node/kind-worker          Ready    <none>                 11m   v1.21.1
```

Be sure the status for all pods is `Running`.

## What's Next

Congratulations! You have deployed a KIND cluster with Edge Conductor tool!

Next Tutorial: [Example: Hello Cluster!](../samples/hello-cluster.md)

Back to: [Edge Conductor Basics](./index.md)

Copyright (C) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
