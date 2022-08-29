/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	"ep/pkg/eputils"
	"errors"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
)

var (
	errCpFile = errors.New("copyfile.error")
	errRmAll  = errors.New("removeall.error")
	errMkdir  = errors.New("makedir.error")
)

func patchfileexists(t *testing.T, yes bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(eputils.FileExists, func(filename string) bool {
		unpatch(t, patch)
		if yes {
			return true
		} else {
			return false
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchcopyfile(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(eputils.CopyFile, func(dstName, srcName string) (written int64, err error) {
		unpatch(t, patch)
		if ok {
			return 64, nil
		} else {
			return 0, errCpFile
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func Test_copyCaRuntimeDataDir(t *testing.T) {
	var patchHandler *mpatch.Patch
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "Failed to remove old CA",
			expectError: errRmAll,
			beforetest: func() {
				patchHandler = patchRemoveAll(t, errRmAll, nil)
				patchfileexists(t, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:        "Failed to create reg folder",
			expectError: errMkdir,
			beforetest: func() {
				patchHandler = patchMakeDir(t, errMkdir)
				patchfileexists(t, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:        "Failed to copy root CA files",
			expectError: errCpFile,
			beforetest: func() {
				patchHandler = patchMakeDir(t, nil)
				patchfileexists(t, false)
				patchcopyfile(t, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:        "Copy CA ok",
			expectError: nil,
			beforetest: func() {
				patchHandler = patchMakeDir(t, nil)
				patchfileexists(t, false)
				patchcopyfile(t, true)
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

				err := copyCaRuntimeDataDir("mockreg", "mockworkspace", "mockrtdir", "mockcertpath")
				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}
}
