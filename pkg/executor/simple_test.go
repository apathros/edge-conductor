/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package executor

import (
	"context"
	pluginapi "ep/pkg/api/plugins"
	"github.com/undefinedlabs/go-mpatch"
	"reflect"
	"testing"
)

func patchSetECParams(t *testing.T, fail bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(New()), "SetECParams", func(e *Executor, epparams *pluginapi.EpParams) error {
		if fail == true {
			return errExEmpty
		} else {
			return nil
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchLoadSpecFromFile(t *testing.T, fail bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(New()), "LoadSpecFromFile", func(e *Executor, specFile string) error {
		if fail == true {
			return errExEmpty
		} else {
			return nil
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchExeRun(t *testing.T) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(New()), "Run", func(e *Executor, ctx context.Context) error {
		return nil
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func TestSimpleShell(t *testing.T) {
	var cases = []struct {
		name           string
		ecparams       *pluginapi.EpParams
		tempvalue      *pluginapi.ExecSimpleShell
		spec           string
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "set temp value failed",
			expectError: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch_1 := patchSetECParams(t, false)
				patch_2 := patchLoadSpecFromFile(t, false)
				patch_3 := patchExeRun(t)
				patch_4 := patchSetTempValue(t, true)
				return []*mpatch.Patch{patch_1, patch_2, patch_3, patch_4}
			},
			ecparams:  &pluginapi.EpParams{},
			tempvalue: &pluginapi.ExecSimpleShell{},
		},
		{
			name:        "set temp file failed",
			expectError: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch_1 := patchSetECParams(t, false)
				patch_2 := patchLoadSpecFromFile(t, true)
				patch_3 := patchExeRun(t)
				return []*mpatch.Patch{patch_1, patch_2, patch_3}
			},
			ecparams:  &pluginapi.EpParams{},
			tempvalue: &pluginapi.ExecSimpleShell{},
		},
		{
			name:        "set ec param failed",
			spec:        "",
			expectError: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch_1 := patchSetECParams(t, true)
				patch_2 := patchLoadSpecFromFile(t, false)
				patch_3 := patchExeRun(t)
				return []*mpatch.Patch{patch_1, patch_2, patch_3}
			},
			ecparams:  &pluginapi.EpParams{},
			tempvalue: &pluginapi.ExecSimpleShell{},
		},
		{
			name:        "default",
			spec:        "",
			expectError: false,
			funcBeforeTest: func() []*mpatch.Patch {
				patch_1 := patchSetECParams(t, false)
				patch_2 := patchLoadSpecFromFile(t, false)
				patch_3 := patchExeRun(t)
				return []*mpatch.Patch{patch_1, patch_2, patch_3}
			},
			ecparams:  &pluginapi.EpParams{},
			tempvalue: &pluginapi.ExecSimpleShell{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := SimpleShell(tc.tempvalue, tc.ecparams)
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}
		})
	}
}

func patchSetTempValue(t *testing.T, fail bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(New()), "SetTempValue", func(e *Executor, value interface{}) error {
		if fail {
			return errExEmpty
		} else {
			return nil
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func TestRun(t *testing.T) {
	var cases = []struct {
		name           string
		specFile       string
		epparams       *pluginapi.EpParams
		value          interface{}
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "LoadSpecFromFile failed",
			expectError: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch_1 := patchSetECParams(t, false)
				patch_2 := patchLoadSpecFromFile(t, true)
				patch_3 := patchExeRun(t)
				patch_4 := patchSetTempValue(t, false)
				return []*mpatch.Patch{patch_1, patch_2, patch_3, patch_4}
			},
			epparams: &pluginapi.EpParams{},
		},
		{
			name:        "set ec param failed",
			expectError: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch_1 := patchSetECParams(t, true)
				patch_2 := patchLoadSpecFromFile(t, false)
				patch_3 := patchExeRun(t)
				patch_4 := patchSetTempValue(t, false)
				return []*mpatch.Patch{patch_1, patch_2, patch_3, patch_4}
			},
			epparams: &pluginapi.EpParams{},
		},
		{
			name:        "set temp value failed",
			expectError: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch_1 := patchSetECParams(t, false)
				patch_2 := patchLoadSpecFromFile(t, false)
				patch_3 := patchExeRun(t)
				patch_4 := patchSetTempValue(t, true)
				return []*mpatch.Patch{patch_1, patch_2, patch_3, patch_4}
			},
			epparams: &pluginapi.EpParams{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := Run(tc.specFile, tc.epparams, tc.value)
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}
		})
	}
}
