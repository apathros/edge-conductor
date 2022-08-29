/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package epplugins

import (
	_ "ep/pkg/epplugins/capi-cluster-deploy"
	_ "ep/pkg/epplugins/capi-deinit"
	_ "ep/pkg/epplugins/capi-host-provision"
	_ "ep/pkg/epplugins/capi-parser"
	_ "ep/pkg/epplugins/capi-provider-launch"
	_ "ep/pkg/epplugins/capi-provision-binary-download"
	_ "ep/pkg/epplugins/debug-dump"
	_ "ep/pkg/epplugins/docker-image-downloader"
	_ "ep/pkg/epplugins/docker-remove"
	_ "ep/pkg/epplugins/docker-run"
	_ "ep/pkg/epplugins/esp-init"
	_ "ep/pkg/epplugins/file-downloader"
	_ "ep/pkg/epplugins/file-exporter"
	_ "ep/pkg/epplugins/kind-deployer"
	_ "ep/pkg/epplugins/kind-parser"
	_ "ep/pkg/epplugins/kind-remover"
	_ "ep/pkg/epplugins/node-join-deploy"
	_ "ep/pkg/epplugins/node-join-prepare"
	_ "ep/pkg/epplugins/pre-service-deploy"
	_ "ep/pkg/epplugins/rke-deployer"
	_ "ep/pkg/epplugins/rke-injector"
	_ "ep/pkg/epplugins/rke-parser"
	_ "ep/pkg/epplugins/service-build"
	_ "ep/pkg/epplugins/service-deployer"
	_ "ep/pkg/epplugins/service-injector"
	_ "ep/pkg/epplugins/service-list"
	_ "ep/pkg/epplugins/service-parser"
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
