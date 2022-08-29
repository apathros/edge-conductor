/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package restfulcli

type GoharborClientWrapper interface {
	TlsBasicAuth(username, password string) string
	RegistryCreateProject(harborUrl, project, authStr, certFilePath string) error
	RegistryProjectExists(harborUrl, project, authStr, certFilePath string) (bool, error)
	MapImageURLCreateHarborProject(harborIP string, harborPort string, harborUser string, harborPass string, image []string) ([]string, error)
}
