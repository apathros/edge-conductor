/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package kindremover

import (
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
)

//nolint:deadcode,unused
func generate_input_ep_params(data []byte, in eputils.SchemaMapData) bool {
	inputStruct := &pluginapi.EpParams{}
	if data != nil {
		if err := inputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	in[__name("ep-params")] = inputStruct
	return true
}

//nolint:deadcode,unused
func generate_input_files(data []byte, in eputils.SchemaMapData) bool {
	inputStruct := &pluginapi.Files{}
	if data != nil {
		if err := inputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	in[__name("files")] = inputStruct
	return true
}

//nolint:deadcode,unused,unparam
func generateInput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()
	if result := generate_input_ep_params(data["ep-params"], n); !result {
		return nil
	}
	if result := generate_input_files(data["files"], n); !result {
		return nil
	}
	return n
}

//nolint:unparam,deadcode,unused
func generateOutput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()
	return n
}
