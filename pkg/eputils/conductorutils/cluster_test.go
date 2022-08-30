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

var test_provider = papi.ClustermanifestClusterProvidersItems0{
	Name: "test1",
	Binaries: []*papi.ClustermanifestClusterProvidersItems0BinariesItems0{
		{
			Name:   "binary1",
			URL:    "http://url/binary1",
			Sha256: "123456",
		},
	},
	Images: []*papi.ClustermanifestClusterProvidersItems0ImagesItems0{
		{
			Name:    "image1",
			RepoTag: "image1:123456",
		},
		{
			Name:    "image2",
			RepoTag: "image2:123456",
		},
	},
	Resources: []*papi.ClustermanifestClusterProvidersItems0ResourcesItems0{
		{
			Name:  "resource1",
			Value: "value1",
		},
	},
}

var test_cluster_manifest = papi.Clustermanifest{
	ClusterProviders: []*papi.ClustermanifestClusterProvidersItems0{
		&test_provider,
	},
}

func TestGetClusterManifest(t *testing.T) {
	cases := []struct {
		testname       string
		inputname      string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			testname:  "success",
			inputname: "test1",
			expectErr: false,
		},
		{
			testname:       "not found",
			inputname:      "no_cluster",
			expectErr:      true,
			expectedErrMsg: eputils.GetError("errManifest").Error(),
		},
	}

	for n, tc := range cases {
		t.Logf("Case %d: %s start", n, tc.testname)
		func() {
			_, err := GetClusterManifest(&test_cluster_manifest, tc.inputname)
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

func TestGetImageFromProvider(t *testing.T) {
	cases := []struct {
		testname       string
		inputname      string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			testname:  "success",
			inputname: "image1",
			expectErr: false,
		},
		{
			testname:       "not found",
			inputname:      "nothing",
			expectErr:      true,
			expectedErrMsg: eputils.GetError("errImage").Error(),
		},
	}

	for n, tc := range cases {
		t.Logf("Case %d: %s start", n, tc.testname)
		func() {
			_, err := GetImageFromProvider(&test_provider, tc.inputname)
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

func TestGetImageListFromProvider(t *testing.T) {
	cases := []struct {
		testname       string
		expectImageLen int
	}{
		{
			testname:       "success",
			expectImageLen: 2,
		},
	}

	for n, tc := range cases {
		t.Logf("Case %d: %s start", n, tc.testname)
		func() {
			imagelist := GetImageListFromProvider(&test_provider)
			if len(imagelist) != tc.expectImageLen {
				t.Errorf("Wrong image list %s", imagelist)
			}
		}()
		t.Logf("Case %d: %s end", n, tc.testname)
	}
	t.Log("Done")
}

func TestGetBinaryFromProvider(t *testing.T) {
	cases := []struct {
		testname       string
		inputname      string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			testname:  "success",
			inputname: "binary1",
			expectErr: false,
		},
		{
			testname:       "not found",
			inputname:      "nothing",
			expectErr:      true,
			expectedErrMsg: eputils.GetError("errBinary").Error(),
		},
	}

	for n, tc := range cases {
		t.Logf("Case %d: %s start", n, tc.testname)
		func() {
			_, _, err := GetBinaryFromProvider(&test_provider, tc.inputname)
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

func TestGetResourceValueFromProvider(t *testing.T) {
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
			_, err := GetResourceValueFromProvider(&test_provider, tc.inputname)
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
