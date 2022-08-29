/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

//nolint: dupl
package servicelist

import (
	repoutils "ep/pkg/eputils/repoutils"
	repomock "ep/pkg/eputils/repoutils/mock"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	gomock "github.com/golang/mock/gomock"
	mpatch "github.com/undefinedlabs/go-mpatch"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
)

var (
	kubeerr      = errors.New("kubernetes error")
	errNotFound  = errors.New("not found")
	errNoRelease = errors.New("release: not found")

	testerrmsg = "test error"
	testerr    = errors.New(testerrmsg)
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPluginMain(t *testing.T) {

	func_err_PullFileFromRepo := func(ctrl *gomock.Controller) []*mpatch.Patch {
		// Repo Utils
		mockRepoWrapper := repomock.NewMockRepoUtilsInterface(ctrl)
		p1, err := mpatch.PatchMethod(repoutils.PullFileFromRepo, mockRepoWrapper.PullFileFromRepo)
		if err != nil {
			t.Fatal(err)
		}
		mockRepoWrapper.EXPECT().PullFileFromRepo(gomock.Any(), gomock.Any()).Return(kubeerr)
		// Status functions
		p2, err := mpatch.PatchMethod(getHelmStatus, func(string, string, string) string { return "Deployed" })
		if err != nil {
			t.Fatal(err)
		}
		p3, err := mpatch.PatchMethod(getYamlStatus, func(string, string) string { return "Deployed" })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1, p2, p3}
	}

	func_expected_status := func(ctrl *gomock.Controller) []*mpatch.Patch {
		// Repo Utils
		mockRepoWrapper := repomock.NewMockRepoUtilsInterface(ctrl)
		p1, err := mpatch.PatchMethod(repoutils.PullFileFromRepo, mockRepoWrapper.PullFileFromRepo)
		if err != nil {
			t.Fatal(err)
		}
		mockRepoWrapper.EXPECT().PullFileFromRepo(gomock.Any(), gomock.Any()).Return(nil)
		// Status functions
		p2, err := mpatch.PatchMethod(getHelmStatus, func(string, string, string) string { return "Deployed" })
		if err != nil {
			t.Fatal(err)
		}
		p3, err := mpatch.PatchMethod(getYamlStatus, func(string, string) string { return "Deployed" })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1, p2, p3}
	}

	cases := []struct {
		name           string
		input          map[string][]byte
		expectedErr    error
		funcBeforeTest func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "err_test_pull_file_fail",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"components":[{"name":"testhelm","type":"helm"},{"name":"testyaml","type":"yaml"}]}`),
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_PullFileFromRepo,
		},
		{
			name: "simple_list",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"components":[{"name":"testhelm","type":"helm"},{"name":"testyaml","type":"yaml"}]}`),
			},
			expectedErr:    nil,
			funcBeforeTest: func_expected_status,
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

func TestHelmService(t *testing.T) {
	// mock functions
	func_status_run_ok := func(*action.Status, string) (*release.Release, error) {
		return &release.Release{
			Info: &release.Info{Status: "Test"},
		}, nil
	}
	func_status_run_err := func(*action.Status, string) (*release.Release, error) {
		return nil, kubeerr
	}
	func_status_run_err_notdeployed := func(*action.Status, string) (*release.Release, error) {
		return nil, errNoRelease
	}

	// before test functions
	func_init_helm_failed := func(ctrl *gomock.Controller, mockfunc func(*action.Status, string) (*release.Release, error)) []*mpatch.Patch {
		p, err := mpatch.PatchMethod(initHelm, func(string, string) error { return kubeerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p}
	}

	func_status_run_mock := func(ctrl *gomock.Controller, mockfunc func(*action.Status, string) (*release.Release, error)) []*mpatch.Patch {
		cli_status := action.NewStatus(new(action.Configuration))
		p, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cli_status), "Run", mockfunc)
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p}
	}

	cases := []struct {
		name           string
		expectResult   string
		funcBeforeTest func(*gomock.Controller, func(*action.Status, string) (*release.Release, error)) []*mpatch.Patch
		funcStatusRun  func(*action.Status, string) (*release.Release, error)
	}{
		{
			name:           "err_test_wrong_kubeconfig",
			expectResult:   "Unknown",
			funcBeforeTest: func_init_helm_failed,
			funcStatusRun:  nil,
		},
		{
			name:           "err_test_wrong_release_status",
			expectResult:   "Unknown",
			funcBeforeTest: func_status_run_mock,
			funcStatusRun:  func_status_run_err,
		},
		{
			name:           "test_release_status_notdeployed",
			expectResult:   "Not Deployed",
			funcBeforeTest: func_status_run_mock,
			funcStatusRun:  func_status_run_err_notdeployed,
		},
		{
			name:           "test_release_status",
			expectResult:   "Test",
			funcBeforeTest: func_status_run_mock,
			funcStatusRun:  func_status_run_ok,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest(ctrl, tc.funcStatusRun)
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}

			status := getHelmStatus("kubeconfig", "name", "ns")

			if status != tc.expectResult {
				t.Error("Expected", tc.expectResult, "but got", status)
			} else {
				t.Log("Done")
			}
		})
	}
}

func TestInitHelmError(t *testing.T) {
	gActionConfig = nil

	p1, err := mpatch.PatchMethod(filepath.Abs, func(string) (string, error) { return "", testerr })
	if err != nil {
		t.Fatal(err)
	}
	defer unpatch(t, p1)

	if err := initHelm("testconfig", "testspace"); errors.Is(err, testerr) {
		t.Log("Error Expected")
	} else {
		t.Error("Expect", testerr, "but found", err)
	}
}

func TestYamlService(t *testing.T) {
	// mock functions
	func_cmdrun_ok := func(*exec.Cmd) error {
		return nil
	}
	func_cmdrun_fail := func(*exec.Cmd) error {
		return kubeerr
	}
	func_cmdrun_notfound := func(c *exec.Cmd) error {
		fmt.Fprintf(c.Stdout, "not found")
		return errNotFound
	}

	cases := []struct {
		name           string
		expectResult   string
		funcMockCMDRun func(*exec.Cmd) error
	}{
		{
			name:           "err_test_cmdrun_failed",
			expectResult:   "Unknown",
			funcMockCMDRun: func_cmdrun_fail,
		},
		{
			name:           "test_cmdrun_ok",
			expectResult:   "Deployed",
			funcMockCMDRun: func_cmdrun_ok,
		},
		{
			name:           "test_cmdrun_notfound",
			expectResult:   "Not Deployed",
			funcMockCMDRun: func_cmdrun_notfound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if tc.funcMockCMDRun != nil {
				cmd := exec.Command("ls")
				p, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(cmd), "Run", tc.funcMockCMDRun)
				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, p)
			}

			status := getYamlStatus("kubeconfig", "yamlfile")

			if status != tc.expectResult {
				t.Error("Expected", tc.expectResult, "but got", status)
			} else {
				t.Log("Done")
			}
		})
	}
}
