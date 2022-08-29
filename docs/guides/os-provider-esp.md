# Enable Edge Software Provisioner (ESP) as an OS Provider

This document describes how to configure and enable Edge Software Provisioner (ESP) as an OS provider.

  * [Preparation](#preparation)
  * [Top Config](#top-config)
  * [Custom Config](#custom-config)
  * [Prepare ESP Config](#prepare-esp-config)
  * [Init Edge Conductor Environment](#init-edge-conductor-environment)
  * [Build ESP as OS provider](#build-esp-as-os-provider)
  * [Start and Stop ESP Services](#start-and-stop-esp-services)
  * [Clean up ESP Codebase](#clean-up-esp-codebase)

## Preparation

Read [ESP Document](https://github.com/intel/edge-software-provisioner), especially the [Network Setup](https://github.com/intel/edge-software-provisioner#network-setup), and prepare the hardware and network.

> *NOTE:*  To run ESP on Day-0 machine, 2 CPUs, 20GB HD and 2GB of RAM are needed additionally. See [ESP Prerequisites](https://github.com/intel/edge-software-provisioner#prerequisites).

Follow [Build-and-Install-Edge-Conductor-Tool](../../README.md#build-and-install-edge-conductor-tool) to build and install Edge Conductor tool.
Enter `_workspace` folder to run Edge Conductor tool.

## Edge Conductor Kit Enabling ESP

Some of the Edge Conductor Kits enabled ESP as the os-provider, with these Kits, users can run os-deploy commands.

```shell
kit/
├── capi_byoh.yml
└── rke_ubuntu_20.04.yml
```

A section for `OS` as shown in the following example is added to these Kits to specify the OS provider as ESP.

```yaml
OS:
  manifests:
  - "config/manifests/os_provider_manifest.yml"
  provider: esp
  # Before running "init" with this Kit config file, please update ESP config
  # with correct "git_username" and "git_token" to access the profile git repo.
  config: "config/os-provider/esp_config_profile-ubuntu-20.04.yml" 
```

## Preconditions

 - Users need to setup the ESP network topologic following the settings in [ESP Network Settings](../../configs/extensions/esp_network.yml), and connect all the nodes to be installed to the ESP network. All the nodes should be able to boot from the ESP network (enter Boot Menu and select to boot from PXE via the ESP network).

 - Before running the "init" command, users need to:

     - Input the MAC addresses and static IP addresses of the nodes in the "Parameters - nodes" config section.

 - About offline installation: ESP profile provided by Edge Conductor will download all packages needed by OS provisioning and cluster deployment.

     - If users want to install additional applications, they need to install from official apt mirror via external network.

## Init Edge Conductor Environment

Run the "init" commands to initialize the Edge Conductor environment with the Edge Conductor Kit config with OS provider set as ESP.

For example:

```
./conductor init -c kit/rke_ubuntu_20.04.yml
```

## Build ESP as OS provider

Run the following command to build ESP.

```
./conductor os-deploy build
```

## Start ESP Services

After `./conductor os-deploy build` finishes, the ESP codebase and build result will be under `runtime/esp`.

> *NOTE:*  DO NOT delete the folder `runtime/esp` manually.

> *NOTE:*  DO NOT run `./conductor os-deploy cleanup` when ESP services are still running.

> *NOTE:*  DO NOT run `./conductor deinit` before you have stopped ESP services and cleaned the ESP codebase.

Run the following commands to start ESP services:

```
./conductor os-deploy start
```

## OS Provisioning for Nodes

Boot the nodes from the ESP network and select the OS profile to be installed.

Wait until the OS provisioning finished.


## Stop ESP Services

After all edge nodes are provisioned, run the following command to stop ESP services:

```
./conductor os-deploy stop
```

> *NOTE:*  There might be "Device or resource busy" error after this step when doing cleanup or build ESP for the 2nd time. This is ESP known issue, please refer to [ESP issue #5](https://github.com/intel/Edge-Software-Provisioner/issues/5). To workaround this issue, please manually umount the busy locations, then do `os-deploy cleanup` again.

## Clean up ESP Codebase

Run the following commands to clean up ESP codebase and file system.

```
./conductor os-deploy cleanup
```

After this, you can run the `deinit` command to stop Edge Conductor services.

Copyright (c) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
