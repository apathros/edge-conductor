/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package servicebuild

import (
	eputils "ep/pkg/eputils"
	svcutils "ep/pkg/eputils/service"
	"ep/pkg/executor"
	log "github.com/sirupsen/logrus"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_serviceconfig := input_serviceconfig(in)

	for _, service := range input_serviceconfig.Components {
		// Prepare certificates for services
		err := svcutils.GenSvcTLSCertFromTLSExtension(input_ep_params.Extensions, service.Name)
		if err != nil {
			return err
		}

		if service.Executor != nil {
			if service.Executor.Build != "" {
				log.Debugf("service %v, build spec: %v\n", service.Name, service)
				err := executor.Run(service.Executor.Build, input_ep_params, service)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
