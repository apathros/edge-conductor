/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package serviceinjector

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
func generate_input_downloadfiles(data []byte, in eputils.SchemaMapData) bool {
	inputStruct := &pluginapi.Files{}
	if data != nil {
		if err := inputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	in[__name("downloadfiles")] = inputStruct
	return true
}

//nolint:deadcode,unused
func generate_input_serviceconfig(data []byte, in eputils.SchemaMapData) bool {
	inputStruct := &pluginapi.Serviceconfig{}
	if data != nil {
		if err := inputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	in[__name("serviceconfig")] = inputStruct
	return true
}

//nolint:deadcode,unused,unparam
func generateInput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()
	if result := generate_input_ep_params(data["ep-params"], n); !result {
		return nil
	}
	if result := generate_input_downloadfiles(data["downloadfiles"], n); !result {
		return nil
	}
	if result := generate_input_serviceconfig(data["serviceconfig"], n); !result {
		return nil
	}
	return n
}

//nolint:deadcode,unused
func generate_output_serviceconfig(data []byte, out eputils.SchemaMapData) bool {
	outputStruct := &pluginapi.Serviceconfig{}
	if data != nil {
		if err := outputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	out[__name("serviceconfig")] = outputStruct
	return true
}

//nolint:unparam,deadcode,unused
func generateOutput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()
	if result := generate_output_serviceconfig(data["serviceconfig"], n); !result {
		return nil
	}
	return n
}
