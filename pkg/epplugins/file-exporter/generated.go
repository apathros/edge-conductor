/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package fileexporter

import (
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	epplugin "github.com/intel/edge-conductor/pkg/plugin"
)

var (
	Name   = "file-exporter"
	Input  = eputils.NewSchemaMapData()
	Output = eputils.NewSchemaMapData()
)

//nolint:unparam,deadcode,unused
func __name(n string) string {
	return Name + "." + n
}

//nolint:deadcode,unused
func input_exportcontent(in eputils.SchemaMapData) *pluginapi.Filecontent {
	return in[__name("exportcontent")].(*pluginapi.Filecontent)
}

//nolint:deadcode,unused
func input_exportpath(in eputils.SchemaMapData) *pluginapi.Filepath {
	return in[__name("exportpath")].(*pluginapi.Filepath)
}

func init() {
	eputils.AddSchemaStruct(__name("exportcontent"), func() eputils.SchemaStruct { return &pluginapi.Filecontent{} })
	eputils.AddSchemaStruct(__name("exportpath"), func() eputils.SchemaStruct { return &pluginapi.Filepath{} })

	Input[__name("exportcontent")] = &pluginapi.Filecontent{}
	Input[__name("exportpath")] = &pluginapi.Filepath{}

	epplugin.RegisterPlugin(Name, &Input, &Output, PluginMain)
}
