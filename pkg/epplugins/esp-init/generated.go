/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package espinit

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	epplugin "ep/pkg/plugin"
)

var (
	Name   = "esp-init"
	Input  = eputils.NewSchemaMapData()
	Output = eputils.NewSchemaMapData()
)

//nolint:unparam,deadcode,unused
func __name(n string) string {
	return Name + "." + n
}

//nolint:deadcode,unused
func input_ep_params(in eputils.SchemaMapData) *pluginapi.EpParams {
	return in[__name("ep-params")].(*pluginapi.EpParams)
}

//nolint:deadcode,unused
func input_os_provider_manifest(in eputils.SchemaMapData) *pluginapi.Osprovidermanifest {
	return in[__name("os-provider-manifest")].(*pluginapi.Osprovidermanifest)
}

func init() {
	eputils.AddSchemaStruct(__name("ep-params"), func() eputils.SchemaStruct { return &pluginapi.EpParams{} })
	eputils.AddSchemaStruct(__name("os-provider-manifest"), func() eputils.SchemaStruct { return &pluginapi.Osprovidermanifest{} })

	Input[__name("ep-params")] = &pluginapi.EpParams{}
	Input[__name("os-provider-manifest")] = &pluginapi.Osprovidermanifest{}

	epplugin.RegisterPlugin(Name, &Input, &Output, PluginMain)
}
