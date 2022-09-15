# Edge Conductor Configurations

This document describes the Edge Conductor top config and custom config files.

  * [Edge Conductor Kit Introduction](#edge-conductor-kit-introduction)

&nbsp;

## Edge Conductor Kit Introduction

The `conductor` tool will configure the target cluster and software components using an Edge Conductor Kit config file. The config file includes different sections, which will be used in different stages of the deployment. The configuration of each section will be described in detail in the deployment stages.

* `Use` Config Section: A "common" config file is provided in the `Use` section, the default configuration of an Edge Conductor Kit is defined in this "common" config. The Edge Conductor Kit will use the default configuration and then override with other configurations defined below. If a user specified config file is provided, the user need to make sure there is no conflict between the configs, otherwise the tool will report error and stop execution. The tool will not try to merge the conflicts, in order to avoid an unexpected result.



```yaml
      Use:
      - kit/common.yml
        
```

* `Parameters` Config Section:'Parameters' Config section includes customconfig, global_setting, nodes, etc. 
  - Custom Config
     The section `customconfig` in `Parameters` config section describes custom settings like local registry authorization, ironic (CAPI only), etc. `customconfig` includes information for:

    - Local registry authorization (Mandatory)

      ```yaml
      Parameters:
        customconfig:
          registry:
            # user/password are required for local registry.
            user: < For local registry, use default user name 'admin' or specify a new user >
            password: < Password to login to the local registry >
      ```

      > *NOTE:*  The permission of the Edge Conductor Kit config file should be set to 0600 so that only the user who owns it has read/write permission. To do this run the command "chmod 600 your_experiece_kit.yml".

    - Ironic configurations (Only mandatory when the cluster type is set to **clusterapi**)

      ```yaml
      Parameters:
        customconfig:
          ironic:
            kubeconfigpath: < point to the kubeconfig file for accessing the Kubernetes cluster provided by Edge Conductor users >
            provisioninginterface: < the interface name of the NIC connected to the provisioning network >
            provisioningip: < the IP address configured on the provisioning NIC >
            dhcprange: < IP allocation range for the provisioning network >
            httpport: < the port of Ironic service >
            ironicuser: < username for Ironic service >
            ironicpassword: < password for ironicuser >
            ironicinspectoruser: < username for Ironic inspector service >
            ironicinspectorpassword: < password for ironicinspectoruser >
            mariadbpassword: < password for mariadb >
      ```
  - Global Settings
      ```yaml
      Parameters:
        global_settings:
          registry_port: < Service port of the local registry >
          provider_ip: < Service IP for the Providers >
      ```
      > *NOTE:*  Some environments require network proxies for Docker operations (e.g. docker pull, docker push, docker run, and so on). You must ensure these proxies are set correctly prior to using the tool. Note that the Host.server need to be added to no_proxy/NO_PROXY list for the docker proxies.
  - nodes
    ```yaml
    Parameters:
      nodes:
        role: < the role of node;controlplane, worker, etcd are supported >
        critype: < certification type. ACPI only >
        bmc_user: < BMC setting of user name. Metal3(CAPI) only >
        bmc_password: < BMC setting of pass word. Metal3(CAPI) only >
        bmc_protocol: < BMC setting of porotocl. Metal3(CAPI) only >
        bmc_endpoint: < BMC setting of endpoint. Metal3(CAPI) only >
        # The following are for all remote nodes.
        user: < user name >
        ip: < node's ip address >
        mac: < node's mac address >
        ssh_key_path: < node's ssh private key path >
        ssh_key: < instead of setting the path, alternative way to specify the node's ssh key >
        ssh_passwd: < node's ssh password >
        ssh_port: < node's ssh port >
    ```

* `OS` Config Section:
This section specifies OS provider and profiles for bare metal OS deployment.
The supported OS provider is Edge Software Provisioner ([ESP](https://github.com/intel/edge-software-provisioner)).
```yaml
OS:
  manifests: < A list of manifest files describing binary files, docker images and other resources needed by the OS providers. >
  provider: < OS provider type, currently "esp" is supported. Default setting is "none". >
  config: < Config file for the specified OS provider. >
```

* `Cluster` Config Section:
This section specifies a sub-config file to describe the customer cluster. Sample config files are pre-installed and located in the `examples/cluster` folder.

```yaml
Cluster:
  manifests: < A list of manifest files describing binary files, docker images and other resources needed by the Cluster providers. >
  provider: < Type of the cluster provider, can be "kind", or "rke", or "tanzu", or other supported cluster types. Default value is 'kind'. >
  config: < Detailed config file for the specified cluster type. >
```

* `Components` Config Section:
This section specifies a list of components to be deployed on the customer cluster. The tool will filter the component list with a selector to determine which components are deployed to the target cluster.

```yaml
Components:
  manifests: < A list of manifest files describing yaml/helm files, binary files, docker images and other resources needed by the components. >
  selector: < A list of selected configurations, operators and services to be deployed on the cluster. >
  - name: < Selected service name >
    override: < Optional: it is used to override the predefined configurations (in the manifest file) of this service. >
      url: <this will override default selected service url>
      type: <this will override default selected service type, one of "helm, yaml,repo or dce">
      images: <this will override default images selected service used >
        - <image 1>
        - <image ...>
        - <image n>
      supported-clusters: <this will override default images selected service supported cluster>
      - <cluster name should be same as Kit cluster config>
      namespace: <this will override default selected service namespace>
      chartoverride: <this will replace default selected service override file>
```

Following is a complete example of the Edge Conductor Kit config: [Edge Conductor Kit Example for KIND](../../kit/kind.yml)

&nbsp;

Copyright (c) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
