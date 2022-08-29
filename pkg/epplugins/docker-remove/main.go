/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package dockerremove

import (
	log "github.com/sirupsen/logrus"

	eputils "ep/pkg/eputils"
	docker "ep/pkg/eputils/docker"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_containers := input_containers(in)

	log.Infoln("Removing containers.")
	for _, c := range input_containers.Containers {
		if err := docker.DockerRemove(c); err != nil {
			log.Errorln(err)
			return err
		}
	}

	return nil
}
