/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//go:generate mockgen -destination=./mock/repoutils_mock.go -package=mock -copyright_file=../../../api/schemas/license-header.txt github.com/intel/edge-conductor/pkg/eputils/repoutils RepoUtilsInterface

package repoutils

import (
	"github.com/intel/edge-conductor/pkg/eputils"
	orasutils "github.com/intel/edge-conductor/pkg/eputils/orasutils"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type (
	RepoUtilsInterface interface {
		PushFileToRepo(filepath, subRef, rev string) (string, error)
		PullFileFromRepo(filepath string, targeturl string) error
	}
)

func PushFileToRepo(filepath, subRef, rev string) (string, error) {
	var ref string
	var err error
	if orasutils.OrasCli != nil {
		ref, err = orasutils.OrasCli.OrasPushFile(filepath, subRef, rev)
		if err != nil {
			log.Errorln("Failed to push file", filepath, err)
			return "", err
		}
	} else {
		return "", eputils.GetError("errNoPushClient")
	}
	return ref, nil
}

func PullFileFromRepo(filepath string, targeturl string) error {
	u, err := url.Parse(targeturl)
	if err != nil {
		log.Errorln("Failed to pull file", filepath, err)
		return err
	}
	if u.Scheme == "oci" {
		if orasutils.OrasCli != nil {
			err := orasutils.OrasCli.OrasPullFile(filepath, targeturl)
			if err != nil {
				log.Errorln("Failed to pull file", filepath, err)
				return err
			}
		} else {
			return eputils.GetError("errNoPullClient")
		}
	}
	return nil
}
