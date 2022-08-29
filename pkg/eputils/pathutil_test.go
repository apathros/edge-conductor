/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package eputils

import (
	"testing"
)

func TestGetBaseUrl(t *testing.T) {
	cases := []struct {
		name            string
		url             string
		baseUrlExpected string
		funcBeforeTest  func()
	}{
		{
			name:            "GetBaseUrl url parse error",
			url:             "http://ab_-|cde",
			baseUrlExpected: "",
		},
		{
			name:            "GetBaseUrl Username and password in the URL",
			url:             "http://user:pass@test.com/id/10",
			baseUrlExpected: "http://user:pass@test.com/id",
		},
		{
			name:            "GetBaseUrl no user info in the URL",
			url:             "http://test.com/id/10",
			baseUrlExpected: "http://test.com/id",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			output := GetBaseUrl(tc.url)
			if output != tc.baseUrlExpected {
				t.Errorf("Unexpected return value: %v, expected: %v", output, tc.baseUrlExpected)
			}
		})
	}
}
