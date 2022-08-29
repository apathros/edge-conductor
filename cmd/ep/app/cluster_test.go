/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//nolint: dupl
package app

import (
	epapiplugins "ep/pkg/api/plugins"
	"errors"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
)

var (
	errClusterCmd = errors.New("checkclustercmd.error")
	errPreinit    = errors.New("epwfpreinit.error")
	errStart      = errors.New("epwfstart.error")
	errStat       = errors.New("stat.error")
)

func Test_check_cluster_cmd(t *testing.T) {
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "cluster kubeconfig does not exist",
			expectError: errStat,
			beforetest: func() {
				patchisnotexist(t, true)
				patchosstat(t, false)
			},
		},

		{
			name:        "check_cluster_cmd ok",
			expectError: nil,
			beforetest: func() {
				patchisnotexist(t, false)
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

				err := check_cluster_cmd()

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

func patchepwfpreinit(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(EpWfPreInit, func(epPms *epapiplugins.EpParams, p map[string]string) (*epapiplugins.EpParams, error) {
		unpatch(t, patch)
		if ok {
			return nil, nil
		} else {
			return nil, errPreinit
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchepwfstart(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(EpWfStart, func(epParams *epapiplugins.EpParams, name string) error {
		unpatch(t, patch)
		if ok {
			return nil
		} else {
			return errStart
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchcheckclustercmd(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(check_cluster_cmd, func() error {
		unpatch(t, patch)
		if ok {
			return nil
		} else {
			return errClusterCmd
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func Test_deployClusterCmd(t *testing.T) {

	cases := []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "Deploy cluster cmd ok",
			expectError: nil,
			beforetest: func() {
				patchepwfpreinit(t, true)
				patchepwfstart(t, true)
			},
		},
		{
			name:        "epwfpreinit fail",
			expectError: errPreinit,
			beforetest: func() {
				patchepwfpreinit(t, false)
			},
		},
		{
			name:        "epwfstart fail",
			expectError: errStart,
			beforetest: func() {
				patchepwfpreinit(t, true)
				patchepwfstart(t, false)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				err := deployClusterCmd.RunE(nil, nil)

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

func Test_BulidClusterCMD(t *testing.T) {
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "Build cluster cmd ok",
			expectError: nil,
			beforetest: func() {
				patchepwfpreinit(t, true)
				patchepwfstart(t, true)
			},
		},
		{
			name:        "epwfpreinit fail",
			expectError: errPreinit,
			beforetest: func() {
				patchepwfpreinit(t, false)
			},
		},
		{
			name:        "epwfstart fail",
			expectError: errStart,
			beforetest: func() {
				patchepwfpreinit(t, true)
				patchepwfstart(t, false)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				err := buildClusterCmd.RunE(nil, nil)

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

func Test_GetClusterInfoCMD(t *testing.T) {
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "Get cluster info cmd ok",
			expectError: nil,
			beforetest: func() {
				patchepwfpreinit(t, true)
				patchepwfstart(t, true)
			},
		},
		{
			name:        "epwfpreinit fail",
			expectError: errPreinit,
			beforetest: func() {
				patchepwfpreinit(t, false)
			},
		},
		{
			name:        "epwfstart fail",
			expectError: errStart,
			beforetest: func() {
				patchepwfpreinit(t, true)
				patchepwfstart(t, false)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				err := getClusterInfoCmd.RunE(nil, nil)

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

func Test_JoinClusterCMD(t *testing.T) {
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "check cluster cmd fail",
			expectError: errClusterCmd,
			beforetest: func() {
				patchcheckclustercmd(t, false)
			},
		},
		{
			name:        "join cluster cmd ok",
			expectError: nil,
			beforetest: func() {
				patchepwfpreinit(t, true)
				patchepwfstart(t, true)
				patchcheckclustercmd(t, true)
			},
		},
		{
			name:        "epwfpreinit fail",
			expectError: errPreinit,
			beforetest: func() {
				patchepwfpreinit(t, false)
				patchcheckclustercmd(t, true)
			},
		},
		{
			name:        "epwfstart fail",
			expectError: errStart,
			beforetest: func() {
				patchepwfpreinit(t, true)
				patchepwfstart(t, false)
				patchcheckclustercmd(t, true)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				err := joinClusterCmd.RunE(nil, nil)

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
