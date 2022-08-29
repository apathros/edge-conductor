/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package eputils

import (
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	errNoExecutable = fmt.Errorf("executable file not found")
)

func isExpectedError(returnErr error, wantError error) bool {
	if !errors.Is(returnErr, wantError) &&
		(returnErr == nil || wantError == nil || !strings.Contains(returnErr.Error(), wantError.Error())) {
		return false
	}
	return true
}

func patchRunCMDEx(t *testing.T, output string, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(RunCMDEx, func(cmd *exec.Cmd, useOsOut bool) (string, error) {
		unpatch(t, patch)
		return output, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchCmdRun(t *testing.T, stdoutStr, stderrStr string, errExpected error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(&exec.Cmd{}), "Run", func(cmd *exec.Cmd) error {
		_, err := cmd.Stdout.Write([]byte(stdoutStr))
		if err != nil {
			t.Log("Error writing to StdOut :", err)
		}
		_, errStr := cmd.Stderr.Write([]byte(stderrStr))
		if errStr != nil {
			t.Log("Error writing to StdErr :", errStr)
		}
		unpatch(t, patch)
		return errExpected
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func TestRunCMD(t *testing.T) {
	cases := []struct {
		name           string
		expectedError  error
		expectedOutput string
		funcBeforeTest func()
	}{
		{
			name: "RunCMD RunCMDEx return error",
			funcBeforeTest: func() {
				patchRunCMDEx(t, "", testErr)
			},
			expectedError:  testErr,
			expectedOutput: "",
		},
		{
			name: "RunCMD RunCMDEx return ok",
			funcBeforeTest: func() {
				patchRunCMDEx(t, "test output", nil)
			},
			expectedError:  nil,
			expectedOutput: "test output",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			output, err := RunCMD(nil)
			if !isExpectedError(err, tc.expectedError) {
				t.Errorf("Unexpected error: %v", err)
			}
			if output != tc.expectedOutput {
				t.Errorf("Unexpected return value: %v, expected: %v", output, tc.expectedOutput)
			}
		})
	}
}

func TestRunCMDEx(t *testing.T) {
	cases := []struct {
		name           string
		cmd            *exec.Cmd
		useOsOut       bool
		expectedError  error
		expectedOutput string
		funcBeforeTest func()
	}{
		{
			name:           "RunCMDEx use os output ok",
			expectedError:  nil,
			expectedOutput: "",
			funcBeforeTest: func() {
				patchCmdRun(t, "", "", nil)
			},
			useOsOut: true,
			cmd:      exec.Command("test"),
		},
		{
			name:           "RunCMDEx use os output error",
			expectedError:  errNoExecutable,
			expectedOutput: "",
			useOsOut:       true,
			cmd:            exec.Command("asdfasdf"),
		},
		{
			name:           "RunCMDEx use buffer output ok",
			expectedError:  nil,
			expectedOutput: "test stdout",
			useOsOut:       false,
			funcBeforeTest: func() {
				patchCmdRun(t, "test stdout", "test stderr", nil)
			},
			cmd: exec.Command("test"),
		},
		{
			name:           "RunCMDEx use buffer output error",
			expectedError:  errGeneral,
			expectedOutput: "",
			useOsOut:       false,
			funcBeforeTest: func() {
				patchCmdRun(t, "test stdout", "test stderr", testErr)
			},
			cmd: exec.Command("test"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			output, err := RunCMDEx(tc.cmd, tc.useOsOut)
			if !isExpectedError(err, tc.expectedError) {
				t.Errorf("Unexpected error: %v", err)
			}
			if output != tc.expectedOutput {
				t.Errorf("Unexpected return value: %v, expected: %v", output, tc.expectedOutput)
			}
		})
	}
}
