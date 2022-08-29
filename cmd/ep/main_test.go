/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package main

import (
	"ep/cmd/ep/app"
	"ep/pkg/eputils"
	"errors"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
)

var (
	errHash = errors.New("checkhash.error")
)

func patcheputilscheckhash(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchMethod(eputils.CheckHash, func(workspace string) error {
		unpatch(t, patch)
		if ok {
			return nil
		} else {
			return errHash
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchappexecute(t *testing.T) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchMethod(app.Execute, func() {})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_main(t *testing.T) {
	var handler *mpatch.Patch
	var cases = []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "Hash is valid in main function",
			expectError: nil,
			beforetest: func() {
				patcheputilscheckhash(t, true)
				handler = patchappexecute(t)
			},
			teardown: func() {
				unpatch(t, handler)
			},
		},
		{
			name:        "Hash is not valid in main function",
			expectError: errHash,
			beforetest: func() {
				patcheputilscheckhash(t, false)
				handler = patchappexecute(t)
			},
			teardown: func() {
				unpatch(t, handler)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				main()

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}

}
