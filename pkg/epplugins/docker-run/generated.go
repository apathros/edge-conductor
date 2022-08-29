/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package dockerrun

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	epplugin "ep/pkg/plugin"
)

var (
	Name   = "docker-run"
	Input  = eputils.NewSchemaMapData()
	Output = eputils.NewSchemaMapData()
)

//nolint:unparam,deadcode,unused
func __name(n string) string {
	return Name + "." + n
}

//nolint:deadcode,unused
func input_containers(in eputils.SchemaMapData) *pluginapi.Containers {
	return in[__name("containers")].(*pluginapi.Containers)
}

func init() {
	eputils.AddSchemaStruct(__name("containers"), func() eputils.SchemaStruct { return &pluginapi.Containers{} })

	Input[__name("containers")] = &pluginapi.Containers{}

	epplugin.RegisterPlugin(Name, &Input, &Output, PluginMain)
}
