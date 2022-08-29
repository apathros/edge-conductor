/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package kindparser

import (
	"ep/pkg/eputils"
	"errors"
	"strings"
	"testing"
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

var cluster_manifest_miss_imgnode = []byte(`{
		"cluster_providers":[
		{
			"name":"kind",
			"images":[
				{"name":"img_haproxy","repo_tag":""}
				],
			"binaries":[{"name":"kindtool","url":"","sha256":""}]
		},
		{
			"name":"rke",
			"binaries":[{"name":"rketool","url":"","sha256":""}]
		}
	]}`)
var cluster_manifest_miss_imghaproxy = []byte(`{
		"cluster_providers":[
		{
			"name":"kind",
	 		"images":[
				{"name":"img_node","repo_tag":""}
				],
			"binaries":[{"name":"kindtool","url":"","sha256":""}]
		},
		{
			"name":"rke",
			"binaries":[{"name":"rketool","url":"","sha256":""}]
		}
	]}`)

func TestPluginMain(t *testing.T) {
	var cases = []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           error
	}{
		{
			name: "Parse_a_non-existent_KIND_config",
			input: map[string][]byte{
				"ep-params":        []byte(`{"kitconfig": {"global_settings": {"provider_ip": "test"}}}`),
				"cluster-manifest": cluster_manifest,
			},

			expectedOutput: map[string][]byte{

				"nodes":         nil,
				"docker-images": []byte(`{"images": [{"name":"kind", "url":""}, {"name":"kindhaproxy", "url": ""}]}`),
				"files":         []byte(`{"files":[{"url": "", "hash":"", "hashtype":"sha256", "mirrorurl": "", "urlreplacement": {"origin": "://.", "new": "binary"}}]}`),
			},
			expectError: nil,
		},
		{
			name: "KIND_manifest_lost",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"global_settings": {"provider_ip": "test"}}}`),
			},

			expectedOutput: nil,
			expectError:    eputils.GetError("errManifest"),
		},
		{
			name: "Malformed `ep-params`",
			input: map[string][]byte{
				"ep-params":        []byte(`{}`),
				"cluster-manifest": cluster_manifest,
			},

			expectedOutput: map[string][]byte{

				"nodes":         nil,
				"docker-images": []byte(`{"images": [{"name":"kind", "url":""}, {"name":"kindhaproxy", "url": ""}]}`),
				"files":         []byte(`{"files":[{"url": "", "hash":"", "hashtype":"sha256", "mirrorurl": "", "urlreplacement": {"origin": "://.", "new": "binary"}}]}`),
			},
			expectError: nil,
		},
		{
			name: "KIND_manifest_miss_imgnode",
			input: map[string][]byte{
				"ep-params":        []byte(`{"kitconfig": {"global_settings": {"provider_ip": "test"}}}`),
				"cluster-manifest": cluster_manifest_miss_imgnode,
			},

			expectedOutput: nil,
			expectError:    eputils.GetError("errImage"),
		},
		{
			name: "KIND_manifest_miss_imghaproxy",
			input: map[string][]byte{
				"ep-params":        []byte(`{"kitconfig": {"global_settings": {"provider_ip": "test"}}}`),
				"cluster-manifest": cluster_manifest_miss_imghaproxy,
			},

			expectedOutput: nil,
			expectError:    eputils.GetError("errImage"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			testOutput := generateOutput(nil)

			expectedOutput := generateOutput(tc.expectedOutput)

			result := PluginMain(input, &testOutput)

			if !isExpectedError(result, tc.expectError) {
				t.Error("Output unexpected")
			}

			if !testOutput.EqualWith(expectedOutput) {
				t.Errorf("Failed to get expected output when input is %s.", tc.input)
			}

		})
	}

}

func isExpectedError(returnErr error, wantError error) bool {
	if !errors.Is(returnErr, wantError) &&
		(returnErr == nil || wantError == nil || !strings.Contains(returnErr.Error(), wantError.Error())) {
		return false
	}
	return true
}
