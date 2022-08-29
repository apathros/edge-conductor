/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

//nolint:deadcode,unused,varcheck
const (
	PROJECTNAME = "Edge Conductor"

	DefaultRegistryPort        = "9000"
	DefaultHttpPort            = "8080"
	DefaultWfPort              = "50088"
	DefaultClusterProvider     = "kind"
	DefaultClusterManifests    = "config/manifests/cluster_provider_manifest.yml"
	DefaultClusterConfig       = "config/cluster-provider/kind_cluster.yml"
	DefaultOSProvider          = "none"
	DefaultOSManifests         = "config/manifests/os_provider_manifest.yml"
	DefaultOSConfig            = ""
	DefaultOSDistro            = "ubuntu2004"
	DefaultComponentsSelector  = "default"
	DefaultComponentsManifests = "config/manifests/component_manifest.yml"
	dirDefaultConfigFile       = "config"
	fnTopConfig                = "kit/kind.yml"
	dirRuntime                 = "runtime"
	fnRuntimeDataDir           = "data"
	fnRuntimeInitParams        = "data/ep-params"
	Epcmdline                  = "cmdline"
	Epkubeconfig               = "kubeconfig"
	WfConfig                   = "workflow/workflow.yml"
	KitConfigPath              = "kit/kind.yml"
	ROOTCACERTFILE             = "cert/pki/ca.pem"
)
