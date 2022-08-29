/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package capideinit

import (
	"errors"
	"testing"

	docker "ep/pkg/eputils/docker"
	clientmock "ep/pkg/eputils/docker/mock"
	gomock "github.com/golang/mock/gomock"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	errDockerRunFail = errors.New("docker run fail")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPluginMain(t *testing.T) {
	func_docker_rm_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		mockWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)
		p, err := mpatch.PatchMethod(docker.RemoveContainer, mockWrapperContainer.RemoveContainer)
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}

		mockWrapperContainer.EXPECT().RemoveContainer(gomock.Any()).AnyTimes().Return(errDockerRunFail)
		return []*mpatch.Patch{p}
	}

	func_docker_rm_succeed := func(ctrl *gomock.Controller) []*mpatch.Patch {
		mockWrapperContainer := clientmock.NewMockDockerClientWrapperContainer(ctrl)
		p, err := mpatch.PatchMethod(docker.RemoveContainer, mockWrapperContainer.RemoveContainer)
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}

		mockWrapperContainer.EXPECT().RemoveContainer(gomock.Any()).AnyTimes().Return(nil)
		return []*mpatch.Patch{p}
	}

	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           bool
		funcBeforeTest        func(*gomock.Controller) []*mpatch.Patch
	}{
		// TODO: Add the values to complete your test cases.
		// Add the values for input and expectedoutput with particular struct marshal data in json format.
		// They will be used to generate "SchemaMapData" as inputs and expected outputs of plugins under test.
		// if the inputs in the Plugin Input List is not required in your test case, keep the value as nil.
		{
			name:  "docker rm fail",
			input: map[string][]byte{},

			expectError:    true,
			funcBeforeTest: func_docker_rm_fail,
		},

		{
			name:  "docker rm succeed",
			input: map[string][]byte{},

			expectError:    false,
			funcBeforeTest: func_docker_rm_succeed,
		},
	}

	for _, tc := range cases {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var plist []*mpatch.Patch
		if tc.funcBeforeTest != nil {
			plist = tc.funcBeforeTest(ctrl)
		}

		t.Run(tc.name, func(t *testing.T) {
			// Run test cases in parallel if necessary.
			// t.Parallel()

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

		for _, p := range plist {
			unpatch(t, p)
		}
	}

	_ = __name("capi-deinit")
}
