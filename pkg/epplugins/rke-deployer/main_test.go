/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package rkedeployer

import (
	"errors"
	"fmt"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	mock_utils "github.com/intel/edge-conductor/pkg/eputils/mock"
	repoutils "github.com/intel/edge-conductor/pkg/eputils/repoutils"
	mock_repoutils "github.com/intel/edge-conductor/pkg/eputils/repoutils/mock"
	"github.com/intel/edge-conductor/pkg/executor"
	mock_executor "github.com/intel/edge-conductor/pkg/executor/mock"
	"os"
	"path/filepath"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	errCreateRKE = errors.New("Failed to create RKE cluster")
	errPullFile  = errors.New("Pulling file failure.")
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
		expectRunCmdRet       error
		expectPullFileRet     error
		expectError           bool
		expectErrorMsg        string
	}{
		{
			name: "RKE deploy test OK",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"type": "rke", "config": "testdata/rke_cluster.yml", "export_config_folder": "testdata"}}, "runtimebin": "testdata", "runtimedir": "testdata"}`),
				"files":     []byte(`{"files":[{"url": "", "hash":"", "hashtype":"sha256", "mirrorurl": "https://github.com/rancher/rke/releases/download/v1.2.11/rke_linux-amd64", "urlreplacement": {"origin": "://.", "new": "binary"}}]}`),
			},
			expectRunCmdRet:   nil,
			expectPullFileRet: nil,
			expectError:       false,
			expectErrorMsg:    "",
		},
		{
			name: "RKE deploy test fail without input files",
			input: map[string][]byte{
				"ep-params": nil,
				"rkeconfig": []byte(`{"content":"hello-world"}`),
			},
			expectRunCmdRet:   nil,
			expectPullFileRet: nil,
			expectError:       true,
			expectErrorMsg:    eputils.GetError("errInputArryEmpty").Error(),
		},
		{
			name: "RKE deploy test fail due to running RKE fail",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"type": "rke", "config": "testdata/rke_cluster.yml", "export_config_folder": "testdata"}}, "runtimebin": "testdata", "runtimedir": "testdata"}`),
				"files":     []byte(`{"files":[{"url": "", "hash":"", "hashtype":"sha256", "mirrorurl": "https://github.com/rancher/rke/releases/download/v1.2.11/rke_linux-amd64", "urlreplacement": {"origin": "://.", "new": "binary"}}]}`),
			},
			expectRunCmdRet:   errCreateRKE,
			expectPullFileRet: nil,
			expectError:       true,
			expectErrorMsg:    "Failed to create RKE cluster",
		},
		{
			name: "RKE deploy test fail due to pulling file fail",
			input: map[string][]byte{
				"ep-params": []byte(`{"kitconfig": {"Cluster": {"type": "rke", "config": "testdata/rke_cluster.yml", "export_config_folder": "testdata"}}, "runtimebin": "testdata", "runtimedir": "testdata"}`),
				"files":     []byte(`{"files":[{"url": "", "hash":"", "hashtype":"sha256", "mirrorurl": "https://github.com/rancher/rke/releases/download/v1.2.11/rke_linux-amd64", "urlreplacement": {"origin": "://.", "new": "binary"}}]}`),
			},
			expectRunCmdRet:   nil,
			expectPullFileRet: errPullFile,
			expectError:       true,
			expectErrorMsg:    eputils.GetError("errPullingFile").Error(),
		},
	}

	errMakeDir := eputils.MakeDir("testdata")
	require.NoError(t, errMakeDir, "Make dir Error:")

	err := eputils.WriteStringToFile("test", filepath.Join("testdata", "rke"))
	require.NoError(t, err, "Write String To File Error:")

	defer os.RemoveAll("testdata")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMkdirWrapper := mock_utils.NewMockFileWrapper(ctrl)
			patch, err := mpatch.PatchMethod(eputils.MakeDir, mockMkdirWrapper.MakeDir)
			defer unpatch(t, patch)
			if err != nil {
				t.Fatal(err)
			}
			mockMkdirWrapper.EXPECT().MakeDir(gomock.Any()).AnyTimes().Return(nil)

			mockFileWrapper := mock_utils.NewMockFileWrapper(ctrl)
			patch, err = mpatch.PatchMethod(eputils.WriteStringToFile, mockFileWrapper.WriteStringToFile)
			defer unpatch(t, patch)
			if err != nil {
				t.Fatal(err)
			}
			mockFileWrapper.EXPECT().WriteStringToFile(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			mockExecWrapper := mock_utils.NewMockExecWrapper(ctrl)
			patch, err = mpatch.PatchMethod(eputils.RunCMDEx, mockExecWrapper.RunCMDEx)
			defer unpatch(t, patch)
			if err != nil {
				t.Fatal(err)
			}
			mockExecWrapper.EXPECT().RunCMDEx(gomock.Any(), gomock.Any()).AnyTimes().Return("", tc.expectRunCmdRet)

			mockRepoWrapper := mock_repoutils.NewMockRepoUtilsInterface(ctrl)
			patch, err = mpatch.PatchMethod(repoutils.PullFileFromRepo, mockRepoWrapper.PullFileFromRepo)
			defer unpatch(t, patch)
			if err != nil {
				t.Fatal(err)
			}
			mockRepoWrapper.EXPECT().PullFileFromRepo(gomock.Any(), gomock.Any()).AnyTimes().Return(tc.expectPullFileRet)

			mockSchemaWrapper := mock_utils.NewMockSchemaWrapper(ctrl)
			patch, err = mpatch.PatchMethod(eputils.LoadJsonFile, mockSchemaWrapper.LoadJsonFile)
			defer unpatch(t, patch)
			if err != nil {
				t.Fatal(err)
			}
			mockSchemaWrapper.EXPECT().LoadJsonFile(gomock.Any()).AnyTimes().Return(nil, nil)

			mockExecutorWrapper := mock_executor.NewMockExecutorWrapper(ctrl)
			patch, err = mpatch.PatchMethod(executor.Run, mockExecutorWrapper.Run)
			defer unpatch(t, patch)
			if err != nil {
				t.Fatal(err)
			}
			mockExecutorWrapper.EXPECT().Run(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			patch, err = mpatch.PatchMethod(os.UserHomeDir, func() (string, error) { return "testdata", nil })
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch)

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
