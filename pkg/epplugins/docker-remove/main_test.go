/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package dockerremove

import (
	docker "ep/pkg/eputils/docker"
	dockermock "ep/pkg/eputils/docker/mock"
	"errors"
	gomock "github.com/golang/mock/gomock"
	mpatch "github.com/undefinedlabs/go-mpatch"
	"testing"
)

var (
	errDocker = errors.New("docker error")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPluginMain(t *testing.T) {
	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           error
	}{
		{
			name: "docker_remove",
			input: map[string][]byte{
				"containers": []byte(`{"containers":[{"name":"test","image": "hello-world:latest"}]}`),
			},
			expectError: nil,
		},
		{
			name: "docker_remove_error",
			input: map[string][]byte{
				"containers": []byte(`{"containers":[{"name":"test","image": "hello-world:latest"}]}`),
			},
			expectError: errDocker,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockDockerRunner := dockermock.NewMockDockerInterface(ctrl)
			patch, err := mpatch.PatchMethod(docker.DockerRemove, mockDockerRunner.DockerRemove)
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch)

			// DockerRemove always successful.
			mockDockerRunner.EXPECT().DockerRemove(gomock.Any()).AnyTimes().Return(tc.expectError)

			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			testOutput := generateOutput(nil)

			if result := PluginMain(input, &testOutput); result != nil {
				if errors.Is(result, tc.expectError) {
					t.Log("Output expected.")
				} else {
					t.Error("Output unexpected")
				}
			}

			_ = testOutput
		})
	}
}
