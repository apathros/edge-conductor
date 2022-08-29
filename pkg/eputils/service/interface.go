/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//go:generate mockgen -destination=./mock/service_mock.go -package=mock -copyright_file=../../../api/schemas/license-header.txt ep/pkg/eputils/service ServiceDeployer,HelmDeployerWrapper,YamlDeployerWrapper,ServiceTLSExtensionWrapper

package service

import (
	epplugins "ep/pkg/api/plugins"
)

type ServiceDeployer interface {
	NewHelmDeployer(name, namespace, charts, values string) HelmDeployerWrapper
	NewYamlDeployer(name, namespace, yamlfile string, para ...interface{}) YamlDeployerWrapper
}

type HelmDeployerWrapper interface {
	GetName() string
	HelmStatus(loc_kubeconfig string) (string, int)
	HelmInstall(loc_kubeconfig string, arg ...InstallOpt) error
	HelmUpgrade(loc_kubeconfig string) error
	HelmUninstall(loc_kubeconfig string) error
}

type YamlDeployerWrapper interface {
	GetName() string
	YamlInstall(loc_kubeconfig string) error
	YamlUninstall(loc_kubeconfig string) error
}

type ServiceTLSExtensionWrapper interface {
	GenSvcTLSCertFromTLSExtension(exts []*epplugins.EpParamsExtensionsItems0, tgtSvc string) error
	GenSvcSecretFromTLSExtension(exts []*epplugins.EpParamsExtensionsItems0, tgtSvc, ns, kubeconfig string) error
}
