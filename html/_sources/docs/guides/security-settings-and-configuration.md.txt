# Edge Conductor Security Settings and Configurations

This document describes the Edge Conductor security settings and configurations.

## Certificate

Edge-Conductor Tool(aka conductor) provides a set of parameters for users to
configure the certificates for secure communication as follows:

* Workflow Engine - Plugin Communication

Edge-Conductor Tool is a workflow based tool with flexible plugins running as
workflow steps. The workflow engine and plugins communicate via secure gRPC.
The mutual TLS is apply for workflow engine server and it's clients(plugins).
The parameters are required for `conductor init` command:

```yaml
Usage:
  conductor init [flags]

Flags:
      --cacert string            Edge-Conductor root ca cert file (default "cert/pki/ca.pem")
      --cakey string             Edge-Conductor root ca key file, for signing server and client certificates (default "cert/pki/ca-key.pem")
      --clientcert string        Edge-Conductor workflow client certificate file (default "cert/pki/workflow/client.pem")
      --clientkey string         Edge-Conductor workflow client certificate key file (default "cert/pki/workflow/client-key.pem")
      --servercert string        Edge-Conductor workflow server certificate file (default "cert/pki/workflow/server.pem")
      --serverkey string         Edge-Conductor workflow server certificate key file (default "cert/pki/workflow/server-key.pem")
			...
```

All the parameters are file paths that could be a provided certificate file or
intended certificate path. If the certificate and key do not exist, the certmgr
of conductor will generate or issue the certificate bundle for the user.

* Local Registry (Harbor) Certificates

Edge-Conductor Tool sets up a local registry using project Harbor - a CNCF graduated
project for secure registry. Conductor will use or generate the certificates of
the registry due to the existence of the file. The CA bundle of the certificates
are the parameter inputs described in section 'workflow engine-plugin communication'

```yaml
Usage:
  conductor init [flags]

Flags:
      --registrycert string      Edge-Conductor regisrty certificate file (default "cert/pki/registry/registry.pem")
      --registrykey string       Edge-Conductor registry certificate key file (default "cert/pki/registry/registry-key.pem")
			...
```

## Network Settings

Some network settings of Edge-Conductor Tool are configurable in `init` phase.
* Registry port sets the Harbor service port on Day-0 machine.
* Working ip is the access ip of all the services on Day-0 machine, by default it's 
  the default interface ip.
* Workflow engine port defines the port opened to communicate with plugins. 

```yaml
Usage:
  conductor init [flags]

Flags:
      --registry-port string     Edge-Conductor registry port (default 9000)
      --host-ip string           Edge-Conductor Tool working ip (default xxx.xxx.xxx.xxx)
      --wf-port string           Edge-Conductor workflow engine port (default 50088)
      ...
```

## Usernames and Credentials

Edge-Conductor Tool itself does not require or store any username or password.
When a third-party project requires username and password. A user maintained file
is needed with specific schema called custom config.
A custom config requires the username and password for Harbor setup (Supported
in v0.1.0) or an external registry (Not supported in v0.1.0) in `init` phase.

```yaml
Usage:
  conductor init [flags]

Flags:
  -m, --custom-config string     Custom config file
  ...
```

* Custom config schema

```yaml
registry:
  # user/password are required for local registry. Can be optional if externalurl is set.
  user: < For local registry, use default user name 'admin' or specify a new user >
  password: < Password to login to the local registry >
  # externalurl is optional. A local registry will be initialized if it is not specified.
  externalurl: < The external registry url >
  # capath is optional.
  capath: < The 3rd party CA certificate>
```

When `externalurl` is specified, the Edge-Conductor tool will not set up the local registry.
(Not supported in v0.2.0)

## Confidential Content

In Edge-Conductor Tool's concept, each data transferring between plugins are
files with particular schema, which leverages the protobuf for serialization.
The workflow engine ensures the confidential data fetched or generated during
the procedure of a workflow enclosed within permitted components. The workflow
engine enables a parameter `confidential` for the data section to identify the
underlying data is protected by the engine and could not be accessed from a
data file.

E.g. Kubeconfig is a confidential data which is fetched from target cluster.
If the confidential flag is not set, a data file could be found in runtime/data
folder without notification to the user. If the confidential flag is set,
no file will be stored and Kubeconfig is not accessible by conductor users.
In this situation, export-file plugin provides the user an approach to export
any file content by their own.

```yaml
apiVersion: conductor/v1
kind: Workflow
metadata:
  name: conductor-workflow
  namespace: edgeconductor
spec:
  data:
  - name: kubeconfig
    confidential: true
    value: |
      content: |
        {{ printf "%s" .Kubeconfig | readfile | nindent 8 }}
```
## Application  Constraint
In Edge-Conductor Tool's concept, there are some guides which the user shall follow to achieve security requirements in final production.

1. Grafana secure password change
Currently Grafana is not capable of enforcing password policy.
The user shall change the password when the Grafana UI password is changed from its default (which they will be forced to do on initial logic) - a good practice is 10-12 ascii printable characters or longer.

2. SSH key rotation
Currently Edge-Conductor utilizes the user's SSH private key to make remote connections, so the customer shall be responsible for rotation of SSH keys.
It's highly recommended to refer SP 800-57 Part 1, specifically the crypto periods associated with various key lengths for rotation.

Link to current version (rev 5):
[Recommendation for Key Management](https://csrc.nist.gov/publications/detail/sp/800-57-part-1/rev-5/final)
Copyright (C) 2022 Intel Corporation
 
SPDX-License-Identifier: Apache-2.0
