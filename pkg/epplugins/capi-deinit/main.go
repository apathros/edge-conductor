/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package capideinit

import (
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	docker "github.com/intel/edge-conductor/pkg/eputils/docker"
	log "github.com/sirupsen/logrus"
)

var (
	containers = []string{
		"httpd-infra",
		"httpd",
		"ipa-downloader",
		"dnsmasq",
		"ironic",
		"ironic-endpoint-keepalived",
		"ironic-log-watch",
		"ironic-inspector",
	}
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {

	for _, c := range containers {
		if err := docker.RemoveContainer(c); err != nil {
			log.Errorln("Failed to remove", c, err)
			return err
		}
	}

	return nil
}
