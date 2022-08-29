/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package servicebuild

import (
	pluginapi "ep/pkg/api/plugins"
	"ep/pkg/executor"
	"errors"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	mpatch "github.com/undefinedlabs/go-mpatch"
	// TODO: Add Plugin Unit Test Imports Here
)

var (
	epParam        = []byte(`{"kitconfig": {"Parameters": {"customconfig": {"registry": {"user": "a", "password": "b"}}, "global_settings": {"registry_port": "9000", "provider_ip": "test"}}}, "runtimedir": "aa", "workspace": "aa"}`)
	controllerNode = []byte(`{"ip":"aa", "ssh_port":"aa"}`)
	serviceConfig  = []byte(`{"components": [{"executor": { "build": "test" },"name": "castlelake", "type": "repo", "url": "file://release", "override": "http://override", "resources": [{"name": "helm", "value": "http://helm"}, {"name": "other", "value": "http://other"}, {"name": "etcd", "value": "http://etcd"} ]}]}`)

	// incorrectEpParam       = []byte(`{"kitconfig": {"Parameters": {"global_settings": {"registry_port": "9000", "provider_ip": "test"}}}, "runtimedir": "aa", "workspace": "aa"}`)
	// incorrectServiceConfig = []byte(`{"components": [{"name": "castlelake", "type": "repo", "override": "bb"}]}`)
	// noServiceConfig        = []byte(`{"components": []}`)
	executorRunError = errors.New("executor Run failed")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPluginMain(t *testing.T) {

	func_executorRun_err := func(ctrl *gomock.Controller) []*mpatch.Patch {
		patchExecutorRun, err := mpatch.PatchMethod(executor.Run, func(specFile string, epparams *pluginapi.EpParams, value interface{}) error { return executorRunError })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{patchExecutorRun}
	}
	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           bool
		wantErr               error
		funcBeforeTest        func(*gomock.Controller) []*mpatch.Patch
	}{
		// TODO: Add the values to complete your test cases.
		// Add the values for input and expectedoutput with particular struct marshal data in json format.
		// They will be used to generate "SchemaMapData" as inputs and expected outputs of plugins under test.
		// if the inputs in the Plugin Input List is not required in your test case, keep the value as nil.
		{
			name: "CASE/001",
			input: map[string][]byte{
				"ep-params":       epParam,
				"controller_node": controllerNode,
				"serviceconfig":   serviceConfig,
			},
			expectError:    true,
			wantErr:        executorRunError,
			funcBeforeTest: func_executorRun_err,
		},

		{
			name: "CASE/002",
			input: map[string][]byte{
				"ep-params":     nil,
				"serviceconfig": nil,
			},
			expectedOutput: nil,
			expectError:    false,
		},
	}

	// Optional: add setup for the test series
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var plist []*mpatch.Patch
			if tc.funcBeforeTest != nil {
				plist = tc.funcBeforeTest(ctrl)
			}

			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			testOutput := generateOutput(tc.expectedOutput)
			result := PluginMain(input, &testOutput)
			if tc.expectError {
				if result == nil {
					t.Error("Expected error but no error found.")
				} else {
					if fmt.Sprint(result) == fmt.Sprint(tc.wantErr) {
						t.Log("Error expected.")
					} else {
						t.Error("Expect:", tc.expectError, "; But found:", result)
					}
				}
			} else {
				if result != nil {
					t.Error("Unexpected Error:", result)
				}
			}
			for _, p := range plist {
				unpatch(t, p)
			}

			_ = testOutput
		})
	}
}
