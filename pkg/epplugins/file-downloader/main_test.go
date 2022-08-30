/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package filedownloader

import (
	"fmt"
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	"github.com/intel/edge-conductor/pkg/eputils"
	"github.com/intel/edge-conductor/pkg/eputils/repoutils"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
)

var (
	testErr = fmt.Errorf("test error")
)

func getRuntimeFolder() string {
	_, cf, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return filepath.Dir(cf)
}

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func patchPushFileToRepo(t *testing.T, output string, retError error) {
	//mock: repoutils.PushFileToRepo
	var patch *mpatch.Patch
	var err error
	patch, err = mpatch.PatchMethod(repoutils.PushFileToRepo, func(a, b, c string) (string, error) {
		unpatch(t, patch)
		return output, retError
	})

	if err != nil {
		t.Fatal(err)
	}
}

func patchRemoveFile(t *testing.T, retError error) {
	//mock: repoutils.PushFileToRepo
	var patch *mpatch.Patch
	var err error
	patch, err = mpatch.PatchMethod(eputils.RemoveFile, func(path string) error {
		unpatch(t, patch)
		return retError
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestPluginMain(t *testing.T) {
	fakefile := filepath.Join(getRuntimeFolder(), "testdata", "fakefile")

	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		funcBeforeTest        func()
		expectError           bool
		expectErrorMsg        string
	}{
		{
			name: "Success: no file",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimedir":"testdata"}`),
				"files":     []byte(`{"files":[]}`),
			},
			expectedOutput: map[string][]byte{
				"files": []byte(`{"files":null}`),
			},
			expectError:    false,
			expectErrorMsg: "",
		},
		{
			name: "Error Case: invalid file",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimedir":"testdata"}`),
				"files":     []byte(`{"files":[{"url":"file://nodata","hash":"abc","hashtype":"sha256","mirrorurl":"","urlreplacement":{"new":"test"}}]}`),
			},
			expectError:    true,
			expectErrorMsg: eputils.GetError("errDownload").Error(),
		},
		{
			name: "Error Case: invalid temp folder name",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimedir":"testdata"}`),
				"files":     []byte(`{"files":[{"url":"file://` + fakefile + `","hash":"abc","hashtype":"sha256","mirrorurl":"","urlreplacement":{"new":"test"}}]}`),
			},
			funcBeforeTest: func() {
				var patch *mpatch.Patch
				patch, _ = mpatch.PatchMethod(eputils.CreateFolderIfNotExist, func(path string) error {
					unpatch(t, patch)
					return testErr
				})
			},
			expectError:    true,
			expectErrorMsg: testErr.Error(),
		},
		{
			name: "Error Case: invalid sha256",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimedir":"testdata"}`),
				"files":     []byte(`{"files":[{"url":"file://` + fakefile + `","hash":"abc","hashtype":"sha256","mirrorurl":"","urlreplacement":{"new":"test"}}]}`),
			},
			expectError:    true,
			expectErrorMsg: eputils.GetError("errShaCheckFailed").Error(),
		},
		{
			name: "Error Case: remove file error when call CheckFileSHA256 failed",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimedir":"testdata"}`),
				"files":     []byte(`{"files":[{"url":"file://` + fakefile + `","hash":"abc","hashtype":"sha256","mirrorurl":"","urlreplacement":{"new":"test"}}]}`),
			},
			funcBeforeTest: func() {
				patchRemoveFile(t, testErr)
			},
			expectError:    true,
			expectErrorMsg: eputils.GetError("errShaCheckFailed").Error(),
		},
		{
			name: "Error Case: call PushFileToRepo failed",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimedir":"testdata"}`),
				"files":     []byte(`{"files":[{"url":"file://` + fakefile + `","hash":"","hashtype":"sha256","mirrorurl":"","urlreplacement":{"new":"test"}}]}`),
			},
			funcBeforeTest: func() {
				patchPushFileToRepo(t, "", testErr)
			},
			expectError:    true,
			expectErrorMsg: testErr.Error(),
		},
		{
			name: "Error Case: remove file error after call PushFileToRepo",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimedir":"testdata"}`),
				"files":     []byte(`{"files":[{"url":"file://` + fakefile + `","hash":"","hashtype":"sha256","mirrorurl":"","urlreplacement":{"new":"test"}}]}`),
			},
			funcBeforeTest: func() {
				patchPushFileToRepo(t, "testoutput", nil)
				patchRemoveFile(t, testErr)
			},
			expectError:    true,
			expectErrorMsg: testErr.Error(),
		},
		{
			name: "Success",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimedir":"testdata"}`),
				"files":     []byte(`{"files":[{"url":"file://` + fakefile + `","hash":"9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08","hashtype":"sha256","mirrorurl":"","urlreplacement":{"new":"test"}}]}`),
			},
			expectedOutput: map[string][]byte{
				"files": []byte(`{"files":[{"url":"file://` + fakefile + `","hash":"9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08","hashtype":"sha256","mirrorurl":"","urlreplacement":{"new":"test"},"mirrorurl":"` + "testoutput" + `"}]}`),
			},
			funcBeforeTest: func() {
				patchPushFileToRepo(t, "testoutput", nil)
			},
			expectError:    false,
			expectErrorMsg: "",
		},
		{
			name: "Success: RemoveAll error",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimedir":"testdata"}`),
				"files":     []byte(`{"files":[{"url":"file://` + fakefile + `","hash":"9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08","hashtype":"sha256","mirrorurl":"","urlreplacement":{"new":"test"}}]}`),
			},
			expectedOutput: map[string][]byte{
				"files": []byte(`{"files":[{"url":"file://` + fakefile + `","hash":"9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08","hashtype":"sha256","mirrorurl":"","urlreplacement":{"new":"test"},"mirrorurl":"` + "testoutput" + `"}]}`),
			},
			funcBeforeTest: func() {
				patchPushFileToRepo(t, "testoutput", nil)
				var patch *mpatch.Patch
				var err error
				patch, err = mpatch.PatchMethod(os.RemoveAll, func(path string) error {
					unpatch(t, patch)
					return testErr
				})
				if err != nil {
					t.Fatal(err)
				}
			},
			expectError:    false,
			expectErrorMsg: "",
		},
	}

	// Optional: add setup for the test series
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			testOutput := generateOutput(nil)

			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			err := PluginMain(input, &testOutput)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but no error found.")
				} else {
					if fmt.Sprint(err) == tc.expectErrorMsg {
						t.Log("Output expected.")
					} else {
						t.Error("Expect:", tc.expectErrorMsg, "; But found:", err)
					}
				}
			} else {
				if err != nil {
					t.Error("Unexpected Error:", err)
				} else {
					expectedOutput := generateOutput(tc.expectedOutput)
					if testOutput.EqualWith(expectedOutput) {
						t.Log("Done")
					} else {
						output, _ := testOutput.MarshalBinary()
						t.Errorf("Expect output %s but returned %s.", tc.expectedOutput, output)
					}
				}
			}
		})
	}

	// Cleanup test data
	if err := os.RemoveAll("testdata/tmp"); err != nil {
		t.Error("Failed to remove test data", err)
	}
}

func TestInitStructFunc(t *testing.T) {
	schemaStruct := eputils.SchemaStructNew(__name("ep-params"))
	epParams, ok := schemaStruct.(*pluginapi.EpParams)
	if !ok && (epParams != &pluginapi.EpParams{}) {
		t.Errorf(`ep-params struct init error`)
	}

	schemaStruct = eputils.SchemaStructNew(__name("files"))
	files, ok := schemaStruct.(*pluginapi.Files)
	if !ok && files.Files == nil {
		t.Errorf(`files struct init error`)
	}
}
