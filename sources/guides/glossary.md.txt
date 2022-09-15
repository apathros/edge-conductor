# Glossary

This glossary lists common terms and acronyms used in Edge Conductor guides and
specifications.

## General Terms

| Term    | Description |
| -------------- | ----- |
| Bare Metal | A node running without hypervisors. In a bare metal scenario, application workloads run on the operating system which runs on the hardware.
| IoT | Internet Of Things. The interconnection via the internet of computing devices embedded in everyday objects, enabling them to send and receive data.
| VM | Virtual Machine. A compute resource that uses software instead of physical hardware to run a program. Multiple VMs can run independently on the same physical machine, and with their own OS. A hypervisor uses direct access to the underlying machine to create the software environment for sharing and managing hardware resources.    


## Cloud Terms

| Term    | Description |
| -------------- | ----- |
| Control Plane Node | A Node that is running the Kubernetes API server. Systems can have 1 or multiple control plane nodes.
| Day-0 | A lifecycle term used in network automation and cloud deployment, where the high-level task on Day-0 is installation, Day-1 is configuration, and Day-2 is optimization.<br> The Edge Conductor deployment methods rely on a "Day-0" host machine to run the tools that download, build and provision the required software.
| EC | Edge Conductor. A framework to create, manage and operate Kubernetes clusters for IOT workloads.
| ESP | Edge Software Provisioner. A "just-in-time" open source provisioning tool described at: https://github.com/intel/edge-software-provisioner
| FDO | FIDO Device Onboard. Open source software that is developing an implementation of the [FIDO Device Onboard Specification](https://fidoalliance.org/fido2/). You can build and deploy the services that come with FDO using the Edge Conductor tool.
| Harbor | An open source container registry that secures artifacts and signs images as trusted. <br>When the Edge Conductor tool is initialized, it creates a local Harbor registry for images that you build.
| K8s | [Kubernetes](https://kubernetes.io/). An open-source system for automating deployment, scaling, and management of containerized applications.
| Kind | Kubernetes in Docker. A tool for running local Kubernetes clusters where each "node" is a Docker container. [(Learn more about Kind...)](https://kind.sigs.k8s.io/)
| kubectl | [Kubernetes command-line tool](https://kubernetes.io/docs/reference/kubectl/), which allows you to run commands against Kubernetes clusters.
| Node | A machine that can be a Kubernetes Control Plane Node or a Worker Node. A pod runs on a Node and a Node can have multiple pods. Depending on the cluster, a Node may be either a virtual or a physical machine. Each Node is managed by the control plane. The Kubernetes control plane automatically handles scheduling pods across the nodes in the cluster. [(Learn more about Kubernetes...)](https://kubernetes.io/docs/tutorials/kubernetes-basics/explore/explore-intro/)
| On-premise cluster | A mode of operation where all nodes in the cluster are physical (or virtual) machines. The Day-0 and management console functions are performed on machine(s) outside the cluster.
| Pod | Pods are the atomic unit on the Kubernetes platform. A Pod can group one or more application containers (such as Docker) that includes shared storage, IP address and information about how to run them. A Pod can contain different application containers which are relatively tightly coupled. [(Learn more about Kubernetes...)](https://kubernetes.io/docs/tutorials/kubernetes-basics/explore/explore-intro/)
| Virtual cluster | A mode of operation where all nodes in the cluster are virtualized as container images and Kubernetes is deployed on the user’s local machine. The Day-0 host and management console are also the user's local machine.
| Worker Node | A Node that is available to run application workloads and services. Nodes can be both Control Plane and Worker nodes at the same time (for example, a single node Kubernetes cluster).
Copyright (C) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
