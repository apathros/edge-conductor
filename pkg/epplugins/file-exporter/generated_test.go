/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package fileexporter

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
)

//nolint:deadcode,unused
func generate_input_exportcontent(data []byte, in eputils.SchemaMapData) bool {
	inputStruct := &pluginapi.Filecontent{}
	if data != nil {
		if err := inputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	in[__name("exportcontent")] = inputStruct
	return true
}

//nolint:deadcode,unused
func generate_input_exportpath(data []byte, in eputils.SchemaMapData) bool {
	inputStruct := &pluginapi.Filepath{}
	if data != nil {
		if err := inputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	in[__name("exportpath")] = inputStruct
	return true
}

//nolint:deadcode,unused,unparam
func generateInput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()
	if result := generate_input_exportcontent(data["exportcontent"], n); !result {
		return nil
	}
	if result := generate_input_exportpath(data["exportpath"], n); !result {
		return nil
	}
	return n
}

//nolint:unparam,deadcode,unused
func generateOutput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()
	return n
}
