/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//nolint: dupl
package conductorutils

import (
	papi "github.com/intel/edge-conductor/pkg/api/plugins"
	"github.com/intel/edge-conductor/pkg/eputils"

	"testing"
)

var test_cfg = papi.Customconfig{
	Resources: []*papi.CustomconfigResourcesItems0{
		{
			Name:  "resource1",
			Value: "value1",
		},
	},
}

func TestGetResourceValueFromCustomcfg(t *testing.T) {
	cases := []struct {
		testname       string
		inputname      string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			testname:  "success",
			inputname: "resource1",
			expectErr: false,
		},
		{
			testname:       "not found",
			inputname:      "nothing",
			expectErr:      true,
			expectedErrMsg: eputils.GetError("errResource").Error(),
		},
	}

	for n, tc := range cases {
		t.Logf("Case %d: %s start", n, tc.testname)
		func() {
			_, err := GetResourceValueFromCustomcfg(&test_cfg, tc.inputname)
			if tc.expectErr {
				if err == nil {
					t.Errorf("Expect \"%s\" occur but no error found.", tc.expectedErrMsg)
				} else if err.Error() != tc.expectedErrMsg {
					t.Errorf("Expect \"%s\" occur but found \"%s\".", tc.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error \"%s\" found.", err.Error())
				}
			}
		}()
		t.Logf("Case %d: %s end", n, tc.testname)
	}
	t.Log("Done")
}
