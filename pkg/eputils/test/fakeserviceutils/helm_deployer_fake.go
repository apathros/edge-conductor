/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package fakeserviceutils

import (
	serviceutil "ep/pkg/eputils/service"
)

type FakeHelmDeployer struct {
}

func (h *FakeHelmDeployer) GetName() string {
	return "FakeHelmDeployer"
}

func (h *FakeHelmDeployer) HelmStatus(loc_kubeconfig string) (string, int) {
	return "", 0
}

func (h *FakeHelmDeployer) HelmInstall(loc_kubeconfig string, arg ...serviceutil.InstallOpt) error {
	return nil
}

func (h *FakeHelmDeployer) HelmUpgrade(loc_kubeconfig string) error {
	return nil
}

func (h *FakeHelmDeployer) HelmUninstall(loc_kubeconfig string) error {
	return nil
}
