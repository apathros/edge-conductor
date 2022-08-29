/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package debugdump

import (
	"testing"
)

func TestPluginMain(t *testing.T) {
	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           bool
	}{
		{
			name: "simple_test",
			input: map[string][]byte{
				"nodes":               []byte(`{"nodes":[]}`),
				"docker-images":       []byte(`{"images":[]}`),
				"local-docker-images": []byte(`{"images":[]}`),
			},

			expectError: false,
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
					t.Log("Error expected.")
					return
				} else {
					t.Logf("Failed to run PluginMain when input is %s.", tc.input)
					t.Error(result)
				}
			}

			_ = testOutput
		})
	}
}
