/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package kindremover

import (
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	repoutils "github.com/intel/edge-conductor/pkg/eputils/repoutils"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_files := input_files(in)

	log.Infof("Plugin: kind-remover")

	kindbin := filepath.Join(input_ep_params.Runtimebin, "kind")
	if len(input_files.Files) > 0 {
		err := repoutils.PullFileFromRepo(kindbin, input_files.Files[0].Mirrorurl)
		if err != nil {
			return err
		}
	} else {
		err := eputils.GetError("errInvalidFile")
		log.Errorf("No cluster to remove. %s", err)
		return err
	}

	err := os.Chmod(kindbin, 0700)
	if err != nil {
		return err
	}

	cmd := exec.Command(kindbin, "delete", "cluster")
	log.Infof("Removing kind...")
	_, err = eputils.RunCMDEx(cmd, true)
	if err != nil {
		log.Errorf("Failed to remove KIND cluster. %s", err)
		return err
	}

	return nil
}
