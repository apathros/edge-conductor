# Get Started

First time user? This guide has details on how to get up and running with Edge
Conductor, using the KIND cluster and default configuration files.

## Contents

*   [Contents](#contents)
*   [Hardware and Software Requirements](#get-started)
*   [Retrieve, Make, and Install the Tool](#retrieve-make-and-install-the-tool)
*   [Configure Password With Registry](#configure-password-with-registry)
*   [Initialize Edge Conductor Tool](#initialize-edge-conductor-tool)
*   [Build and Deploy a Kind Cluster](#build-and-deploy-a-kind-cluster)
*   [Build and Deploy Services on the Target Cluster](#build-and-deploy-services-on-the-target-cluster)
*   [Interact with Nodes](#interact-with-nodes)
*   [Remove the Kind Cluster](#remove-the-kind-cluster)
*   [Deinit Edge Conductor Services](#deinit-edge-conductor-services)
*   [Next Steps](#next-steps)


## Hardware and Software Requirements

Be sure your host meets the following requirements.

Hardware:

*   CPU: 2 or more cores
*   Memory: 2 gigabytes (GB) or greater.
*   Storage: 10 GB or greater available storage is required to build and run the
    Day-0 environment.
*   Internet connection: Internet connectivity is necessary to download and use
    some features.

> *NOTE:*  For each kind node, you need to add 2 CPU cores and 2 gigabytes (GB)
memory.

OS and System:

*   Ubuntu 18.04+ LTS
*   make 4.1+
*   DockerCE
    * 20.10.3+ (for DockerCE v20)
    * After you install DockerCE, configure a user group so you can use Docker
      without the `sudo` command. For details, follow the Docker steps:
      [Post-installation steps for
      Linux](https://docs.docker.com/engine/install/linux-postinstall/).
*   git 2.33.0+

Additional software:

*   Install [kubectl (v1.20.0)](https://kubernetes.io/docs/tasks/tools/) to
    interact with the clusters created with Edge Conductor. For details, follow
    the Kubernetes steps: [Install and Set Up kubectl on Linux](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/).

Proxy setup:

*   Be sure to set your proxies correctly. Some environments require network
    proxies for Docker operations (e.g. docker pull, docker push, docker run,
    and so on).
*   For http_proxy/HTTP_PROXY and https_proxy/HTTPS_PROXY, you must test they
    are actually working. Also note that the no_proxy/NO_PROXY list **MUST NOT
    contain spaces** between the addresses.


## Retrieve, Make, and Install the Tool

1.  Retrieve the code from the repository:

* (For internal users only) Clone the code from the git repo:

  ```bash
  git clone https://github.com/intel/edge-conductor.git edge-conductor
  ```

* Use tar and ln to compress and link the code package, for example, `edge_conductor_<version>.tar.gz`:

  ```bash
  tar zxvf edge_conductor_<version>.tar.gz
  ln -sf edge_conductor_<version> edge-conductor
  ```

2. Make the code:

    ```
    cd edge-conductor
    make
    ```

    You will see output similar to:

    ```
    make -C api/schemas
    make -C api/proto build
    go mod tidy
    go run build/plugingenerator.go
    go vet ./pkg/... ./cmd/...
    Going to format and build code.
    ...
    go build -v -o _workspace/bin/conductor cmd/conductor/main.go
    ...
    ```

    If `make` is successful, no ERROR messages are displayed on the console.

    

3.  Check installation:

    With the `make` command, binary files are generated under `_workspace`
    folder and configuration files are also copied to `_workspace`.

    Change to the `_workspace` directory, which is created after `make`
    completes:

    ```
    cd _workspace
    ```

    The file structure generated under the `_workspace` folder is:

    ```
    _workspace/
    ├── bin
    ├── config
    ├── kit
    ├── conductor -> bin/conductor
    ├── kubectl -> bin/kubectl
    └── workflow
    ```

## Configure Password With Registry

Set up a local registry by editing the `kind.yml` config file in
`_workspace/kit` folder using your preferred text editor.

Add your password details:


* This password must be within `"` symbols and there must be a space before
   the first `"`. We recommend that it contains at 
   least 8 characters with 1 lowercase letter, 1 uppercase letter and 1 numeric
   character. For example, a valid password line is: `password: "PassWord4"`.

```
Parameters:
  customconfig:
    registry:
      password:
```

## Initialize Edge Conductor Tool

Initialize the Edge Conductor tool using the `kind.yml` file you edited in the
previous step.  If you are not in the `_workspace` directory, cd to it before doing the init:

```
cd _workspace
./conductor init -c kit/kind.yml
```

> *NOTE:*  The `kit/` path is mandatory for the `conductor init`
command. Check for more Edge Conductor Kits [here](../../kit/).

You will see output similar to:

```
INFO[0000] Init Edge Conductor
INFO[0000] ==
INFO[0000] Current workflow: init
...
INFO[0005] workflow finished
INFO[0005] ==
INFO[0005] Done
```

After `conductor init` completes, the credentials are stored in your local
registry.

## Build and Deploy a Kind Cluster

Enter the command:

```
./conductor cluster build
```

You will see output similar to:

```
INFO[0000] Edge Conductor - Build Cluster
INFO[0000] ==
INFO[0000] Current workflow: cluster-build
...
INFO[0005] workflow finished
INFO[0005] ==
INFO[0005] Done
```

Enter the command:

```
./conductor cluster deploy
```

You will see output similar to:

```
INFO[0000] Edge Conductor - Deploy Cluster
INFO[0000] ==
INFO[0000] Current workflow: cluster-deploy
...
INFO[0000] Deploying kind...
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.23.4)
 ✓ Preparing nodes
 ✓ Writing configuration
 ✓ Starting control-plane
 ✓ Installing CNI
 ✓ Installing StorageClass
...
INFO[0005] workflow finished
INFO[0005] ==
INFO[0005] Done
```

## Build and Deploy Services on the Target Cluster

Enter the command:

```
./conductor service build
```

You will see output similar to:

```
INFO[0000] Edge Conductor - Build Services
INFO[0000] ==
INFO[0000] Current workflow: service-build
...
INFO[0007] Downloading service resource files.
INFO[0008] Downloaded successfully.
...
INFO[0009] workflow finished
INFO[0009] ==
INFO[0009] Done
```

Enter the command:

```
./conductor service deploy
```

You will see output similar to:

```
INFO[0000] Edge Conductor - Deploy Services
INFO[0000] ==
INFO[0000] Current workflow: service-deploy
...
INFO[0001] Yaml services will be deployed.
INFO[0002] Helm services will be deployed.
INFO[0003] Successfully installed releases...
...
INFO[0004] workflow finished
INFO[0004] ==
INFO[0004] Done
```


## Interact With Nodes

After the services are deployed, you can interact with the target cluster using
the kubeconfig exported by the Edge Conductor tool.

To learn more about kind clusters, refer to the
[Kind Quick Start guide](https://kind.sigs.k8s.io/docs/user/quick-start).

### Get Nodes

Check the nodes that are available with the command:

```
kubectl get nodes
```

You will see output similar to:

```
NAME                 STATUS   ROLES                  AGE     VERSION
kind-control-plane   Ready    control-plane,master   2m49s   v1.23.4
kind-worker          Ready    <none>                 68s     v1.23.4
```

### Get Services

Check the services deployed to the kind cluster with the command:

```
kubectl get services,pods -A
```

You will see output similar to:

```
NAMESPACE       NAME                                         TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
default         service/kubernetes                           ClusterIP      10.96.0.1       <none>        443/TCP                         8m51s
ingress-nginx   service/ingress-nginx-controller             NodePort       10.96.184.207   <none>        80:32672/TCP,443:31014/TCP      8m36s
ingress-nginx   service/ingress-nginx-controller-admission   ClusterIP      10.96.255.153   <none>        443/TCP                         8m36s
kube-system     service/kube-dns                             ClusterIP      10.96.0.10      <none>        53/UDP,53/TCP,9153/TCP          8m49s
portainer       service/portainer                            LoadBalancer   10.96.163.134   <pending>     9000:31121/TCP,8000:30837/TCP   8m36s
prometheus      service/prometheus-alertmanager              ClusterIP      10.96.8.136     <none>        80/TCP                          8m34s
prometheus      service/prometheus-blackbox-exporter         ClusterIP      10.96.233.20    <none>        9115/TCP                        8m33s
prometheus      service/prometheus-kube-state-metrics        ClusterIP      10.96.146.195   <none>        8080/TCP                        8m34s
prometheus      service/prometheus-node-exporter             ClusterIP      None            <none>        9100/TCP                        8m34s
prometheus      service/prometheus-pushgateway               ClusterIP      10.96.241.40    <none>        9091/TCP                        8m34s
prometheus      service/prometheus-server                    ClusterIP      10.96.111.23    <none>        80/TCP                          8m34s

NAMESPACE            NAME                                                READY   STATUS              RESTARTS   AGE
ingress-nginx        pod/ingress-nginx-admission-create-7lxjr            0/1     Completed           0          8m34s
ingress-nginx        pod/ingress-nginx-admission-patch-t867v             0/1     Completed           0          8m34s
ingress-nginx        pod/ingress-nginx-controller-6c85cb7b5d-qj9wn       1/1     Running             0          8m34s
kube-system          pod/coredns-558bd4d5db-68kbq                        1/1     Running             0          8m34s
kube-system          pod/coredns-558bd4d5db-t5jwg                        1/1     Running             0          8m34s
kube-system          pod/etcd-kind-control-plane                         1/1     Running             0          8m37s
kube-system          pod/kindnet-7r42x                                   1/1     Running             0          8m34s
kube-system          pod/kube-apiserver-kind-control-plane               1/1     Running             0          8m37s
kube-system          pod/kube-controller-manager-kind-control-plane      1/1     Running             0          8m37s
kube-system          pod/kube-multus-ds-amd64-rstfb                      1/1     Running             0          8m34s
kube-system          pod/kube-proxy-4gcth                                1/1     Running             0          8m34s
kube-system          pod/kube-scheduler-kind-control-plane               1/1     Running             0          8m37s
local-path-storage   pod/local-path-provisioner-547f784dff-xh4wq         1/1     Running             0          8m34s
portainer            pod/portainer-fc654c454-nnls8                       1/1     Running             0          8m34s
prometheus           pod/prometheus-alertmanager-84f4ccb57d-gqpt5        2/2     Running             0          8m34s
prometheus           pod/prometheus-blackbox-exporter-5fcfbd67b6-m9rmm   1/1     Running             0          8m33s
prometheus           pod/prometheus-kube-state-metrics-bc6c8c864-z729b   1/1     Running             0          8m34s
prometheus           pod/prometheus-node-exporter-m4dxx                  1/1     Running             0          8m17s
prometheus           pod/prometheus-pushgateway-685556f875-nxt9j         1/1     Running             0          8m34s
prometheus           pod/prometheus-server-c8f78b8d6-kgphw               2/2     Running             0          8m34s
```

## Remove the Kind Cluster

To remove the kind cluster, enter the command:

```
./conductor cluster remove
```

You will see output similar to:

```
INFO[0000] Edge Conductor - Remove Cluster
INFO[0000] ==
INFO[0000] Current workflow: cluster-remove
...
INFO[0000] Removing kind...
Deleting cluster "kind" ...
INFO[0001] workflow finished
INFO[0001] ==
INFO[0001] Done
```

## Deinit Edge Conductor Services

> *NOTE:*  If you are going to try more Edge Conductor tutorials, skip this step
> and go to the [Next Steps](#next-steps) section.

Run the following command to stop Edge Conductor services. Using the ``--purge``
option will delete Edge Conductor configuration files and the certification and
keys generated by the Edge Conductor tool.

```bash
./conductor deinit --purge
```

You will see output similar to:

```
INFO[0000] Deinit Edge Conductor
INFO[0000] ==
INFO[0000] Current workflow: deinit
...
INFO[0051] workflow finished
INFO[0051] ==
INFO[0051] Done
```

> *NOTE:*  This `conductor deinit` operation will not destroy the target
cluster, it will only stop services launched at `conductor init` stage. If you
want to restart Edge Conductor services, you must redo the `conductor init`
step to initialize the environment.


## Next Steps

Congratulations on running a simple setup using Edge Conductor.

Next, you can explore the Edge Conductor tool functionality and try different
types of clusters. For details, refer to [Tutorials](../tutorials/index.md).





Copyright (c) 2022 Intel Corporation SPDX-License-Identifier: Apache-2.0
