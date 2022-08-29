/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	epapiplugins "ep/pkg/api/plugins"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
)

var (
	errOSReadLink  = errors.New("osreadlink.error")
	errChDir       = errors.New("chdir.error")
	errMkDirAll    = errors.New("mkdirall.error")
	errCdWorkspace = errors.New("changetoworkspacepath.error")
	errHomeDir     = errors.New("userhomedir.error")
)

func utilsIsExpectedError(returnErr error, wantError error) bool {
	if !errors.Is(returnErr, wantError) &&
		(returnErr == nil || wantError == nil || !strings.Contains(returnErr.Error(), wantError.Error())) {
		return false
	}
	return true
}

func getparentwd() string {
	v, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	epPath = fmt.Sprintf("%s/..", v)
	return epPath
}

func getcwd() string {
	v, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return v
}

func patchosreadlink(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Readlink, func(name string) (string, error) {
		unpatch(t, patch)
		if ok {
			return "readlink.ok", nil
		} else {
			return "", errOSReadLink
		}
	})

	if patchErr != nil {
		t.Errorf("error patching %v", patchErr)
	}
}

func patchosgetenv(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Getenv, func(key string) string {
		unpatch(t, patch)
		if ok {
			return "getenv.ok"
		} else {
			return ""
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

}

func Test_Getworkspacepath(t *testing.T) {
	var cases = []struct {
		name         string
		expectError  error
		expectOutput string
		beforetest   func()
		teardown     func()
	}{
		{
			name:         "os.GetEnv ok",
			expectError:  nil,
			expectOutput: "getenv.ok",
			beforetest: func() {
				patchosgetenv(t, true)
			},
		},
		{
			name:         "os.readlink ok, os.GetEnv fail",
			expectError:  nil,
			expectOutput: getcwd(),
			beforetest: func() {
				patchosgetenv(t, false)
				patchosreadlink(t, true)
			},
		},
		{
			name:         "os.readlink fail, os.GetEnv fail",
			expectError:  nil,
			expectOutput: getparentwd(),
			beforetest: func() {
				patchosgetenv(t, false)
				patchosreadlink(t, false)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				workspacepath := GetWorkspacePath()

				if workspacepath != tc.expectOutput {
					t.Errorf("Output unexpected!")
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func patchoschdir(t *testing.T, ok bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Chdir, func(dir string) error {
		if ok {
			return nil
		} else {
			return errChDir
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func Test_ChangeToWorkspacePath(t *testing.T) {
	var patchHandler *mpatch.Patch
	var cases = []struct {
		name         string
		expectError  error
		expectOutput string
		beforetest   func()
		teardown     func()
	}{
		{
			name:        "EPPATH exist",
			expectError: nil,
			beforetest: func() {
				patchosgetenv(t, true)
			},
		},
		{
			name:        "os.readlink ok",
			expectError: nil,
			beforetest: func() {
				patchosgetenv(t, false)
				patchosreadlink(t, true)
			},
		},
		{
			name:        "os.readlink fail",
			expectError: nil,
			beforetest: func() {
				patchosgetenv(t, false)
				patchosreadlink(t, true)
				patchHandler = patchoschdir(t, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				err := ChangeToWorkspacePath()

				if !utilsIsExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}

}

func Test_MakeDir(t *testing.T) {
	var patchHandler *mpatch.Patch
	cases := []struct {
		name        string
		in_path     string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "os.MkdirAll fail",
			expectError: errMkDirAll,
			in_path:     "EPPATH",
			beforetest: func() {
				patchHandler = patchmkdirall(t, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:        "os.MkdirAll ok",
			expectError: nil,
			in_path:     "EPPATH",
			beforetest: func() {
				patchHandler = patchmkdirall(t, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				err := MakeDir(tc.in_path)

				if !utilsIsExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func patchmkdirall(t *testing.T, ok bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchMethod(os.MkdirAll, func(path string, perm fs.FileMode) error {
		if ok {
			return nil
		} else {
			return errMkDirAll
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func Test_Utils_Init(t *testing.T) {
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "Change to workspace path fail",
			expectError: errCdWorkspace,
			beforetest: func() {
				patchchangetoworkspacepath(t, false)
			},
		},
		{
			name:        "Utils_Init ok",
			expectError: nil,
			beforetest: func() {
				patchchangetoworkspacepath(t, true)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				err := Utils_Init()

				if !utilsIsExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func patchchangetoworkspacepath(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(ChangeToWorkspacePath, func() error {
		unpatch(t, patch)
		if ok {
			return nil
		} else {
			return errCdWorkspace

		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func TestGetRuntimeFolder(t *testing.T) {
	cases := []struct {
		name       string
		beforetest func()
		teardown   func()
	}{
		{
			name: "test getruntimefolder",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				rtfolder := GetRuntimeFolder()

				if len(rtfolder) <= 0 {
					t.Errorf("runtime folder is empty")
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func patchosstat(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Stat, func(name string) (fs.FileInfo, error) {
		unpatch(t, patch)
		if ok {
			return nil, nil
		} else {
			return nil, errStat
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchmakedir_utils(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(MakeDir, func(path string) error {
		unpatch(t, patch)
		if ok {
			return nil
		} else {
			return errMkDir
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func TestFileNameofRuntime(t *testing.T) {
	cases := []struct {
		name          string
		in_targetname string
		expectError   error
		beforetest    func()
		teardown      func()
	}{
		{
			name:          "Stat runtime folder fail, Makedir fail",
			in_targetname: "ep",
			expectError:   errMkDir,
			beforetest: func() {
				patchosstat(t, false)
				patchmakedir_utils(t, false)
			},
		},
		{
			name:          "Stat runtime folder fail",
			in_targetname: "ep",
			expectError:   nil,
			beforetest: func() {
				patchosstat(t, false)
			},
		},
		{
			name:          "Stat runtime folder ok",
			in_targetname: "ep",
			expectError:   nil,
			beforetest: func() {
				patchosstat(t, true)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				_, err := FileNameofRuntime(tc.in_targetname)

				if !utilsIsExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func TestGetDefaultTopConfigName(t *testing.T) {
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name: "GetDefaultTopConfigName",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				GetDefaultTopConfigName()

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func TestGetRuntimeTopConfig(t *testing.T) {
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name: "GetRuntimeTopConfig",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				GetRuntimeTopConfig(&epapiplugins.EpParams{})

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}

func TestGetHostDefaultIP(t *testing.T) {
	GetHostDefaultIP()
}

func patchuserhomedir_utils(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.UserHomeDir, func() (string, error) {
		unpatch(t, patch)
		if ok {
			return "userhomedir", nil
		} else {
			return "", errHomeDir
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func TestGetDefaultKubeConfig(t *testing.T) {
	cases := []struct {
		name         string
		expectOutput string
		beforetest   func()
		teardown     func()
	}{
		{
			name:         "UserHomeDir fail",
			expectOutput: "",
			beforetest: func() {
				patchuserhomedir_utils(t, false)
			},
		},
		{
			name:         "UserHomeDir ok",
			expectOutput: "userhomedir/.kube/config",
			beforetest: func() {
				patchuserhomedir_utils(t, true)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				kubeconfg := GetDefaultKubeConfig()

				if kubeconfg != tc.expectOutput {
					t.Errorf("Unexpected Output!")
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}
