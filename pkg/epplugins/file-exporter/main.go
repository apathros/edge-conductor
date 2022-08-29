/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package fileexporter

import (
	eputils "ep/pkg/eputils"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_exportcontent := input_exportcontent(in)
	input_exportpath := input_exportpath(in)

	log.Infof("Plugin: file-exporter")

	path := input_exportpath.Path
	content := input_exportcontent.Content
	if path == "" {
		log.Infoln("export path is empty")
		return nil
	}

	if err := eputils.CreateFolderIfNotExist(filepath.Dir(path)); err != nil {
		log.Errorf("Create folder %s failed", filepath.Dir(path))
		return err
	}

	if err := eputils.WriteStringToFile(content, path); err != nil {
		log.Errorf("Fail to write file %s", path)
		return err
	}

	return nil
}
