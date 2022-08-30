/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package rkeparser

import (
	"fmt"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var cluster_manifest = []byte(`{
	"cluster_providers":[
		{
			"name":"kind",
			"images":[
				{"name":"img_node","repo_tag":""},
				{"name":"img_haproxy","repo_tag":""}
				],
			"binaries":[{"name":"kindtool","url":"","sha256":""}]
		},
		{
			"name":"rke",
			"binaries":[{"name":"rketool","url":"","sha256":""}]
		}
	]}`)

func TestPluginMain(t *testing.T) {
	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           bool
		expectErrorMsg        string
	}{
		{
			name: "RKE_manifest_lost",
			input: map[string][]byte{
				"ep-params": []byte(`{
					"kitconfig": {
						"Cluster": {
							"provider": "rke",
							"config": "../../../configs/cluster-provider/rke_cluster.yml"}}}`),
			},

			expectedOutput: nil,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errManifest").Error(),
		},
		{
			name: "RKE_config_file_load_fail",
			input: map[string][]byte{
				"ep-params": []byte(`{
					"kitconfig": {
						"Cluster": {
							"provider": "rke",
							"config": "../../../examples/cluster/rke_cluster_fake.yml"}}}`),
				"cluster-manifest": cluster_manifest,
			},

			expectedOutput: nil,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errLoadJson").Error(),
		},
		{
			name: "RKE_viper_read_fail",
			input: map[string][]byte{
				"ep-params": []byte(`{
					"kitconfig": {
						"Cluster": {"provider": "rke", "config": "testfile"}}}`),
				"cluster-manifest": cluster_manifest,
			},

			expectedOutput: nil,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errConfViper").Error(),
		},
		{
			name: "RKE_parse_success",
			input: map[string][]byte{
				"ep-params":        []byte(`{"kitconfig": {"Cluster": {"provider": "rke", "config": "../../../configs/cluster-provider/rke_cluster.yml"}}}`),
				"cluster-manifest": cluster_manifest,
			},

			expectedOutput: map[string][]byte{
				"docker-images": []byte(`{"images": []}`),
				"files":         []byte(`{"files":[{"url": "", "hash":"", "hashtype":"sha256", "mirrorurl": "", "urlreplacement": {"origin": "://.", "new": "binary"}}]}`),
			},
			expectError:    false,
			expectErrorMsg: "",
		},
	}

	err := eputils.WriteStringToFile("test", "testfile")
	require.NoError(t, err, "Write String To File Error:")

	defer os.RemoveAll("testfile")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			testOutput := generateOutput(nil)

			expectedOutput := generateOutput(tc.expectedOutput)

			if result := PluginMain(input, &testOutput); result != nil {
				if tc.expectError {
					if fmt.Sprint(result) == tc.expectErrorMsg {
						t.Logf("Expected error: {%s} catched, done.", tc.expectErrorMsg)
						return

					} else if tc.expectErrorMsg == "" {
						return
					} else {
						t.Logf("Unexpected error {%s} occurred, expect {%s}.", result, tc.expectErrorMsg)
						t.Fatal("Unexpected error occurred.")
					}

				}
				t.Logf("Failed to run PluginMain when input is %s.", tc.input)
				t.Error(result)
			}

			if testOutput.EqualWith(expectedOutput) {
				t.Log("Done")
			} else {
				t.Errorf("Failed to get expected output when input is %s.", tc.input)
			}

		})
	}
}
