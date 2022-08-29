/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package dockerrun

import (
	eputils "ep/pkg/eputils"
	docker "ep/pkg/eputils/docker"
	log "github.com/sirupsen/logrus"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_containers := input_containers(in)

	log.Infoln("Start to run containers.")
	for _, c := range input_containers.Containers {
		if err := docker.DockerRun(c); err != nil {
			log.Errorln(err)
			return err
		}
	}

	return nil
}
