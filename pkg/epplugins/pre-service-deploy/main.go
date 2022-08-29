/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package preservicedeploy

import (
	epplugins "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	"ep/pkg/executor"

	log "github.com/sirupsen/logrus"
)

func enable_sriov_vf(epparams *epplugins.EpParams) error {
	sriovEnabled := "false"
	for _, ext := range epparams.Extensions {
		if ext.Name == "sriov" {
			for _, cfg := range ext.Extension.Extension[0].Config {
				if cfg.Name == "sriov_enabled" {
					sriovEnabled = cfg.Value
					break
				}
			}
		}
	}

	if sriovEnabled == "true" {
		err := executor.Run("config/executor/enable_sriov_vf.yml", epparams, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)

	log.Infof("Plugin: pre-service-deploy")

	err := enable_sriov_vf(input_ep_params)
	if err != nil {
		return err
	}

	return nil
}
