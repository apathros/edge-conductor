/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//nolint: dupl
package executor

import (
	"context"
	"io"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	mpatch "github.com/undefinedlabs/go-mpatch"
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func unpatchAll(t *testing.T, pList []*mpatch.Patch) {
	for _, p := range pList {
		if p != nil {
			if err := p.Unpatch(); err != nil {
				t.Errorf("unpatch error: %v", err)
			}
		}
	}
}

func Test_Connect(t *testing.T) {
	cases := []struct {
		name           string
		day0Client     *day0Client
		expectError    bool
		funcBeforeTest func()
	}{
		{
			name:        "Establish a connection",
			day0Client:  &day0Client{},
			expectError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			result := tc.day0Client.Connect()

			if (result != nil && !tc.expectError) ||
				(result == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(result)
			} else {
				t.Log("Done")
			}
		})
	}
}

func Test_Disconnect(t *testing.T) {
	cases := []struct {
		name           string
		day0Client     *day0Client
		expectError    bool
		funcBeforeTest func()
	}{
		{
			name:        "Disconnect client",
			day0Client:  &day0Client{},
			expectError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			result := tc.day0Client.Disconnect()

			if (result != nil && !tc.expectError) ||
				(result == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(result)
			} else {
				t.Log("Done")
			}
		})
	}

}

type MockReader struct {
	io.Reader
}

func (w *MockReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

type MockWriter struct {
	io.Writer
}

func (w *MockWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}

func patchCommandContext(t *testing.T) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(exec.CommandContext, func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return &exec.Cmd{}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchRun(t *testing.T) *mpatch.Patch {

	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&exec.Cmd{}), "Run", func(c *exec.Cmd) error {
		return nil
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func Test_CmdWithAttachIO(t *testing.T) {

	cases := []struct {
		name           string
		day0Client     *day0Client
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "Simulate executing the command line over a connection",
			day0Client:  &day0Client{},
			expectError: false,
			funcBeforeTest: func() []*mpatch.Patch {
				patch_cmdctx := patchCommandContext(t)
				patch_run := patchRun(t)
				return []*mpatch.Patch{patch_cmdctx, patch_run}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			result := tc.day0Client.CmdWithAttachIO(context.TODO(), []string{"watch", "-n", "/dev/null"}, strings.NewReader("stdin"), &MockWriter{}, &MockWriter{}, false)

			if (result != nil && !tc.expectError) ||
				(result == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(result)
			} else {
				t.Log("Done")
			}
		})
	}
}
