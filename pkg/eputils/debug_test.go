/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package eputils

import (
	"testing"

	mpatch "github.com/undefinedlabs/go-mpatch"
	"sigs.k8s.io/yaml"
)

func TestPPrint(t *testing.T) {
	PPrint("hello world!")
}

func TestD(t *testing.T) {
	D("hello world!")
}

func patchmarshal(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(yaml.Marshal, func(o interface{}) ([]byte, error) {
		unpatch(t, patch)
		if ok {
			return []byte{}, nil
		} else {
			return nil, GetError("errMarshal")
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func TestDumpVar(t *testing.T) {
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
	}{
		{
			name:        "Marshal failed",
			expectError: nil,
			beforetest: func() {
				patchmarshal(t, false)
			},
		},
		{
			name:        "Marshal ok",
			expectError: nil,
			beforetest: func() {
				patchmarshal(t, true)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}
				DumpVar("hello world!")
			}
		})
	}
}
