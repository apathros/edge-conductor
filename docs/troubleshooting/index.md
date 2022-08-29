
[Edge Conductor]: https://github.com/intel/edge-conductor
[Troubleshooting]: ./index.md
[Edge Conductor] / [Troubleshooting]
# Edge Conductor Troubleshooting
This document provides more detailed information about the error code you have encountered during the build/deploy cluster or service on Edge Conductor.

##  E000: Unknown error
* E000.099: Unknown error
* E000.100: test error
##  E001: EC tools errors

// E001.0**: common config errors
* [E001.001](./e001.001.md): kitconfig section is not found in runtime file
* [E001.002](./e001.002.md): kitconfigpath section is not found in runtime file
* [E001.003](./e001.003.md): Edge Conductor kit runtime file is not found 
* E001.004: Edge Conductor kit customconfig is not found
* E001.005: Harbor Register password is not set
* E001.006: Cert manager config is missing in manifest
* E001.007: Malformat Yaml file
* E001.008: Can not find docker config folder
* E001.009: the 'name' section under components. selector is missing
* E001.010: component.selector section has error
* E001.011: component.selector.override section has error
* E001.012: 'unmarshal' operation failed of the component.selector.override section
* E001.013: The length of Args exceeds the MAX COUNT(100)
* E001.014: Edge Conductor Kit Config's parameter is not correct
* E001.015: Edge Conductor Kit Config's parameter missing
* E001.016: Invalid parameter
* E001.017: CopyFromDay0: must end with "/"
* E001.018: CopyToDay0: must end with "/"
* E001.019: cluster manifest is not found
* E001.020: StringOverrideWithNode
* E001.021: resource is not specified in cluster manifest
* E001.022: grpc connect error
* E001.023: workflow is not found
* E001.024: Plugin Complete error
* E001.025: Failed to get ep-params in init data
* E001.026: This command is not supported for current config
* E001.027: Unmarshal data error
* E001.028: Unmarshal plugin data error
* E001.029: Unknown return value
* E001.030: Failed to get schema in output
* E001.031: Cannot find plugin
* E001.032: Failed to find input schema
* E001.033: Cannot find input schema
* E001.034: Get log stream error
* E001.035: Plugin connection error
* E001.036: Failed to convert struct to map
* E001.037: Failed to convert map to service
* E001.038: Unmarshal override error
* E001.039: failed to load ep-params data from init data
* E001.040: failed to load plugin data from init data
* E001.041: marshal schemaMap data error
* E001.042: Failed to get schema in previous step's init or output data
* E001.043: convert containers data error
* E001.044: marshal initdata failed
* E001.045: Node login username missing
* E001.046: image is not specified in cluster manifest
* E001.047: Node login password and ssh key missing
* E001.048: Ignore override error
* E001.049: Ignore format error
* E001.050: Unknown command type
* E001.051: binary is not specified in cluster manifest

// E001.1**: kind cluster errors
* E001.101: Failed to create KIND cluster
* E001.102: Kind image info missing in manifest.
* E001.103: Failed to delete KIND cluster
* E001.104: Kind info missing, pls check manifest

// E001.2**: rke cluster errors
* E001.202: Failed to run rke command
* E001.203: Could not config viper.

// E001.3**: CAPI Error
* E001.301: Cert manager config is missing in manifest
* E001.302: CAPI provider parameter missing in manifest
* E001.303: require one infra provider, pls select one infra provider at kit config
* E001.304: CAPI providers missing in manifest, four providers are required 
* E001.305: invalid worker node number pls check kit config
* E001.306: CAPI manifest missing
* E001.307: validated CAPI infra provider missing in manifest
* E001.308: kind config missing in manifest
* E001.309: capi download error
* E001.310: error in getting clients
* E001.311: internal server error
* E001.312: Failed to find binary of kind in capi binary list
* E001.320: failed to generate capi binary list
* E001.321: failed to get capi clusterctl configuration
* E001.322: failed to generate local provider repo for clusterctl
* E001.323: failed to generate config files for clusterctl
* E001.324: failed to launch management cluster
* E001.325: failed to init clusterapi for define target cluster
* E001.326: failed to start/restart management cluster of CAPI
* E001.327: Failed to run clusterctl cmd
* E001.328: Failed to launch byoh controller manager
* E001.329: BYOH host not ready
* E001.330: no controller plane node ready
* E001.331: Failed to get management cluster binary list

