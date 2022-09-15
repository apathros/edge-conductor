[Edge Conductor]: https://github.com/intel/edge-conductor
[Get Started]: ./get-started.md
[EC Configuration]: ./ec-configurations.md

# Config and Deploy Components

This guide describes how to select components and deploy on clusters with [Edge Conductor] tool.

## Preparation

Follow any of the [Cluster Deployment Guides](./index.md#cluster-deployment) to deploy a cluster.

Once a cluster has been created (or re-used as per the existing cluster deployment model), the Edge Conductor tool can be used to deploy components by performing the following 3 steps:
- Download Helm charts, yaml and other files and upload to a local HTTP server
- Download needed container images and push to local container registry
- Deploy services on the target cluster

For example, to deploy components:
1. Wrap it up as the helm chart tarball or a .yml file
1. Place the helm chart tarball or .yml file in a public accessible web server or directly in a Day-0 host local folder
1. Add the URL (https://... or file://...) into the component_manifest.yml file.  

In this guide we will take [KIND Cluster](./cluster-deploy-KIND.md) as an example to config and deploy the components.




## How to Deploy Default Components
Enter the workspace folder which is used to deploy the cluster.

Edge Conductor tool defined a list of components that have been tested on the clusters, in [component_manifest.yml](../../configs/manifests/component_manifest.yml).

Edge Conductor tool also defined a list of default components selected for each type of clusters. This list is defined as a Component Selector list, which can be found in the Edge Conductor kit file. It is made up by
- A list of common components in [kit/common.yml](../../kit/common.yml)
- And a list of components specific for the cluster in cluster kit configs. Take KIND cluster as an example, you can find the list here in [kit/kind.yml](../../kit/kind.yml#L40).

To build and deploy the default components defined for the cluster, enter the commands:

```bash
./conductor service build
./conductor service deploy
```

To check what components have been deployed on the cluster, enter this command:

```bash
./conductor service list
```


## How to Add/Remove/Modify Components from Current Cluster

The Component Selector list is defined in a declarative way.

Users can add/remove/modify the components in the selector list, like in the following example of a selector list in a kit config yaml.

In this example, 
- Users can add their own component_manifest.yml to define customized components. For more details of how to add a component manifest, please refer to the tutorial [Hello Cluster!](../tutorials/samples/hello-cluster.md).
- All components defined in these component manifests can be selected by name.
- A selected component can be override in the selector list via "override" field.
- Users can add/remove/modify the selector list, and then run "conductor service build/deploy" again, no need to redo the "conductor init" or cluster deployment.

Example of a selector list in a kit config yaml file:
```yaml
...
Components:
  manifests:
  - "config/manifests/component_manifest.yml"
  # newly added manifest
  - "my/own/component_manifest.yml"
  selector:
  - name: nginx-ingress
    override:
      # modified component
      url: file:///my/own/kind-nginx-ingress.yml
      type: yaml
      images:
        - k8s.gcr.io/ingress-nginx/controller:v1.1.2
        - k8s.gcr.io/ingress-nginx/kube-webhook-certgen:v1.1.1
      supported-clusters:
      - kind
      namespace: ingress-nginx

  # deleted
  # - name: portainer-ce

  # newly added (defined in "my/own/component_manifest.yml")
  - name: my-important-service
```

To apply these changes in the selector list, enter the commands again:

```bash
./conductor service build
./conductor service deploy
```

During this service build/deploy operations,
- The deleted components will be uninstalled from the cluster.
- The newly added components will be installed to the cluster.
- The modified components will be upgraded in the cluster.


## How to Handle the Dependency of Multiple Components

All the components in the selector list will be applied to clusters one-by-one, it's following the order defined in the list.

If there's hard dependency on a component, use "wait" field so Edge Conductor tool will hold on to wait until this component is successfully running on the cluster, with a timeout limitation.

This "wait" mechanism can be defined in the component manifest, or it can also be specified in the kit config selector list. See the following examples.

Example of a component manifest:
```yaml
Components:
  - name: my-important-service
    url: https://my/service/helm.tgz
    type: helm
    supported-clusters:
    - kind
    # Add a wait for 300 seconds
    wait:
      timeout: 300
```

Example of a selector list in a kit config yaml file:
```yaml
...
Components:
  manifests:
  - "config/manifests/component_manifest.yml"
  - "my/own/component_manifest.yml"
  selector:
  - name: my-important-service
    override:
      # Override the wait timeout to 500 seconds
      wait:
        timeout: 500
```

During the service build/deploy operations, the tool will wait until "my-important-service" is successfully deployed, or it comes to a timeout.


## What Types of Components Can Be Supported

Edge Conductor tool can deploy the following types of components:
- yaml: Apply to the cluster with a yaml file. See detailed descriptions below.
- helm: Apply to the cluster with a helm charts. See detailed descriptions below.
- dce: Apply to the cluster with DCE specs. See descriptions below. More details refer to DCE guide.

Users can define the components in a component manifest file, like [component_manifest.yml](../../configs/manifests/component_manifest.yml). Following are details of the different types of the components.

Detailed description of yaml type:
```yaml
Components:
  - name: name-of-yaml-component
    type: yaml
    url: <url of the yaml file for this component, http|https|file are supported>
    namespace: <optional, the namespace to apply the component>
    images: <optional, upstream images used by the yaml file, will be downloaded at "service build">
      - ...
    executor: <optional, object of dce executor, which is to help users provide some build operations before applied>
      build: <optional, dce spec to run at "service build">
    resources: <optional, a list of name-value pairs for customized information, can be used by dce executor>
      - name: <name>
        value: <value>
    supported-clusters: <a list of clusters that this component can be successfully applied, like kind|rke|capi>
      - ...
```

Detailed description of helm type:
```yaml
Components:
  - name: name-of-helm-component
    type: helm
    # Use <url> or <helmrepo> + <chartname> + <chartversion> to get the helm charts
    url: <url of the helm file for this component, http|https|file are supported.>
    helmrepo: <url of a helm repo storing the helm charts>
    chartname: <name of the helm charts>
    chartversion: <version of the helm charts>
    chartoverride: <optional, url of the file to override the helm values>
    namespace: <optional, the namespace to apply the component>
    images: <optional, upstream images used by the helm file, will be downloaded at "service build">
      - ...
    executor: <optional, object of dce executor, which is to help users provide some build operations before applied>
      build: <optional, dce spec to run at "service build">
    resources: <optional, a list of name-value pairs for customized information, can be used by dce executor>
      - name: <name>
        value: <value>
    supported-clusters: <a list of clusters that this component can be successfully applied, like kind|rke|capi>
      - ...
```

Detailed description of dce type:
```yaml
Components:
  - name: name-of-dce-component
    type: dce
    url: <optional, url of the source code or resource used by dce executor, http|https|file are supported>
    namespace: <optional, the namespace to apply the component>
    executor: <object of dce executor>
      build: <dce spec to run at "service build">
      # <executor>.<deploy> is only supported in dce components. An error will pop-up if this field is defined in other types.
      deploy: <dce spec to run at "service deploy", only supported for dce components>
    resources: <optional, a list of name-value pairs for customized information, can be used by dce executor>
      - name: <name>
        value: <value>
    supported-clusters: <a list of clusters that this component can be successfully applied, like kind|rke|capi>
      - ...
```



Copyright (c) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
