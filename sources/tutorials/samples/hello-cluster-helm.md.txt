[Edge Conductor]: https://github.com/intel/edge-conductor
[Tutorials]: ../index.md
[Sample Applications]: ./index.md
[Hello Cluster!]: ./hello-cluster.md
[Hello Cluster! Helm]: ./hello-cluster-helm.md
[Web Indexing Application]: ./web-indexing.md
[Get Started]: ../../guides/get-started.md

[Edge Conductor] / [Tutorials] / [Sample Applications] / [Hello Cluster! Helm]

# Hello Cluster! Helm Version

This tutorial describes how to prepare an Edge Conductor Kit to deploy a Helm chart package. This tutorial uses Hello Cluster! application to demonstrate how to deploy the Hello Cluster! Helm package on Kind cluster


## Contents
- [Prerequisites](#prerequisites)

- [Prepare an Edge Conductor Kit for a Helm Chart Package Deployment](#prepare-an-edge-conductor-kit-for-a-helm-chart-package-deployment)

- [Deploy Hello Cluster Application by the Helm Chart on Edge Conductor](#deploy-hello-cluster-application-by-helm-chart-on-edge-conductor)

## Prerequisites

This tutorial provides a Helm Chart packaged with the Hello Cluster! application as a sample. Please find the sample **Helm Chart package** [HERE](../../../examples/sample_hello_cluster_helm).  The following description, command, configuration, and output will be based on this sample Helm package. The reader can change the Helm chart package as their own just make sure the corresponding part uses the customized package name and configure files as well. 

Follow [HW Requirements for Edge Conductor Day-0
Host](../../../README.md#hw-requirements-for-edge-conductor-day-0-host) and [OS and
System Requirements for Edge Conductor Day-0
Host](../../../README.md#os-and-system-requirements-for-edge-conductor-day-0-host)
to prepare the Day-0 host hardware and software.

> *NOTE:*  For each KIND node, 2 CPU cores and 2 gigabytes (GB) of memory are
> needed at least.

Follow
[Download and Build Edge Conductor Code (Internal users only)](../../../README.md#download-and-build-edge-conductor-code--internal-users-only-)
to build and install the Edge Conductor tool.
Enter `_workspace` folder to run the Edge Conductor tool.

```bash
edge-conductor$ cd _workspace
edge-conductor/_workspace$
```

Follow the [Installing Helm](https://helm.sh/docs/intro/install/) steps to
install the Helm CLI.

## Prepare an Edge Conductor Kit for a Helm Chart Package Deployment

This document guides the user to prepare an Edge Conductor Kit that deploys a Helm Chart Package on a KIND cluster.

This document uses three kinds of files:
 - Edge Conductor Kit configure file [kit.yml](../../../examples/sample_hello_cluster_helm/kit.yml): the file passed to the conductor tool directly
 - Component Manifest file [manifest.yml](../../../examples/sample_hello_cluster_helm/manifest.yml): define the component's attribute
 - Helm Chart Package:  [hello-cluster-helm.tgz ](../../../_workspace/sample_hello_cluster_helm/hello-cluster-helm.tgz): The Helm Chart packaged with the Hello Cluster! application to be deployed.  (*Note: This package will be generated after EC build).

The following sections explain how the 3 files co-work as an Edge Conductor Kit to install Helm Chart Package on the KIND cluster.


### Edge Conductor Kit Config File
This file is the top-level file of Edge Conductor Kit which will be passed to the conductor tool as the configuration of the target cluster and software component. Refer to [kit.yml](../../../examples/sample_hello_cluster_helm/kit.yml) for the detail.

```yaml
Use:
  - kit/kind.yml

Parameters:
  customconfig:
    registry:
      password: "123456"

Components:
  manifests:
  - "config/manifests/component_manifest.yml"
  - "config/sample_hello_cluster_helm/manifest.yml"
  selector:
  - name: hello-cluster-helm
```
In this example:
- It imports kind.yml, which is the official Kit for KIND cluster.
- It imports config/sample_hello_cluster_helm/manifest.yml as the supplemented manifest of the component manifest file to provide the configuration of the helm chart package. The below section [Component Manifest](#component-manifest) gives the detailed information of this supplemented manifest YAML file.
- It selects the hello-cluster-helm component (defined in the  [manifest.yml](../../../examples/sample_hello_cluster_helm/manifest.yml))  to be deployed on the Kind cluster. 

### Component Manifest
EC has a [default component manifest file](../../../configs/manifests/component_manifest.yml), which contains all default components supported by EC. Customers can write a supplement component manifest file. This example provides the supplement manifest in [manifest.yml](../../../examples/sample_hello_cluster_helm/manifest.yml).
Detailed definition of **components** refer to [EC Configuration](../../guides/ec-configurations.md) components section.

```yaml
Components:
  - name: hello-cluster-helm
    url: file://{{ .Workspace }}/config/sample_hello_cluster_helm/hello-cluster-helm-0.1.0.tgz
    type: helm
    images:
      - gcr.io/google-samples/node-hello:1.0
    supported-clusters:
      - kind
```
In this sample:
- It defines the component's name as "hello-cluster-helm", which will be used as the identifier to be selected by kit.yml selector section. 
- It uses the hello-cluster-helm-0.1.0.tgz package to install the helm chart. 
- It selects ‘helm’ type of the component. Refer to [EC Component](../../guides/components.md) for supported component type.
- It provides offline deployment capability with the list of image URLs under the "images" section. 
- It requests the service to be deployed based on KIND cluster.

## Deploy Hello Cluster Application by Helm Chart on Edge Conductor
After all the above steps are accomplished, we can go to deploy the service based on the created hello-cluster-helm Edge Conductor Kit. Following steps, deploy the Hello Cluster Application by Helm Chart based on the Edge Conductor Kit created above.

### Init Edge Conductor Environment
Run the "init" command with hello-cluster-helm Kit config file to initialize the Edge Conductor environment.

```bash
edge-conductor/_workspace$ ./conductor init -c ./config/sample_hello_cluster_helm/kit.yml
```
### Build and Deploy KIND Cluster
Run the following commands to build and deploy KIND cluster.
```bash
edge-conductor/_workspace$ ./conductor cluster build
edge-conductor/_workspace$ ./conductor cluster deploy
```

### Check the KIND Cluster
Install the [kubectl tool (v1.20.0 or above)](https://kubernetes.io/docs/tasks/tools/) to
interact with the target cluster.

```bash
kubectl get nodes
```
### Build and Deploy hello-cluster Service
To build and deploy the services, enter the commands:

```bash
edge-conductor/_workspace$ ./conductor service build
edge-conductor/_workspace$ ./conductor service deploy
```
Run the following commands to list the pods that are running the Hello Cluster application:

```bash
edge-conductor/_workspace$ kubectl get pods --output=wide
```
You will see output similar to:
```bash
NAME                              READY   STATUS    RESTARTS   AGE    IP           NODE          NOMINATED NODE   READINESS GATES
hello-cluster-helm-f86d6448f-2c98c   1/1     Running   0          114m   10.244.1.4   kind-worker   <none>           <none>
hello-cluster-helm-f86d6448f-r2dqc   1/1     Running   0          114m   10.244.1.3   kind-worker   <none>           <none>
```
### Access the Service
Use port-forward to access the hello cluster application
Run the following command:
```bash
kubectl port-forward -n default service/hello-cluster-helm 5999:8080
```
It should be output like below:
```bash
Forwarding from 127.0.0.1:5999 -> 8080
Forwarding from [::1]:5999 -> 8080
```
In a new terminal, run the following command to access the service:

```bash
curl http://127.0.0.1:5999
```
The response to a successful request is a hello message:

```bash
Hello, world!
Version: 1.0.0
Hostname: hello-cluster-helm-f86d6448f-2c98c
```
### Clean Up
To remove the whole KIND cluster and deinit EC by running the following commands:

```bash
edge-conductor/_workspace$ ./conductor cluster remove
edge-conductor/_workspace$ ./conductor deinit
```

## What's Next

Congratulations! You have deployed an application using a Helm chart.
Next, you can try to deploy a web indexing service on the Kubernetes cluster.

-----\
Previous Tutorial: [Hello Cluster!]\
Next Tutorial: [Web Indexing Application]\
\
Back to: [Tutorials](/docs/tutorials/index.md)


Copyright (C) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
