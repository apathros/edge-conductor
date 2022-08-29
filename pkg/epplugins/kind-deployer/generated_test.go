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

//nolint:deadcode,unused
func generate_input_kind_config(data []byte, in eputils.SchemaMapData) bool {
	inputStruct := &pluginapi.Filecontent{}
	if data != nil {
		if err := inputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	in[__name("kind-config")] = inputStruct
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
	if result := generate_input_kind_config(data["kind-config"], n); !result {
		return nil
	}
	return n
}

//nolint:deadcode,unused
func generate_output_kubeconfig(data []byte, out eputils.SchemaMapData) bool {
	outputStruct := &pluginapi.Filecontent{}
	if data != nil {
		if err := outputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	out[__name("kubeconfig")] = outputStruct
	return true
}

//nolint:unparam,deadcode,unused
func generateOutput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()
	if result := generate_output_kubeconfig(data["kubeconfig"], n); !result {
		return nil
	}
	return n
}
