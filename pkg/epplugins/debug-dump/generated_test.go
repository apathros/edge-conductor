/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package debugdump

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
)

//nolint:deadcode,unused
func generate_input_nodes(data []byte, in eputils.SchemaMapData) bool {
	inputStruct := &pluginapi.Nodes{}
	if data != nil {
		if err := inputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	in[__name("nodes")] = inputStruct
	return true
}

//nolint:deadcode,unused
func generate_input_docker_images(data []byte, in eputils.SchemaMapData) bool {
	inputStruct := &pluginapi.Images{}
	if data != nil {
		if err := inputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	in[__name("docker-images")] = inputStruct
	return true
}

//nolint:deadcode,unused
func generate_input_local_docker_images(data []byte, in eputils.SchemaMapData) bool {
	inputStruct := &pluginapi.Images{}
	if data != nil {
		if err := inputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	in[__name("local-docker-images")] = inputStruct
	return true
}

//nolint:deadcode,unused,unparam
func generateInput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()
	if result := generate_input_nodes(data["nodes"], n); !result {
		return nil
	}
	if result := generate_input_docker_images(data["docker-images"], n); !result {
		return nil
	}
	if result := generate_input_local_docker_images(data["local-docker-images"], n); !result {
		return nil
	}
	return n
}

//nolint:unparam,deadcode,unused
func generateOutput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()
	return n
}
