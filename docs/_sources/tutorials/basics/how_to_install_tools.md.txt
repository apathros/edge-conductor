[Edge Conductor]: https://github.com/intel/edge-conductor
[Tutorials]: ../index.md
[How to install tools]: ./how_to_install_tools.md

[Edge Conductor] / [Tutorials] / [How to install tools]

# Prepare the Edge Conductor tool

This tutorial shows you how to prepare and install the Edge Conductor tool.

## Contents
*   [Before You Begin](#before-you-begin)
*   [Build and Install the Tool](#build-and-install-the-tool)
*   [What's Next](#whats-next)

##  Before You Begin

Be sure your host meets the following requirements.

Hardware:

*   CPU: 2 or more cores
*   Memory: 2 gigabytes (GB) or greater.
*   Storage: 10 GB or greater available storage is required to build and run the Day-0 environment.
*   Internet connection: Internet connectivity is necessary to download and use some features.

> *NOTE:*  For each KIND node, you need to add 2 CPU cores and 2 gigabytes (GB) memory.

OS and System:

*   Ubuntu 18.04+
*   make 4.1+
*   DockerCE
    * 18.09.11+ (for DockerCE v18)
    * 19.03.13+ (for DockerCE v19)
    * After you install DockerCE, configure a user group so you can use Docker without the `sudo` command.
      For details, follow the Docker steps: [Post-installation steps for Linux](https://docs.docker.com/engine/install/linux-postinstall/).
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
*   After you set your proxies, you can enter some commands to verify your proxies.

   ```bash
      docker pull hello-world
   ```

##  Build and Install the Tool

### Get the code from the repository:

* (For internal users only) Get the code from the git repo: 

  ```bash
  git clone https://github.com/intel/edge-conductor.git edge-conductor
  ```

* Unzip the code package, for example, `edge_conductor_<version>.tar.gz`:

  ```bash
  tar zxvf edge_conductor_<version>.tar.gz
  ln -sf edge_conductor_<version> edge-conductor
  ```

### Build the code:

```bash
cd edge-conductor
make
```

If `make` is successful, no ERROR messages are displayed on the console.

You will see output similar to the following.

```bash
make -C api/schemas
make -C api/proto build
go mod tidy
go run build/plugingenerator.go
go vet ./pkg/... ./cmd/...
Going to format and build code.
go build -v -o _workspace/bin/conductor cmd/conductor/main.go
```

The `make` command does the following:

*   Creates the `_workspace` folder.
*   Generates binary files in the `_workspace` folder.
*   Copies configuration files to the `_workspace` folder.

### Change to the `_workspace` folder:

```bash
cd _workspace
```

The file structure generated under the `_workspace` folder is:

```bash
_workspace/
├── bin
├── config
├── conductor -> bin/conductor
├── kubectl -> bin/kubectl
├── services
└── workflow

```


## What's Next

Congratulations on running a simple setup using Edge Conductor.

Next Tutorial: [Deploy a KIND cluster](./deploy_a_kind_cluster.md)

Back to: [Edge Conductor Basics](./index.md)

Copyright (C) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
