/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package epplugins

import (
	_ "github.com/intel/edge-conductor/pkg/epplugins/capi-cluster-deploy"
	_ "github.com/intel/edge-conductor/pkg/epplugins/capi-deinit"
	_ "github.com/intel/edge-conductor/pkg/epplugins/capi-host-provision"
	_ "github.com/intel/edge-conductor/pkg/epplugins/capi-parser"
	_ "github.com/intel/edge-conductor/pkg/epplugins/capi-provider-launch"
	_ "github.com/intel/edge-conductor/pkg/epplugins/capi-provision-binary-download"
	_ "github.com/intel/edge-conductor/pkg/epplugins/debug-dump"
	_ "github.com/intel/edge-conductor/pkg/epplugins/docker-image-downloader"
	_ "github.com/intel/edge-conductor/pkg/epplugins/docker-remove"
	_ "github.com/intel/edge-conductor/pkg/epplugins/docker-run"
	_ "github.com/intel/edge-conductor/pkg/epplugins/esp-init"
	_ "github.com/intel/edge-conductor/pkg/epplugins/file-downloader"
	_ "github.com/intel/edge-conductor/pkg/epplugins/file-exporter"
	_ "github.com/intel/edge-conductor/pkg/epplugins/kind-deployer"
	_ "github.com/intel/edge-conductor/pkg/epplugins/kind-parser"
	_ "github.com/intel/edge-conductor/pkg/epplugins/kind-remover"
	_ "github.com/intel/edge-conductor/pkg/epplugins/node-join-deploy"
	_ "github.com/intel/edge-conductor/pkg/epplugins/node-join-prepare"
	_ "github.com/intel/edge-conductor/pkg/epplugins/pre-service-deploy"
	_ "github.com/intel/edge-conductor/pkg/epplugins/rke-deployer"
	_ "github.com/intel/edge-conductor/pkg/epplugins/rke-injector"
	_ "github.com/intel/edge-conductor/pkg/epplugins/rke-parser"
	_ "github.com/intel/edge-conductor/pkg/epplugins/service-build"
	_ "github.com/intel/edge-conductor/pkg/epplugins/service-deployer"
	_ "github.com/intel/edge-conductor/pkg/epplugins/service-injector"
	_ "github.com/intel/edge-conductor/pkg/epplugins/service-list"
	_ "github.com/intel/edge-conductor/pkg/epplugins/service-parser"
)

var PluginList []string = []string{
	"docker-run",
	"docker-remove",
	"esp-init",
	"kind-parser",
	"kind-deployer",
	"kind-remover",
	"rke-parser",
	"rke-deployer",
	"rke-injector",
	"capi-parser",
	"capi-provision-binary-download",
	"capi-provider-launch",
	"capi-host-provision",
	"capi-cluster-deploy",
	"capi-deinit",
	"debug-dump",
	"docker-image-downloader",
	"file-downloader",
	"file-exporter",
	"service-parser",
	"service-build",
	"service-injector",
	"pre-service-deploy",
	"service-deployer",
	"service-list",
	"node-join-deploy",
	"node-join-prepare",
}
