/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

//nolint: dupl
package nodejoindeploy

import (
	"testing"
	// TODO: Add Plugin Unit Test Imports Here
)

func TestPluginMain(t *testing.T) {
	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           bool
	}{
		// TODO: Add the values to complete your test cases.
		// Add the values for input and expectedoutput with particular struct marshal data in json format.
		// They will be used to generate "SchemaMapData" as inputs and expected outputs of plugins under test.
		// if the inputs in the Plugin Input List is not required in your test case, keep the value as nil.
		{
			name: "CASE/001",
			input: map[string][]byte{
				"ep-params":     nil,
				"downloadfiles": nil,
			},

			expectError: false,
		},

		{
			name: "CASE/002",
			input: map[string][]byte{
				"ep-params":     nil,
				"downloadfiles": nil,
			},

			expectError: true,
		},
	}

	// Optional: add setup for the test series
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Run test cases in parallel if necessary.
			// t.Parallel()

			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			testOutput := generateOutput(nil)

			// TODO: Remove the '//' before following check condition to enable plugin test.
			// if result := PluginMain(input, &testOutput); result != nil {
			// 	if tc.expectError {
			// 		t.Log("Error expected.")
			// 		return
			// 	} else {
			// 		t.Logf("Failed to run PluginMain when input is %s.", tc.input)
			// 		t.Error(result)
			// 	}
			// }

			_ = testOutput
			// TODO: Add check conditions
		})
	}

	// Optional: add teardown for the test series
}
