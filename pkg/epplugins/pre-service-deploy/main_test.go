/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package preservicedeploy

import (
	epplugins "ep/pkg/api/plugins"
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	"ep/pkg/executor"
	"errors"
	"strings"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
)

var (
	errEmpty = errors.New("")
)

/**
 * Test function PluginMain
 **/
func Test_PluginMain(t *testing.T) {
	var cases = []struct {
		name           string
		expectError    error
		in             eputils.SchemaMapData
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name: "Successfully Enabled SR-IOV",
			in: func() eputils.SchemaMapData {
				data := eputils.NewSchemaMapData()
				data[__name("ep-params")] = &pluginapi.EpParams{}
				return data
			}(),
			expectError: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				return nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := PluginMain(tc.in, nil)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

/**
 * Test function enable_sriov_vf
 **/
func Test_enable_sriov_vf(t *testing.T) {
	var cases = []struct {
		name           string
		expectError    error
		epparams       *epplugins.EpParams
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "success",
			expectError: nil,
			epparams: &epplugins.EpParams{
				Extensions: []*epplugins.EpParamsExtensionsItems0{{
					Extension: &epplugins.Extension{
						Extension: []*epplugins.ExtensionItems0{{
							Config: []*epplugins.ExtensionItems0ConfigItems0{{
								Name: "sriov_enabled",
							}},
						}},
					},
					Name: "sriov",
				}},
			},
			funcBeforeTest: func() []*mpatch.Patch {
				return nil
			},
		},
		{
			name:        "success, sriov already enabled",
			expectError: nil,
			epparams: &epplugins.EpParams{
				Extensions: []*epplugins.EpParamsExtensionsItems0{{
					Extension: &epplugins.Extension{
						Extension: []*epplugins.ExtensionItems0{{
							Config: []*epplugins.ExtensionItems0ConfigItems0{{
								Name:  "sriov_enabled",
								Value: "true",
							}},
						}},
					},
					Name: "sriov",
				}},
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchExecutorRun(t, false)
				return []*mpatch.Patch{patch1}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := enable_sriov_vf(tc.epparams)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func patchExecutorRun(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(executor.Run, func(specFile string, epparams *epplugins.EpParams, value interface{}) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}

	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func isExpectedError(returnErr error, wantError error) bool {
	if !errors.Is(returnErr, wantError) &&
		(returnErr == nil || wantError == nil || !strings.Contains(returnErr.Error(), wantError.Error())) {
		return false
	}
	return true
}

/**
 * unpatchAll
 * This function will remove all the monkey patches passed to it,
 * before using it you need to consider whether the patch slice you
 * passed into had already been unpatched, if you try to do this,
 * your code will panic.
 */
func unpatchAll(t *testing.T, pList []*mpatch.Patch) {
	for _, p := range pList {
		if p != nil {
			if err := p.Unpatch(); err != nil {
				t.Errorf("unpatch error: %v", err)
			}
		}
	}
}
