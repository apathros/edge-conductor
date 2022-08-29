/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	"ep/pkg/eputils"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func copyCaRuntimeDataDir(registry string, workspace string, runtimedir string, certpath string) error {
	//move root CA CERT to runtime folder, WA for docker mount bug,
	//please refer to https://nickjanetakis.com/blog/docker-tip-66-fixing-error-response-from-daemon-invalid-mode

	runtimeDataCert := filepath.Join(runtimedir, "cert")

	if eputils.FileExists(runtimeDataCert) {
		err := os.RemoveAll(runtimeDataCert)
		if err != nil {
			log.Errorf("Failed to remove existing runtimeDataCert folder")
			return err
		}
	}

	runtimeDataCertRegistry := filepath.Join(runtimeDataCert, registry)
	if err := MakeDir(runtimeDataCertRegistry); err != nil {
		log.Errorf("Failed to create registry certificate folder")
		return err
	}

	rootCertTgt := filepath.Join(runtimeDataCertRegistry, "ca.crt")
	rootCertSrc := filepath.Join(workspace, certpath)

	_, err := eputils.CopyFile(rootCertTgt, rootCertSrc)
	if err != nil {
		log.Errorf("Failed to copy root CA CERT")
		return err
	}

	return nil
}
