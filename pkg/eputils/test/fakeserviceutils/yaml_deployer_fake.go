/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package fakeserviceutils

type FakeYamlDeployer struct {
}

func (y *FakeYamlDeployer) GetName() string {
	return "FakeYamlDeployer"
}

func (y *FakeYamlDeployer) YamlInstall(loc_kubeconfig string) error {
	return nil
}

func (y *FakeYamlDeployer) YamlUninstall(loc_kubeconfig string) error {
	return nil
}
