/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	"ep/pkg/api/plugins"
	"os"
	"testing"

	mpatch "github.com/undefinedlabs/go-mpatch"
)

func patchRemoveAll(t *testing.T, err error, patchNext func()) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(os.RemoveAll, func(_ string) error {
		if patchNext != nil {
			patchNext()
		}
		return err
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchRemoveAllOnce(t *testing.T, err error, patchNext func()) {
	var patch *mpatch.Patch
	patch, _ = mpatch.PatchMethod(os.RemoveAll, func(_ string) error {
		unpatch(t, patch)
		if patchNext != nil {
			patchNext()
		}
		return err
	})
}

func TestEpDeinit(t *testing.T) {
	isFunctionCorrectlyFunc := func(err, wantError error) {
		if !isWantedError(err, wantError) {
			t.Errorf("Unexpected error: %v", err)
		}
	}
	cases := []struct {
		funcBeforeTest      func() []*mpatch.Patch
		wantError           error
		isFunctionCorrectly func(err, wantError error)
	}{
		{
			wantError: testError,
			funcBeforeTest: func() []*mpatch.Patch {
				patch := patchEpWfPreInit(t, nil, testError)
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: testError,
			funcBeforeTest: func() []*mpatch.Patch {
				patchEpWfPreInit := patchEpWfPreInit(t, &plugins.EpParams{}, nil)
				patchEpWfStart := patchEpWfStart(t, testError)
				return []*mpatch.Patch{patchEpWfPreInit, patchEpWfStart}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: testError,
			funcBeforeTest: func() []*mpatch.Patch {
				patchEpWfPreInit := patchEpWfPreInit(t, &plugins.EpParams{}, nil)
				patchEpWfStart := patchEpWfStart(t, nil)
				patchRemoveAllOnce(t, testError, nil)
				return []*mpatch.Patch{patchEpWfPreInit, patchEpWfStart}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: testError,
			funcBeforeTest: func() []*mpatch.Patch {
				purge = true
				patchEpWfPreInit := patchEpWfPreInit(t, &plugins.EpParams{}, nil)
				patchEpWfStart := patchEpWfStart(t, nil)
				patchRemoveAllOnce(t, nil, func() {
					patchRemoveAllOnce(t, testError, nil)
				})
				return []*mpatch.Patch{patchEpWfPreInit, patchEpWfStart}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: testError,
			funcBeforeTest: func() []*mpatch.Patch {
				purge = true
				patchEpWfPreInit := patchEpWfPreInit(t, &plugins.EpParams{}, nil)
				patchEpWfStart := patchEpWfStart(t, nil)
				patchRemoveAllOnce(t, nil, func() {
					patchRemoveAllOnce(t, nil, func() {
						patchRemoveAllOnce(t, testError, nil)
					})
				})
				return []*mpatch.Patch{patchEpWfPreInit, patchEpWfStart}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				purge = true
				patchEpWfPreInit := patchEpWfPreInit(t, &plugins.EpParams{}, nil)
				patchEpWfStart := patchEpWfStart(t, nil)
				patchRemoveAllOnce(t, nil, func() {
					patchRemoveAllOnce(t, nil, func() {
						patchRemoveAllOnce(t, nil, nil)
					})
				})
				return []*mpatch.Patch{patchEpWfPreInit, patchEpWfStart}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
	}

	for n, testCase := range cases {
		t.Logf("%s case %d start", getFuncName(), n)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}
			err := ep_deinit()
			testCase.isFunctionCorrectly(err, testCase.wantError)
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}
	t.Log("Done")
}

func TestEpDeinitCmd(t *testing.T) {
	isFunctionCorrectlyFunc := func(err, wantError error) {
		if !isWantedError(err, wantError) {
			t.Errorf("Unexpected error: %v", err)
		}
	}
	cases := []struct {
		funcBeforeTest      func() []*mpatch.Patch
		wantError           error
		isFunctionCorrectly func(err, wantError error)
	}{
		{
			wantError: testError,
			funcBeforeTest: func() []*mpatch.Patch {
				patch := patchEpWfPreInit(t, nil, testError)
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				purge = true
				patchEpWfPreInit := patchEpWfPreInit(t, &plugins.EpParams{}, nil)
				patchEpWfStart := patchEpWfStart(t, nil)
				patchRemoveAllOnce(t, nil, func() {
					patchRemoveAllOnce(t, nil, func() {
						patchRemoveAllOnce(t, nil, nil)
					})
				})
				return []*mpatch.Patch{patchEpWfPreInit, patchEpWfStart}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
	}

	for n, testCase := range cases {
		t.Logf("%s case %d start", getFuncName(), n)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}
			err := deinitCmd.RunE(nil, nil)
			testCase.isFunctionCorrectly(err, testCase.wantError)
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}
	t.Log("Done")
}
