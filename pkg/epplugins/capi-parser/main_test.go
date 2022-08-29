/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package capiparser

import (
	"ep/pkg/eputils"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func readf(fn string) []byte {
	jsonFile, _ := os.Open(fn)
	fb, _ := ioutil.ReadAll(jsonFile)
	return fb
}

func TestPluginMain(t *testing.T) {
	cases := []struct {
		name           string
		input          map[string][]byte
		expectedOutput []byte
		expectError    bool
		expectErrorMsg string
	}{
		{
			name: "kitconfig_lost",
			input: map[string][]byte{
				"ep-params": nil,
			},
			expectError:    true,
			expectErrorMsg: eputils.GetError("errParameter").Error(),
		},
		{
			name: "infra_provider_lost",
			input: map[string][]byte{
				"ep-params":        readf("test_files/infra_provider_lost/ep-params.json"),
				"cluster-manifest": readf(`test_files/infra_provider_lost/cluster-manifest.json`),
			},

			expectedOutput: nil,
			expectError:    true,
			expectErrorMsg: "Failed to get CAPI Provider config in Manifest.",
		},
		{
			name: "multiple_infra_provider",
			input: map[string][]byte{
				"ep-params":        readf("test_files/multiple_infra_provider/ep-params.json"),
				"cluster-manifest": readf("test_files/multiple_infra_provider/cluster-manifest.json"),
			},
			expectedOutput: nil,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errCAPIProvider").Error(),
		},
		{
			name: "capi_manifest_lost",
			input: map[string][]byte{
				"ep-params": readf("test_files/capi_manifest_lost/ep-params.json"),
			},

			expectedOutput: nil,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errCAPIManifest").Error(),
		},
		{
			name: "provider_binary_list_missing",
			input: map[string][]byte{
				"ep-params":        readf("test_files/provider_binary_list_missing/ep-params.json"),
				"cluster-manifest": readf("test_files/provider_binary_list_missing/cluster-manifest.json"),
			},
			expectedOutput: nil,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errCAPIKindLost").Error(),
		},
		{
			name: "kind_info_lost",
			input: map[string][]byte{
				"ep-params":        readf("test_files/kind_info_lost/ep-params.json"),
				"cluster-manifest": readf("test_files/kind_info_lost/cluster-manifest.json"),
			},
			expectedOutput: nil,
			expectError:    true,
			expectErrorMsg: eputils.GetError("errCAPIKindLost").Error(),
		},
		{
			name: "fail_to_find_container_image_for_kind",
			input: map[string][]byte{
				"ep-params":        readf("test_files/fail_to_find_container_image_for_kind/ep-params.json"),
				"cluster-manifest": readf("test_files/fail_to_find_container_image_for_kind/cluster-manifest.json"),
			},
			expectError:    true,
			expectErrorMsg: eputils.GetError("errMgmtCluster").Error(),
		},
		{
			name: "success",
			input: map[string][]byte{
				"ep-params":        readf("test_files/success/ep-params.json"),
				"cluster-manifest": readf("test_files/success/cluster-manifest.json"),
			},
			expectedOutput: readf("test_files/success/output_cluster_manifest.json"),
			expectError:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			testOutput := generateOutput(nil)

			if result := PluginMain(input, &testOutput); result != nil {
				if tc.expectError {
					if fmt.Sprint(result) == tc.expectErrorMsg {
						t.Logf("Expected error: {%s} catched, done.", tc.expectErrorMsg)
						return
					} else if tc.expectErrorMsg == "" {
						return
					} else {
						t.Logf("Expect error: {%s} and active is %s", tc.expectErrorMsg, result)
						t.Fatal("Unexpected error occurred.")
					}

				}
				t.Logf("Failed to run PluginMain when input is %s.", tc.input)
				t.Error(result)
			}

			expOut := generateOutput(nil)

			_ = expOut.UnmarshalBinary(tc.expectedOutput)

			if testOutput.EqualWith(expOut) {
				t.Log("Output expected.")
			} else {
				testOstr, _ := testOutput.MarshalBinary()
				expectOstr, _ := expOut.MarshalBinary()
				t.Logf("Output is %s, Expectoutput is %s", testOstr, expectOstr)
				t.Errorf("Failed to get expected output when input is %s.", tc.input)
			}

		})
	}
}

func TestInitStructFunc(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{
			name:  "ep-params",
			input: __name("ep-params"),
		},
		{
			name:  "cluster-manifest",
			input: __name("cluster-manifest"),
		},
		{
			name:  "docker-images",
			input: __name("docker-images"),
		},
		{
			name:  "files",
			input: __name("files"),
		},
	}

	// Optional: add setup for the test series
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eputils.SchemaStructNew(tc.input)

		})
	}
}
