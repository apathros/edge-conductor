/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package kinddeployer

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	epplugin "ep/pkg/plugin"
)

var (
	Name   = "kind-deployer"
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
func input_files(in eputils.SchemaMapData) *pluginapi.Files {
	return in[__name("files")].(*pluginapi.Files)
}

//nolint:deadcode,unused
func input_kind_config(in eputils.SchemaMapData) *pluginapi.Filecontent {
	return in[__name("kind-config")].(*pluginapi.Filecontent)
}

//nolint:deadcode,unused
func output_kubeconfig(outp *eputils.SchemaMapData) *pluginapi.Filecontent {
	return (*outp)[__name("kubeconfig")].(*pluginapi.Filecontent)
}

func init() {
	eputils.AddSchemaStruct(__name("ep-params"), func() eputils.SchemaStruct { return &pluginapi.EpParams{} })
	eputils.AddSchemaStruct(__name("files"), func() eputils.SchemaStruct { return &pluginapi.Files{} })
	eputils.AddSchemaStruct(__name("kind-config"), func() eputils.SchemaStruct { return &pluginapi.Filecontent{} })
	eputils.AddSchemaStruct(__name("kubeconfig"), func() eputils.SchemaStruct { return &pluginapi.Filecontent{} })

	Input[__name("ep-params")] = &pluginapi.EpParams{}
	Input[__name("files")] = &pluginapi.Files{}
	Input[__name("kind-config")] = &pluginapi.Filecontent{}
	Output[__name("kubeconfig")] = &pluginapi.Filecontent{}

	epplugin.RegisterPlugin(Name, &Input, &Output, PluginMain)
}
