/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package rkeinjector

import (
	"fmt"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	mock_utils "github.com/intel/edge-conductor/pkg/eputils/mock"
	repoutils "github.com/intel/edge-conductor/pkg/eputils/repoutils"
	mock_repoutils "github.com/intel/edge-conductor/pkg/eputils/repoutils/mock"
	"os"
	"path/filepath"
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
		expectRunCmdErr       error
		expectPullFileErr     error
		expectErrorMsg        string
	}{
		{
			name: "RKE inject test OK",
			input: map[string][]byte{
				"ep-params":     []byte(`{"kitconfig": {"Cluster": {"provider": "rke", "config": "../../../configs/cluster-provider/rke_cluster.yml"}}, "runtimebin": "testdata", "runtimedir": "testdata"}`),
				"docker-images": []byte(`{"images": []}`),
				"files":         []byte(`{"files":[{"url": "", "hash":"", "hashtype":"sha256", "mirrorurl": "", "urlreplacement": {"origin": "://.", "new": "binary"}}]}`),
			},

			expectError:       false,
			expectRunCmdErr:   nil,
			expectPullFileErr: nil,
			expectErrorMsg:    "",
		},
		{
			name: "RKE inject test Fail",
			input: map[string][]byte{
				"ep-params":     []byte(`{"kitconfig": {"Cluster": {"provider": "rke", "config": "../../../configs/cluster-provider/rke_cluster.yml"}}, "runtimebin": "testdata", "runtimedir": "testdata"}`),
				"docker-images": []byte(`{"images": []}`),
				"files":         []byte(`{"files":[{"url": "", "hash":"", "hashtype":"sha256", "mirrorurl": "", "urlreplacement": {"origin": "://.", "new": "binary"}}]}`),
			},

			expectError:       true,
			expectRunCmdErr:   errRunRKE,
			expectPullFileErr: nil,
			expectErrorMsg:    "Failed to run rke command!",
		},
		{
			name: "RKE inject test Fail - Pull file failure",
			input: map[string][]byte{
				"ep-params":     []byte(`{"kitconfig": {"Cluster": {"provider": "rke", "config": "../../../configs/cluster-provider/rke_cluster.yml"}}, "runtimebin": "testdata", "runtimedir": "testdata"}`),
				"docker-images": []byte(`{"images": []}`),
				"files":         []byte(`{"files":[{"url": "", "hash":"", "hashtype":"sha256", "mirrorurl": "", "urlreplacement": {"origin": "://.", "new": "binary"}}]}`),
			},

			expectError:       true,
			expectRunCmdErr:   nil,
			expectPullFileErr: errPullFile,
			expectErrorMsg:    "Pulling file failure!",
		},
	}

	errMakeDir := eputils.MakeDir("testdata")
	if errMakeDir != nil {
		t.Fatal(errMakeDir)
	}
	err := eputils.WriteStringToFile("test", filepath.Join("testdata", "rke"))
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("testdata")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFileWrapper := mock_utils.NewMockFileWrapper(ctrl)
			patch, err := mpatch.PatchMethod(eputils.MakeDir, mockFileWrapper.MakeDir)
			defer unpatch(t, patch)
			if err != nil {
				t.Fatal(err)
			}
			mockFileWrapper.EXPECT().MakeDir(gomock.Any()).AnyTimes().Return(nil)

			mockRepoWrapper := mock_repoutils.NewMockRepoUtilsInterface(ctrl)
			patch, err = mpatch.PatchMethod(repoutils.PullFileFromRepo, mockRepoWrapper.PullFileFromRepo)
			defer unpatch(t, patch)
			if err != nil {
				t.Fatal(err)
			}
			mockRepoWrapper.EXPECT().PullFileFromRepo(gomock.Any(), gomock.Any()).AnyTimes().Return(tc.expectPullFileErr)

			mockExecWrapper := mock_utils.NewMockExecWrapper(ctrl)
			patch, err = mpatch.PatchMethod(eputils.RunCMD, mockExecWrapper.RunCMD)
			defer unpatch(t, patch)
			if err != nil {
				t.Fatal(err)
			}
			mockExecWrapper.EXPECT().RunCMD(gomock.Any()).AnyTimes().Return("INFO[0012] out:\ntime=\"2021-09-06T16:27:42+08:00\" level=debug msg=\"Loglevel set to [debug]\"\ntime=\"2021-09-06T16:27:42+08:00\" level=debug msg=\"Loading data.json from local source\"\ntime=\"2021-09-06T16:27:42+08:00\" level=debug msg=\"data.json SHA256 checksum: bf80f308534769b239b8d33b115fe4c1085f35ce4e04b2a5eea007d73c327687\"\ntime=\"2021-09-06T16:27:42+08:00\" level=debug msg=\"metadata initialized successfully\"\ntime=\"2021-09-06T16:27:42+08:00\" level=info msg=\"Generating images list for version [v1.20.9-rancher1-1]:\"\nrancher/mirrored-coreos-etcd:v3.4.15-rancher1\nrancher/rke-tools:v0.1.77", tc.expectRunCmdErr)

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

					} else {
						t.Fatal("Unexpected error occurred.")
					}

				}
				t.Logf("Failed to run PluginMain when input is %s.", tc.input)
				t.Error(result)
			}
			t.Log("Done")
		})
	}
}

func TestGetDefaultRkeSystemImages(t *testing.T) {
	cases := []struct {
		name           string
		input          string
		expectedOutput []string
		expectError    bool
		expectErrorMsg string
	}{
		{
			name:           "Get Default System Images OK",
			input:          "INFO[0012] out:\ntime=\"2021-09-06T16:27:42+08:00\" level=debug msg=\"Loglevel set to [debug]\"\ntime=\"2021-09-06T16:27:42+08:00\" level=debug msg=\"Loading data.json from local source\"\ntime=\"2021-09-06T16:27:42+08:00\" level=debug msg=\"data.json SHA256 checksum: bf80f308534769b239b8d33b115fe4c1085f35ce4e04b2a5eea007d73c327687\"\ntime=\"2021-09-06T16:27:42+08:00\" level=debug msg=\"metadata initialized successfully\"\ntime=\"2021-09-06T16:27:42+08:00\" level=info msg=\"Generating images list for version [v1.20.9-rancher1-1]:\"\nrancher/mirrored-coreos-etcd:v3.4.15-rancher1\nrancher/rke-tools:v0.1.77",
			expectedOutput: []string{"ancher/mirrored-coreos-etcd:v3.4.15-rancher1", "rancher/rke-tools:v0.1.77"},
			expectError:    false,
			expectErrorMsg: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := getDefaultRkeSystemImages(tc.input)
			if result[0] == tc.expectedOutput[0] && result[1] == tc.expectedOutput[1] {
				t.Log("Done")
			}

		})
	}
}
