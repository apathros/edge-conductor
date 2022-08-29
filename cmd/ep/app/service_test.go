/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
//nolint: dupl
package app

import (
	"os"
	"testing"

	mpatch "github.com/undefinedlabs/go-mpatch"
)

func patchCheckServiceCmd(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(check_service_cmd, func() error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func TestCheckServiceCmd(t *testing.T) {
	cases := []struct {
		funcBeforeTest      func() []*mpatch.Patch
		isFunctionCorrectly func(err error)
	}{
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patch := patchStat(t, nil, nil)
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, nil) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patch := patchStat(t, nil, os.ErrNotExist)
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, os.ErrNotExist) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
	}

	for n, testCase := range cases {
		t.Logf("%s case %d start", getFuncName(), n)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}
			testCase.isFunctionCorrectly(check_service_cmd())
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}

	t.Log("Done")
}

func TestBuildServiceCmd(t *testing.T) {
	cases := []struct {
		funcBeforeTest      func() []*mpatch.Patch
		isFunctionCorrectly func(err error)
	}{
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patch := patchEpWfPreInit(t, nil, testError)
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patchEpWfPreInit := patchEpWfPreInit(t, nil, nil)
				patchEpwfLoadServices := patchEpWfLoadServices(t, testError)
				return []*mpatch.Patch{patchEpWfPreInit, patchEpwfLoadServices}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patchEpWfPreInit := patchEpWfPreInit(t, nil, nil)
				patchEpwfLoadServices := patchEpWfLoadServices(t, nil)
				patchEpWfTearDown := patchEpWfTearDown(t, nil)
				patchEpWfStart := patchEpWfStart(t, testError)
				return []*mpatch.Patch{patchEpWfPreInit, patchEpwfLoadServices, patchEpWfTearDown, patchEpWfStart}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patchEpWfPreInit := patchEpWfPreInit(t, nil, nil)
				patchEpwfLoadServices := patchEpWfLoadServices(t, nil)
				patchEpWfStart := patchEpWfStart(t, nil)
				patchEpWfTearDown := patchEpWfTearDown(t, nil)
				return []*mpatch.Patch{patchEpWfPreInit, patchEpwfLoadServices, patchEpWfTearDown, patchEpWfStart}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, nil) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
	}

	for n, testCase := range cases {
		t.Logf("%s case %d start", getFuncName(), n)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}
			testCase.isFunctionCorrectly(buildServiceCmd.RunE(nil, nil))
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}

	t.Log("Done")
}

func TestDeployServiceCmd(t *testing.T) {
	cases := []struct {
		funcBeforeTest      func() []*mpatch.Patch
		isFunctionCorrectly func(err error)
	}{
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patch := patchCheckServiceCmd(t, testError)
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patchCheckServiceCmd := patchCheckServiceCmd(t, nil)
				patchEpWfPreInit := patchEpWfPreInit(t, nil, testError)
				return []*mpatch.Patch{patchCheckServiceCmd, patchEpWfPreInit}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patchCheckServiceCmd := patchCheckServiceCmd(t, nil)
				patchEpWfPreInit := patchEpWfPreInit(t, nil, nil)
				patchEpWfStart := patchEpWfStart(t, testError)
				return []*mpatch.Patch{patchCheckServiceCmd, patchEpWfPreInit, patchEpWfStart}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patchCheckServiceCmd := patchCheckServiceCmd(t, nil)
				patchEpWfPreInit := patchEpWfPreInit(t, nil, nil)
				patchEpWfStart := patchEpWfStart(t, nil)
				return []*mpatch.Patch{patchCheckServiceCmd, patchEpWfPreInit, patchEpWfStart}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, nil) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
	}

	for n, testCase := range cases {
		t.Logf("%s case %d start", getFuncName(), n)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}
			testCase.isFunctionCorrectly(deployServiceCmd.RunE(nil, nil))
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}

	t.Log("Done")
}

func TestListServiceCmd(t *testing.T) {
	cases := []struct {
		funcBeforeTest      func() []*mpatch.Patch
		isFunctionCorrectly func(err error)
	}{
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patch := patchCheckServiceCmd(t, testError)
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patchCheckServiceCmd := patchCheckServiceCmd(t, nil)
				patchEpWfPreInit := patchEpWfPreInit(t, nil, testError)
				return []*mpatch.Patch{patchCheckServiceCmd, patchEpWfPreInit}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patchCheckServiceCmd := patchCheckServiceCmd(t, nil)
				patchEpWfPreInit := patchEpWfPreInit(t, nil, nil)
				patchEpWfStart := patchEpWfStart(t, testError)
				return []*mpatch.Patch{patchCheckServiceCmd, patchEpWfPreInit, patchEpWfStart}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patchCheckServiceCmd := patchCheckServiceCmd(t, nil)
				patchEpWfPreInit := patchEpWfPreInit(t, nil, nil)
				patchEpWfStart := patchEpWfStart(t, nil)
				return []*mpatch.Patch{patchCheckServiceCmd, patchEpWfPreInit, patchEpWfStart}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, nil) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
	}

	for n, testCase := range cases {
		t.Logf("%s case %d start", getFuncName(), n)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}
			testCase.isFunctionCorrectly(listServiceCmd.RunE(nil, nil))
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}

	t.Log("Done")
}
