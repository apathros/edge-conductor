[Edge Conductor]: https://github.com/intel/edge-conductor
[Tutorials]: ../index.md
[Sample Applications]: ./index.md
[Hello Cluster!]: ./hello-cluster.md
[Hello Cluster! Helm]: ./hello-cluster-helm.md
[Web Indexing Application]: ./web-indexing.md
[Get Started]: ../../guides/get-started.md
[EC Configuration]: ../../guides/ec-configurations.md

[Edge Conductor] / [Tutorials] / [Sample Applications] / [Hello Cluster!]

# Hello Cluster!

This tutorial describes how to prepare an Edge Conductor Kit to deploy a hello-cluster service based on the KIND cluster.

## Contents

- [Prerequisites](#prerequisites)
- [Prepare an Edge Conductor Kit to Deploy a Service based on KIND Cluster](#prepare-an-edge-conductor-kit-to-deploy-a-service-based-on-kind-cluster)
- [Deploy hello-cluster Service based on the Edge Conductor Kit](#deploy-hello-cluster-service-based-on-the-edge-conductor-kit)

## Prerequisites

Follow [HW Requirements for Edge Conductor Day-0
Host](../../../README.md#hw-requirements-for-edge-conductor-day-0-host) and [OS and
System Requirements for Edge Conductor Day-0
Host](../../../README.md#os-and-system-requirements-for-edge-conductor-day-0-host)
to prepare the Day-0 host hardware and software.

> *NOTE:*  For each KIND node, 2 CPU cores and 2 gigabytes (GB) memory are
> needed at least.

Follow
[Download and Build Edge Conductor Code (Internal users only)](../../../README.md#download-and-build-edge-conductor-code--internal-users-only-)
to build and install Edge Conductor tool.
Enter `_workspace` folder to run Edge Conductor tool.
```bash
edge-conductor$ cd _workspace
edge-conductor/_workspace$
```

## Prepare an Edge Conductor Kit to Deploy a Service based on KIND Cluster


This document guides to the user to prepare an Edge Conductor Kit that deploys a service on a KIND cluster.

This document uses three kinds of YAML files:
 - Edge Conductor Kit configure file [kit.yml](../../../examples/sample_hello_cluster/kit.yml): the file passed to the conductor tool directly
 - Component Manifest file [manifest.yml](../../../examples/sample_hello_cluster/manifest.yml): define the component's attribute
 - Service Deployment file [hello-cluster.yml](../../../examples/sample_hello_cluster/hello-cluster.yml): define the way to deploy the service to the KIND cluster

The following sections explains how the 3 files co-work as an Edge Conductor Kit to provide a "hello-cluster" service on the KIND cluster.


### Edge Conductor Kit Config File
This file is the top level file of Edge Conductor Kit which will be passed to the conductor tool as configuration of target cluster and software component. Refer to [kit.yml](../../../examples/sample_hello_cluster/kit.yml) for the detail.

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
  - "config/sample_hello_cluster/manifest.yml"
  selector:
  - name: hello-cluster
```
In this example:
- It imports kind.yml, which is the official Kit for KIND cluster.
- It imports config/sample_hello_cluster/manifest.yml as the supplemented manifest of the component manifest file to provide the configuration of the hello-cluster service. The below section [Component Manifest](#component-manifest) gives the detailed information of this supplemented manifest YAML file.
- It selects the hello-cluster as the component to be deployed on the Kind cluster. 

### Component Manifest
EC has a [default component manifest file](../../../configs/manifests/component_manifest.yml), which contains all default components supported by EC. Customer can write a supplement component manifest file. This example provides the supplement manifest in [manifest.yml](../../../examples/sample_hello_cluster/manifest.yml).
Detailed definition of **components** refer to [EC Configuration](../../guides/ec-configurations.md) components section.

```yaml
Components:
  - name: hello-cluster
    url: file://{{ .Workspace }}/config/sample_hello_cluster/hello-cluster.yml
    type: yaml
    images:
      - gcr.io/google-samples/node-hello:1.0
    supported-clusters:
      - kind
    namespace: hello-cluster
```
In this sample:
- It defines the component's name as "hello-cluster". 
- It uses the hello-cluster.yml file as the deployment YAML file. Refer to section [Deployment File](#deployment-file).
- It provides offline deployment capability with the list of image URLs under the "images" section. 
- It requests the service to be deployed based on KIND cluster.

### Deployment File
The deployment file in this Kit example describes a desired status in a Deployment by Kubectl command. The formal of this YAML file should follow the definition of Kubernetes deployment file. Refer to the [hello-cluster.yml](../../../examples/sample_hello_cluster/hello-cluster.yml) for the detail.
In this sample:
 It creates a ReplicaSet to bring up two hello-cluster Pods.

## Deploy hello-cluster Service based on the Edge Conductor Kit
After all the above steps are accomplished, we can go to deploy the service based on the created hello-cluster Edge Conductor Kit. Following steps deploy hello-cluster service based on the Edge Conductor Kit created above.

### Init Edge Conductor Environment
Run the "init" command with hello-cluster Kit config file to initialize the Edge Conductor environment.

```bash
edge-conductor/_workspace$ ./conductor init -c ./config/sample_hello_cluster/kit.yml
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
NAME                            READY   STATUS    RESTARTS   AGE    IP           NODE          NOMINATED NODE   READINESS GATES
hello-cluster-f86d6448f-2c98c   1/1     Running   0          114m   10.244.1.4   kind-worker   <none>           <none>
hello-cluster-f86d6448f-r2dqc   1/1     Running   0          114m   10.244.1.3   kind-worker   <none>           <none>
```
### Access the Service
Use port-forward to access the hello cluster application
Run the following command:
```bash
kubectl port-forward -n default service/hello-cluster 5999:8080
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
Hostname: hello-cluster-f86d6448f-2c98c
```
### Clean Up
To remove the whole KIND cluster and deinit EC by running the following commands:

```bash
edge-conductor/_workspace$ ./conductor cluster remove
edge-conductor/_workspace$ ./conductor deinit
```

##  What's Next

Congratulations! You have deployed an application based on Edge Conductor (EC) using an Edge Conductor Kit.
Next you can try to deploy a cluster using a Helm chart.

Next Tutorial: [Hello Cluster! Helm](./hello-cluster-helm.md)

Back to: [Tutorials](/docs/tutorials/index.md)

Copyright (c) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
