/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

//nolint: dupl
package kindremover

import (
	eputils "ep/pkg/eputils"
	repoutils "ep/pkg/eputils/repoutils"
	repomock "ep/pkg/eputils/repoutils/mock"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"testing"

	gomock "github.com/golang/mock/gomock"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var testerr = errors.New("test error case")

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPluginMain(t *testing.T) {

	func_error_filepull := func(ctrl *gomock.Controller) []*mpatch.Patch {
		mockRepoWrapper := repomock.NewMockRepoUtilsInterface(ctrl)
		p1, err := mpatch.PatchMethod(repoutils.PullFileFromRepo, mockRepoWrapper.PullFileFromRepo)
		if err != nil {
			t.Fatal(err)
		}
		mockRepoWrapper.EXPECT().PullFileFromRepo(gomock.Any(), gomock.Any()).Return(testerr)

		return []*mpatch.Patch{p1}
	}

	func_error_chmod := func(ctrl *gomock.Controller) []*mpatch.Patch {
		mockRepoWrapper := repomock.NewMockRepoUtilsInterface(ctrl)
		p1, err := mpatch.PatchMethod(repoutils.PullFileFromRepo, mockRepoWrapper.PullFileFromRepo)
		if err != nil {
			t.Fatal(err)
		}
		mockRepoWrapper.EXPECT().PullFileFromRepo(gomock.Any(), gomock.Any()).Return(nil)

		p2, err := mpatch.PatchMethod(os.Chmod, func(name string, mode fs.FileMode) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1, p2}
	}

	func_error_cmd := func(ctrl *gomock.Controller) []*mpatch.Patch {
		mockRepoWrapper := repomock.NewMockRepoUtilsInterface(ctrl)
		p1, err := mpatch.PatchMethod(repoutils.PullFileFromRepo, mockRepoWrapper.PullFileFromRepo)
		if err != nil {
			t.Fatal(err)
		}
		mockRepoWrapper.EXPECT().PullFileFromRepo(gomock.Any(), gomock.Any()).Return(nil)

		p2, err := mpatch.PatchMethod(os.Chmod, func(name string, mode fs.FileMode) error { return nil })
		if err != nil {
			t.Fatal(err)
		}
		p3, err := mpatch.PatchMethod(eputils.RunCMDEx, func(*exec.Cmd, bool) (string, error) { return "log", testerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1, p2, p3}
	}

	func_successful := func(ctrl *gomock.Controller) []*mpatch.Patch {
		mockRepoWrapper := repomock.NewMockRepoUtilsInterface(ctrl)
		p1, err := mpatch.PatchMethod(repoutils.PullFileFromRepo, mockRepoWrapper.PullFileFromRepo)
		if err != nil {
			t.Fatal(err)
		}
		mockRepoWrapper.EXPECT().PullFileFromRepo(gomock.Any(), gomock.Any()).Return(nil)

		p2, err := mpatch.PatchMethod(os.Chmod, func(name string, mode fs.FileMode) error { return nil })
		if err != nil {
			t.Fatal(err)
		}
		p3, err := mpatch.PatchMethod(eputils.RunCMDEx, func(*exec.Cmd, bool) (string, error) { return "log", nil })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1, p2, p3}
	}

	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           error
		funcBeforeTest        func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "CASE/successful",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimebin":"testdata"}`),
				"files":     []byte(`{"files":[{"mirrorurl":"testbinary"}]}`),
			},

			expectError:    nil,
			funcBeforeTest: func_successful,
		},
		{
			name: "CASE/error001",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimebin":"testdata"}`),
				"files":     []byte(`{"files":[{"mirrorurl":"testbinary"}]}`),
			},

			expectError:    testerr,
			funcBeforeTest: func_error_filepull,
		},
		{
			name: "CASE/error002",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimebin":"testdata"}`),
				"files":     []byte(`{"files":[{"mirrorurl":"testbinary"}]}`),
			},

			expectError:    testerr,
			funcBeforeTest: func_error_chmod,
		},
		{
			name: "CASE/error003",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimebin":"testdata"}`),
				"files":     []byte(`{"files":[{"mirrorurl":"testbinary"}]}`),
			},

			expectError:    testerr,
			funcBeforeTest: func_error_cmd,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest(ctrl)
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}

			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			testOutput := generateOutput(nil)

			err := PluginMain(input, &testOutput)

			if tc.expectError != nil {
				if err == nil {
					t.Error("Expected error but no error found.")
				} else {
					if fmt.Sprint(err) == fmt.Sprint(tc.expectError) {
						t.Log("Error expected.")
					} else {
						t.Error("Expect:", tc.expectError, "; But found:", err)
					}
				}
			} else {
				if err != nil {
					t.Error("Unexpected Error:", err)
				}
			}

			t.Log("Done")
		})
	}
}
