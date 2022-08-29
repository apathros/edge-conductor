/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package serviceinjector

import (
	docker "ep/pkg/eputils/docker"
	dockermock "ep/pkg/eputils/docker/mock"
	"testing"

	gomock "github.com/golang/mock/gomock"
	mpatch "github.com/undefinedlabs/go-mpatch"
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
		expectError           bool
	}{
		{
			name: "Change_Service_URL",
			input: map[string][]byte{
				"ep-params":     []byte(`{"kitconfig":{"Parameters": {"global_settings": {"provider_ip": "test","registry_port": "9000"}, "customconfig": {"registry": {"user": "test", "password": "test123"}}}}}`),
				"downloadfiles": []byte(`{"files":[{"mirrorurl":"http://localhost","url":"http://127.0.0.1"},{"mirrorurl":"http://localoverride","url":"http://override"}]}`),
				"serviceconfig": []byte(`{"components":[{"url":"http://127.0.0.1","supported-clusters": ["default"],"chartoverride":"http://override","images":["k8s.gcr.io/ingress-nginx/controller:v1.1.0"]}]}`),
			},
			expectedOutput: map[string][]byte{
				"serviceconfig": []byte(`{"components":[{"url":"http://localhost","supported-clusters": ["default"],"chartoverride":"http://localoverride","images":["10.10.10.10:5678/k8s.gcr.io/ingress-nginx/controller:v1.1.0"]}]}`),
			},
			expectError: false,
		},
		{
			name: "No_Change_for_Repo",
			input: map[string][]byte{
				"ep-params":     []byte(`{"kitconfig":{"Parameters": {"global_settings": {"provider_ip": "test","registry_port": "9000"}, "customconfig": {"registry": {"user": "test", "password": "test123"}}}}}`),
				"downloadfiles": []byte(`{"files":[{"mirrorurl":"http://localhost","url":"http://127.0.0.1"}]}`),
				"serviceconfig": []byte(`{"components":[{"url":"http://127.0.0.1","type":"repo","images":["k8s.gcr.io/ingress-nginx/controller:v1.1.0"]}]}`),
			},
			expectedOutput: map[string][]byte{
				"serviceconfig": []byte(`{"components":[{"url":"http://127.0.0.1","type":"repo","images":["10.10.10.10:5678/k8s.gcr.io/ingress-nginx/controller:v1.1.0"]}]}`),
			},
			expectError: false,
		},
		{
			name: "Not_Downloaded",
			input: map[string][]byte{
				"ep-params":     []byte(`{"kitconfig":{"Parameters": {"global_settings": {"provider_ip": "test","registry_port": "9000"}, "customconfig": {"registry": {"user": "test", "password": "test123"}}}}}`),
				"downloadfiles": []byte(`{"files":[]}`),
				"serviceconfig": []byte(`{"components":[{"url":"http://127.0.0.1","supported-clusters": ["default"],"chartoverride":"http://override","images":["k8s.gcr.io/ingress-nginx/controller:v1.1.0"]}]}`),
			},
			expectedOutput: map[string][]byte{
				"serviceconfig": []byte(`{}`),
			},
			expectError: false,
		},
		{
			name: "Not_Downloaded_Override",
			input: map[string][]byte{
				"ep-params":     []byte(`{"kitconfig":{"Parameters": {"global_settings": {"provider_ip": "test","registry_port": "9000"}, "customconfig": {"registry": {"user": "test", "password": "test123"}}}}}`),
				"downloadfiles": []byte(`{"files":[{"mirrorurl":"http://localhost","url":"http://127.0.0.1"}]}`),
				"serviceconfig": []byte(`{"components":[{"url":"http://127.0.0.1","supported-clusters": ["default"],"chartoverride":"http://override","images":["k8s.gcr.io/ingress-nginx/controller:v1.1.0"]}]}`),
			},
			expectedOutput: map[string][]byte{
				"serviceconfig": []byte(`{}`),
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockDockerCli := dockermock.NewMockDockerClientWrapperImage(ctrl)
			p1, err := mpatch.PatchMethod(docker.TagImageToLocal, mockDockerCli.TagImageToLocal)
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, p1)
			mockDockerCli.EXPECT().TagImageToLocal(gomock.Any(), gomock.Any()).AnyTimes().Return("10.10.10.10:5678/k8s.gcr.io/ingress-nginx/controller:v1.1.0", nil)

			testOutput := generateOutput(nil)
			expectedOutput := generateOutput(tc.expectedOutput)

			if result := PluginMain(input, &testOutput); result != nil {
				if tc.expectError {
					t.Log("Done")
					return
				} else {
					t.Logf("Failed to run PluginMain when input is %s.", tc.input)
					t.Error(result)
				}
			}

			if testOutput.EqualWith(expectedOutput) {
				t.Log("Done")
			} else {
				t.Errorf("Failed to get expected output when input is %s.", tc.input)
				t.Errorf("Expected: %s", tc.expectedOutput)
				t.Errorf("Found: %s", testOutput)
			}
		})
	}
}
