/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package dockerremove

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
)

//nolint:deadcode,unused
func generate_input_containers(data []byte, in eputils.SchemaMapData) bool {
	inputStruct := &pluginapi.Containers{}
	if data != nil {
		if err := inputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	in[__name("containers")] = inputStruct
	return true
}

//nolint:deadcode,unused,unparam
func generateInput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()
	if result := generate_input_containers(data["containers"], n); !result {
		return nil
	}
	return n
}

//nolint:unparam,deadcode,unused
func generateOutput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()
	return n
}
