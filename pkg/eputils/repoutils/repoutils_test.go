/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package repoutils

import (
	"errors"
	"github.com/intel/edge-conductor/pkg/eputils"
	"github.com/intel/edge-conductor/pkg/eputils/orasutils"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/undefinedlabs/go-mpatch"
)

var (
	errOrasPull = errors.New("oraspullfile.error")
	errOrasPush = errors.New("oraspushfile.error")
	errParsePrt = errors.New("parse \"http://example.com:123abc/foo\": invalid port \":123abc\" after host")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func isExpectedError(returnErr error, wantError error) bool {
	if !errors.Is(returnErr, wantError) &&
		(returnErr == nil || wantError == nil || !strings.Contains(returnErr.Error(), wantError.Error())) {
		return false
	}
	return true
}

func initorascli() {
	_ = orasutils.OrasNewClient(&types.AuthConfig{
		Username:      "test",
		Password:      "test123",
		ServerAddress: "10.10.10.10",
	}, "")
}

func patchoraspullfilefail(t *testing.T) {
	initorascli()
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(orasutils.OrasCli), "OrasPullFile", func(orasClient *orasutils.OrasClient, targetfile string, regref string) error {
		unpatch(t, patch)
		return errOrasPull
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchoraspullfileok(t *testing.T) {
	initorascli()
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(orasutils.OrasCli), "OrasPullFile", func(orasClient *orasutils.OrasClient, targetfile string, regref string) error {
		unpatch(t, patch)
		return nil
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchoraspushfilefail(t *testing.T) {
	initorascli()
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(orasutils.OrasCli), "OrasPushFile", func(orasClient *orasutils.OrasClient, filename, subRef, rev string) (string, error) {
		unpatch(t, patch)
		return "", errOrasPush
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchoraspushfileok(t *testing.T) {
	initorascli()
	var patch *mpatch.Patch
	var patchErr error

	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(orasutils.OrasCli), "OrasPushFile", func(orasClient *orasutils.OrasClient, filename, subRef, rev string) (string, error) {
		unpatch(t, patch)
		return "", nil
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func TestPushFileToRepo(t *testing.T) {
	cases := []struct {
		name        string
		in_filepath string
		in_subref   string
		in_rev      string
		expectError error
		beforetest  func()
		teardown    func()
	}{
		{
			name:        "client is not available",
			in_filepath: filepath.Join("testdata", "repomock.yml"),
			in_subref:   "subref",
			in_rev:      "1.0.0",
			expectError: eputils.GetError("errNoPushClient"),
			beforetest: func() {
				orasutils.OrasCli = nil
			},
		},
		{
			name:        "client ok, oraspushfile ok",
			in_filepath: filepath.Join("testdata", "repomock.yml"),
			in_subref:   "subref",
			in_rev:      "1.0.0",
			expectError: nil,
			beforetest: func() {
				patchoraspushfileok(t)
			},
		},
		{
			name:        "client ok, oraspushfile fail",
			in_filepath: filepath.Join("testdata", "repomock.yml"),
			in_subref:   "subref",
			in_rev:      "1.0.0",
			expectError: errOrasPush,
			beforetest: func() {
				patchoraspushfilefail(t)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()

				}

				_, err := PushFileToRepo(tc.in_filepath, tc.in_subref, tc.in_rev)

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}

}

func TestPullFileFromRepo(t *testing.T) {
	cases := []struct {
		name         string
		in_filepath  string
		in_targeturl string
		expectError  error
		beforetest   func()
		teardown     func()
	}{
		{
			name:         "client is not available",
			in_filepath:  filepath.Join("testdata", "repomock2.yml"),
			in_targeturl: "oci://123.123.123.123/mockurl",
			expectError:  eputils.GetError("errNoPullClient"),
			beforetest: func() {
				orasutils.OrasCli = nil
			},
		},
		{
			name:         "client ok, oraspullfile fail",
			in_filepath:  filepath.Join("testdata", "repomock2.yml"),
			in_targeturl: "oci://123.123.123.123/mockurl",
			expectError:  nil,
			beforetest: func() {
				patchoraspullfileok(t)
			},
		},
		{
			name:         "client ok, oraspullfile fail",
			in_filepath:  filepath.Join("testdata", "repomock.yml"),
			in_targeturl: "oci://123.123.123.123/mockurl",
			expectError:  errOrasPull,
			beforetest: func() {
				patchoraspullfilefail(t)
			},
		},
		{
			name:         "url illegal",
			in_filepath:  filepath.Join("testdata", "repomock.yml"),
			in_targeturl: "http://example.com:123abc/foo",
			expectError:  errParsePrt,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				err := PullFileFromRepo(tc.in_filepath, tc.in_targeturl)

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

				if tc.teardown != nil {
					tc.teardown()
				}

			}

		})
	}

}
