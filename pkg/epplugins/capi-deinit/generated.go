/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package capideinit

import (
	eputils "ep/pkg/eputils"
	epplugin "ep/pkg/plugin"
)

var (
	Name   = "capi-deinit"
	Input  = eputils.NewSchemaMapData()
	Output = eputils.NewSchemaMapData()
)

//nolint:unparam,deadcode,unused
func __name(n string) string {
	return Name + "." + n
}

func init() {

	epplugin.RegisterPlugin(Name, &Input, &Output, PluginMain)
}
