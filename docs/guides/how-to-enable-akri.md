# How to enable Akri with Edge conductor
Akri is a Kubernetes Resource Interface that lets you easily expose heterogeneous leaf devices (such as IP cameras and USB devices) as resources in a Kubernetes cluster, while also supporting the exposure of embedded hardware resources such as GPUs and FPGAs. Akri continually detects nodes that have access to these devices and schedules workloads based on them.

## Contents

  * [Preparation](#preparation)
  * [Basic configuration for Akri](#basic-configuration-for-Akri)
  * [Testing](#Testing)
  * [Limitations](#Limitations)

## Preparation

Follow [HW Requirements for Edge Conductor Day-0 Host](../../README.md#hw-requirements-for-edge-conductor-day-0-host) and [OS and System Requirements for Edge Conductor Day-0 Host](../../README.md#os-and-system-requirements-for-edge-conductor-day-0-host) to prepare the Day-0 host hardware and software.

Follow [Build and Install Edge Conductor Tool](../../README.md#build-and-install-edge-conductor-tool) to build and install Edge Conductor tool.
Enter `_workspace` folder to run Edge Conductor tool.



## Basic configuration for Akri

Akri service is supported in RKE and CAPI clusters currently.
We use Edge Conductor Kit to deploy them with Akri enabled in this document.

Akri will be enabled by modifying below part in [common.yml](../../kit/common.yml).


```yaml
Components:
  manifests:
  - "config/manifests/component_manifest.yml"
  selector:
  - name: akri
```
The discovery configuration is in [akri-override.yml](../../configs/service-overrides/akri/akri-override.yml).

If you want to discover udev camera with name "video0", using the following configuration in udevRules.

```yaml
udev:
  configuration:
    enabled: true
    name: akri-udev
    discoveryDetails:
      udevRules:
      -  'KERNEL=="video0"'
```

If you want to discover any ONVIF IP camera with IP address exclude "10.0.0.1","10.0.0.2", using the following configuration in ipAddresses.

```yaml
onvif:
  configuration:
    enabled: true
    name: akri-onvif
    discoveryDetails:
      ipAddresses:
        action: Exclude
        items: ["10.0.0.1","10.0.0.2"]
```

Both of these configurations can take effect together.

After finishing the above configuration, we can start deployment of cluster following [Deploy a Target Cluster](../../README.md#edge-conductor-documentation).

## Testing
Using Akri to discover mock USB cameras attached to nodes in a Kubernetes cluster. 
Here using RKE deployment as an example.

Precondition:
1. Make sure that, the worker node in the cluster is virtual machine or physical machine with security boot disabled.
2. Make sure that, the cpu core number of the worker node is at least 4.

Testing steps:
- Install v4l2loopback kernel module and its prerequisites in worker node.

  Option 1: by deb package
  ```shell
  sudo apt update
  sudo apt -y install linux-modules-extra-azure
  sudo apt update
  sudo apt -y install linux-headers-$(uname -r)
  sudo apt -y install linux-modules-extra-$(uname -r)
  sudo apt -y install dkms
  curl http://deb.debian.org/debian/pool/main/v/v4l2loopback/v4l2loopback-dkms_0.12.5-1_all.deb -o   v4l2loopback-dkms_0.12.5-1_all.deb
  sudo dpkg -i v4l2loopback-dkms_0.12.5-1_all.deb
  ```
  Option 2: by git repo
  If error happened during install v4l2loopback-dkms_0.12.5-1_all.deb with dpkg, we can clone the repo, build the module, and setup the module dependencies like so:

  ```shell
  git clone https://github.com/umlaeute/v4l2loopback.git
  cd v4l2loopback
  make & sudo make install
  sudo make install-utils
  sudo depmod -a
  ```
- Setting up mock udev video devices in worker node.

  ```shell
  sudo modprobe v4l2loopback exclusive_caps=1 video_nr=0
  ```

- Confirm the video device (video0) has been created.

  ```shell
  ls /dev/video*
  /dev/video0
  ```

- Pass fake video streams through Gstreamer.

  ```shell
  sudo apt-get install -y \
       libgstreamer1.0-0 gstreamer1.0-tools gstreamer1.0-plugins-base \
       gstreamer1.0-plugins-good gstreamer1.0-libav
  mkdir camera-logs
  sudo gst-launch-1.0 -v videotestsrc pattern=ball ! "video/x-raw,width=640,height=480,framerate=10/1" ! avenc_mjpeg ! v4l2sink device=/dev/video0 > camera-logs/ball.log 2>&1 &
  ```

- Build and Deploy the RKE cluster.

  ```shell
  ./conductor cluster build
  ./conductor cluster deploy
  ```

- Build and deploy the services using following commands.

  ```shell
  ./conductor service build
  ./conductor service deploy
  ```

- Once the services are deployed, display the pods using the following command and make sure that akri agent, controller and relevant discovery pod are running.

  ```shell
  ~$ kubectl get po -n akri-component
  NAME                                        READY   STATUS    RESTARTS  AGE
  akri-agent-daemonset-jrtfh                  1/1     Running   0         12m
  akri-controller-deployment-54584d9474-2nghj 1/1     Running   0         12m
  akri-onvif-discovery-daemonset-dsvjm        1/1     Running   0         12m
  akri-udev-discovery-daemonset-nb4hl         1/1     Running   0         12m
  akri-udev-290b9e-pod                        1/1     Running   0         12m
  ```
For other details, please refer to [Akri user guide](https://docs.akri.sh/user-guide/getting-started).

## Limitations

Currently, ONVIF camera that requires authentication is not supported in Akri upstream now.
[Support ONVIF cameras that require authentication](https://github.com/project-akri/akri/issues/250)

Copyright (C) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
