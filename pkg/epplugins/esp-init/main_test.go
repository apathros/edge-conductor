/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package espinit

import (
	"ep/pkg/eputils"
	mock_utils "ep/pkg/eputils/mock"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	gomock "github.com/golang/mock/gomock"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	testerrmsg = "test error"
	testerr    = errors.New(testerrmsg)
	errError   = errors.New("error")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func getRuntimeFolder() string {
	_, cf, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return filepath.Dir(cf)
}

func TestPluginMain(t *testing.T) {
	faketarball := filepath.Join(getRuntimeFolder(), "testdata", "fakeesp1.0.tar.gz")
	// Cleanup test data
	if err := os.RemoveAll("testdata/esp"); err != nil {
		t.Error("Failed to remove test data", err)
	}

	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           bool
		expectErrorMsg        string
	}{
		{
			name: "success case1",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake-config2.yml"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    false,
			expectErrorMsg: "",
		},
		{
			name: "success case2",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake-config1.yml"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    false,
			expectErrorMsg: "",
		},
		{
			name: "Error case: Could not read config file from viper!",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake.cfg"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: "While parsing config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `test` into map[string]interface {}",
		},
		{
			name: "Error Case: no os session in top config.",
			input: map[string][]byte{
				"ep-params":            []byte(`{"kitconfig": {"global_settings": {"provider_ip": "test"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"","rel_sha256":"","rel_version":""}}`),
			},
			expectError:    true,
			expectErrorMsg: eputils.GetError("errOSSession").Error(),
		},
		{
			name: "Error Case: os provider not ESP",
			input: map[string][]byte{
				"ep-params":            []byte(`{"kitconfig":{"OS": {"provider": "none", "config": ""}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"","rel_sha256":"","rel_version":""}}`),
			},
			expectError:    true,
			expectErrorMsg: eputils.GetError("errOSProvider").Error(),
		},
		{
			name: "Error Case: no ESP manifest",
			input: map[string][]byte{
				"ep-params":            []byte(`{"kitconfig":{"OS": {"provider": "esp", "config": ""}}}`),
				"os-provider-manifest": []byte(`{}`),
			},
			expectError:    true,
			expectErrorMsg: eputils.GetError("errESPManifest").Error(),
		},
		{
			name: "Error Case: invalid file",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": ""}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://testdata/nodata","rel_sha256":"","rel_version":""}}`),
			},
			expectError:    true,
			expectErrorMsg: eputils.GetError("errDownload").Error(),
		},
		{
			name: "Error Case: invalid sha256",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": ""}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"abcd","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: eputils.GetError("errShaCheckFailed").Error(),
		},
		{
			name: "Error Case: no config",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "nodata"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: "open nodata: no such file or directory",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			testOutput := generateOutput(nil)

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
				}
			}

			_ = testOutput
		})
	}

	// Cleanup test data
	if err := os.RemoveAll("testdata/esp"); err != nil {
		t.Error("Failed to remove test data", err)
	}
}

func TestPluginMainSystemError(t *testing.T) {
	// Cleanup test data
	if err := os.RemoveAll("testdata/esp"); err != nil {
		t.Error("Failed to remove test data", err)
	}

	func_err_createCleanDir := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(createCleanDir, func(string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1}
	}

	func_err_UncompressTgz := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(eputils.UncompressTgz, func(string, string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1}
	}

	func_err_removefile := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(eputils.RemoveFile, func(string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		p2, err := mpatch.PatchMethod(eputils.FileExists, func(string) bool { return false })
		if err != nil {
			t.Fatal(err)
		}
		p3, err := mpatch.PatchMethod(eputils.CheckFileSHA256, func(string, string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}

		return []*mpatch.Patch{p1, p2, p3}
	}

	func_err_WriteStringToFile := func(ctrl *gomock.Controller) []*mpatch.Patch {
		// Cleanup test data
		if err := os.RemoveAll("testdata/esp"); err != nil {
			t.Error("Failed to remove test data", err)
		}
		p1, err := mpatch.PatchMethod(eputils.WriteStringToFile, func(string, string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1}
		// return []*mpatch.Patch{p1, p2, p3, p4}
	}

	func_err_deleteStringToFile := func(ctrl *gomock.Controller) []*mpatch.Patch {
		// Cleanup test data
		if err := os.RemoveAll("testdata/esp"); err != nil {
			t.Error("Failed to remove test data", err)
		}
		p1, err := mpatch.PatchMethod(deleteStrfromFile, func(string, []string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1}
	}

	func_err_appendStringToFile := func(ctrl *gomock.Controller) []*mpatch.Patch {
		// Cleanup test data
		if err := os.RemoveAll("testdata/esp"); err != nil {
			t.Error("Failed to remove test data", err)
		}
		p1, err := mpatch.PatchMethod(appendStringToFile, func(string, string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1}
	}

	func_err_CheckFileSHA256 := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(eputils.CheckFileSHA256, func(string, string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1}
	}

	func_err_remove_config := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(createCleanDir, func(string) error { return nil })
		if err != nil {
			t.Fatal(err)
		}
		p2, err := mpatch.PatchMethod(os.RemoveAll, func(string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		p3, err := mpatch.PatchMethod(checkESPCodebase, func(string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1, p2, p3}
	}

	func_err_RunCMD := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(os.RemoveAll, func(string) error { return nil })
		if err != nil {
			t.Fatal(err)
		}
		p2, err := mpatch.PatchMethod(checkESPCodebase, func(string) error { return nil })
		if err != nil {
			t.Fatal(err)
		}

		p3, err := mpatch.PatchMethod(createCleanDir, func(string) error { return nil })
		if err != nil {
			t.Fatal(err)
		}
		p4, err := mpatch.PatchMethod(eputils.CopyFile, func(string, string) (int64, error) { return 1, nil })
		if err != nil {
			t.Fatal(err)
		}
		p5, err := mpatch.PatchMethod(eputils.RunCMD, func(cmd *exec.Cmd) (string, error) { return "", testerr })
		if err != nil {
			t.Fatal(err)
		}
		p6, err := mpatch.PatchMethod(eputils.CheckFileSHA256, func(string, string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		p7, err := mpatch.PatchMethod(initESPcode, func(string, string, string, string, string) error { return nil })
		if err != nil {
			t.Fatal(err)
		}
		p8, err := mpatch.PatchMethod(eputils.WriteStringToFile, func(string, string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1, p2, p3, p4, p5, p6, p7, p8}
	}

	func_err_remove_all := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(os.RemoveAll, func(string) error { return testerr })
		if err != nil {
			t.Fatal(err)
		}
		p2, err := mpatch.PatchMethod(checkESPCodebase, func(string) error { return nil })
		if err != nil {
			t.Fatal(err)
		}

		p3, err := mpatch.PatchMethod(createCleanDir, func(string) error { return nil })
		if err != nil {
			t.Fatal(err)
		}

		return []*mpatch.Patch{p1, p2, p3}
	}

	faketarball := filepath.Join(getRuntimeFolder(), "testdata", "fakeesp1.0.tar.gz")

	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           bool
		expectErrorMsg        string
		funcBeforeTest        func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "Error createCleanDir",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake.cfg"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: testerrmsg,
			funcBeforeTest: func_err_createCleanDir,
		},
		{
			name: "Error UncompressTgz",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake.cfg"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: testerrmsg,
			funcBeforeTest: func_err_UncompressTgz,
		},
		{
			name: "Error Remove Config",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake.cfg"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: testerrmsg,
			funcBeforeTest: func_err_remove_config,
		},
		{
			name: "Error RunCMD",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake.cfg"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: testerrmsg,
			funcBeforeTest: func_err_RunCMD,
		},
		{
			name: "Error remove file",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake.cfg"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: testerrmsg,
			funcBeforeTest: func_err_removefile,
		},
		{
			name: "Error write string to file",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake.cfg"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: testerrmsg,
			funcBeforeTest: func_err_WriteStringToFile,
		},
		{
			name: "Error delete string to file",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake-config2.yml"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: testerrmsg,
			funcBeforeTest: func_err_deleteStringToFile,
		},
		{
			name: "Error append string to file",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake-config2.yml"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: testerrmsg,
			funcBeforeTest: func_err_appendStringToFile,
		},
		{
			name: "Error Check File SHA256",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake.cfg"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: testerrmsg,
			funcBeforeTest: func_err_CheckFileSHA256,
		},
		{
			name: "Error remove all",
			input: map[string][]byte{
				"ep-params":            []byte(`{"workspace":"testdata","kitconfig":{"OS": {"provider": "esp", "config": "testdata/fake.cfg"}}}`),
				"os-provider-manifest": []byte(`{"esp":{"rel_url":"file://` + faketarball + `","rel_sha256":"aeb3883c4e817f1266cbbb4f6efc44490dbdd0e35e097a665abe54869c04fb38","rel_version":"1.0"}}`),
			},
			expectError:    true,
			expectErrorMsg: testerrmsg,
			funcBeforeTest: func_err_remove_all,
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
				}
			}

			_ = testOutput
		})
	}

	// Cleanup test data
	if err := os.RemoveAll("testdata/esp"); err != nil {
		t.Error("Failed to remove test data", err)
	}
}

func patchFuncStringInBoolOut(t *testing.T, target interface{}, ret bool) *mpatch.Patch {
	p, err := mpatch.PatchMethod(target,
		func(a string) bool {
			return ret
		})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func patchFuncStringInErrorOut(t *testing.T, target interface{}, reterr error) *mpatch.Patch {
	p, err := mpatch.PatchMethod(target,
		func(a string) error {
			return reterr
		})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCreateCleanDirErrRemove(t *testing.T) {
	experr := errError
	p := patchFuncStringInBoolOut(t, eputils.FileExists, true)
	defer unpatch(t, p)
	p = patchFuncStringInErrorOut(t, os.RemoveAll, experr)
	defer unpatch(t, p)
	err := createCleanDir("dir")
	if fmt.Sprint(err) == fmt.Sprint(experr) {
		t.Log("Error expected.")
	} else {
		t.Error("Expect:", experr, "; But found:", err)
	}
}

func TestCreateCleanDirErrMakeDir(t *testing.T) {
	experr := errError
	p := patchFuncStringInBoolOut(t, eputils.FileExists, true)
	defer unpatch(t, p)
	p = patchFuncStringInErrorOut(t, os.RemoveAll, nil)
	defer unpatch(t, p)
	p = patchFuncStringInErrorOut(t, eputils.MakeDir, experr)
	defer unpatch(t, p)
	err := createCleanDir("dir")
	if fmt.Sprint(err) == fmt.Sprint(experr) {
		t.Log("Error expected.")
	} else {
		t.Error("Expect:", experr, "; But found:", err)
	}
}

func TestCheckESPCodebase(t *testing.T) {

	codebase := "test"

	func_err_codebase := func(ctrl *gomock.Controller) []*mpatch.Patch {
		mockFileWrapper := mock_utils.NewMockFileWrapper(ctrl)
		p, err := mpatch.PatchMethod(eputils.FileExists, mockFileWrapper.FileExists)
		if err != nil {
			t.Fatal(err)
		}
		mockFileWrapper.EXPECT().FileExists(codebase).Return(false)
		return []*mpatch.Patch{p}
	}

	func_err_script_build := func(ctrl *gomock.Controller) []*mpatch.Patch {
		mockFileWrapper := mock_utils.NewMockFileWrapper(ctrl)
		p, err := mpatch.PatchMethod(eputils.FileExists, mockFileWrapper.FileExists)
		if err != nil {
			t.Fatal(err)
		}
		mockFileWrapper.EXPECT().FileExists(codebase).Return(true)
		mockFileWrapper.EXPECT().FileExists(filepath.Join(codebase, "build.sh")).Return(false)
		return []*mpatch.Patch{p}
	}

	func_err_script_run := func(ctrl *gomock.Controller) []*mpatch.Patch {
		mockFileWrapper := mock_utils.NewMockFileWrapper(ctrl)
		p, err := mpatch.PatchMethod(eputils.FileExists, mockFileWrapper.FileExists)
		if err != nil {
			t.Fatal(err)
		}
		mockFileWrapper.EXPECT().FileExists(codebase).Return(true)
		mockFileWrapper.EXPECT().FileExists(filepath.Join(codebase, "build.sh")).Return(true)
		mockFileWrapper.EXPECT().FileExists(filepath.Join(codebase, "run.sh")).Return(false)
		return []*mpatch.Patch{p}
	}

	cases := []struct {
		name           string
		expectedErr    error
		funcBeforeTest func(ctrl *gomock.Controller) []*mpatch.Patch
	}{
		{
			name:           "Codebase Error",
			expectedErr:    eputils.GetError("errESPCodebase"),
			funcBeforeTest: func_err_codebase,
		},
		{
			name:           "Build Script Error",
			expectedErr:    eputils.GetError("errESPBuild"),
			funcBeforeTest: func_err_script_build,
		},
		{
			name:           "Run Script Error",
			expectedErr:    eputils.GetError("errESPRun"),
			funcBeforeTest: func_err_script_run,
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

			err := checkESPCodebase(codebase)

			if tc.expectedErr != nil {
				if err == nil {
					t.Error("Expected error but no error found.")
				} else {
					if fmt.Sprint(err) == fmt.Sprint(tc.expectedErr) {
						t.Log("Error expected.")
					} else {
						t.Error("Expect:", tc.expectedErr, "; But found:", err)
					}
				}
			} else {
				if err != nil {
					t.Error("Unexpected Error:", err)
				}
			}
		})
	}

}
