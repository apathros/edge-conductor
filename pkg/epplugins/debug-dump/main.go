/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package debugdump

import (
	// TODO: Add Plugin Imports Here
	eputils "ep/pkg/eputils"
	log "github.com/sirupsen/logrus"
	"time"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_nodes := input_nodes(in)
	input_docker_images := input_docker_images(in)
	input_local_docker_images := input_local_docker_images(in)

	// TODO: Add Plugin Code Here
	log.Infof("dump input_nodes: %v\n", input_nodes)
	eputils.PPrint(input_nodes)
	time.Sleep(3 * time.Second)
	log.Infof("dump input_docker_images: %v\n", input_docker_images)
	eputils.PPrint(input_docker_images)
	log.Infof("dump input_local_docker_images: %v\n", input_local_docker_images)
	eputils.PPrint(input_local_docker_images)

	return nil
}