// E001.4**: Service errors
* E001.401: service's tls extension of  is not found
* E001.402: pullFile: invalid command
* E001.403: service's tls config is not found
* E001.404: Service TLS error: CSR filename not found
* E001.405: pullFile: only supported on day-0
* E001.406: pushFile: invalid command
* E001.407: pullFile: only supported on day-0
* E001.408: Helm repo or chart name is empty
* E001.409: Helm service is in a wrong status
* E001.410: Helm service status is unknown
* E001.411: failed to return schemaMap data
* E001.412: File not found in download list.
* E001.413: Server address or port is missing in kitconfig
##  E002: Network errors
* E002.002:  provide_ip under global setings is not set
* E002.003: SSH path for provision is not found.
* E002.004: remote copy failed with invalid file path
##  E003: Kubernetes
* E003.001: No k8s node
##  E004: Security errors
* E004.001: unsupported certificate type
* E004.002: failed to decode Cert
* E004.003: certbundle.Ca.Cert is a directory
* E004.004: certbundle.Ca.Key is a directory
* E004.005: certbundle.Server.Cert is a directory
* E004.006: certbundle.Server.Key is a directory
* E004.007: certcertbundle.Client.Cert is a directory
* E004.008: certbundle.Client.Key is a directory
* E004.009: cert path or Key path is nil
* E004.010: unsupported key algo
* E004.011: failed to parse root certificate
##  E005: Utility errors

// E005.0**: Docker errors
* E005.001: no auth in docker client configuration file
* E005.002: registry server or port is not found
* E005.003: invalid docker config path，check the DOCKER
* E005.004: failed to open docker client config file
* E005.005: container name is not found
* E005.006: pushImage only supported on day-0
* E005.007: Pulling file failure.
* E005.008: docker start failed
* E005.009: host ip as "0.0.0.0"  in docker config  is not accepted
* E005.010: failed to parse docker config file
* E005.011: failed to decode authentication information in config file
* E005.012: auth string in docker config file is invalid
* E005.013: docker config file is not existed
* E005.014: File path should not be empty
* E005.015: Oras default resolver not found
* E005.016: Oras resolver not found
* E005.017: It's an oras error

// E005.1**: Harbor errors
* E005.101: input harbor IP is empty
* E005.102: input harbor user is empty
* E005.103: input harbor password is empty
* E005.104: input auth server address is empty
* E005.105: input harbor port is empty
* E005.106: input harbor url is empty
* E005.107: error in harbor response 
* E005.108: Abnormal harbor response received
* E005.109: Cert file is null
* E005.110: Harbor project name is empty
* E005.111: Harbor auth string is empty

// E005.2**: File utility errors
* E005.201: file is not valid
* E005.202: file is not valid tgz
* E005.203: unmarshal operation failed
* E005.204: marshal operation failed
* E005.205: tar operation failed with invalid tar file path
* E005.206: uncompress tgz failed with invalid file header
* E005.207: create file failed, the file is existed
* E005.208: Failed to load json file
* E005.209: Files for cluster deployment are not found. Please run "cluster build" first
* E005.210: URL Schema Not Supported
* E005.211: no such file or directory

// E005.3**: Hash errors
* E005.301: SHA256 check failed
* E005.302: Hash check failed

// E005.4**: Repo utility errors
* E005.401: push to repo failed
* E005.402: pull from repo failed

// E005.5**: ESP errors
* E005.501: Cannot find OS session in top config
* E005.502: ESP manifest missing
* E005.503: Cannot find ESP codebase
* E005.504: Cannot find ESP build.sh script
* E005.505: Cannot find ESP run.sh script
* E005.506: OS provider is wrong

Copyright (C) 2022 Intel Corporation
SPDX-License-Identifier: Apache-2.0
