/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package orasutils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	mock_orasutils "ep/pkg/eputils/orasutils/mock"

	"ep/pkg/eputils"
	oras "github.com/deislabs/oras/pkg/oras"
	types "github.com/docker/docker/api/types"
	gomock "github.com/golang/mock/gomock"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	log "github.com/sirupsen/logrus"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var testdatapath string

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestOrasPushFile(t *testing.T) {
	cases := []struct {
		name         string
		input        []string
		retOrasError error
		expectError  bool
	}{
		{
			"No_filename",
			[]string{"", "subref", "rev"},
			nil,
			true,
		},
		{
			"Invalid_filename_with_subref_rev",
			[]string{"filename", "subref", "rev"},
			nil,
			true,
		},
		{
			"Valid_filename_without_subref",
			[]string{filepath.Join(testdatapath, "orasfake.yml"), "", "rev"},
			nil,
			true,
		},
		{
			"Valid_filename_without_rev",
			[]string{filepath.Join(testdatapath, "orasfake.yml"), "subref", ""},
			nil,
			false,
		},
		{
			"Valid_filename_with_subref_rev",
			[]string{filepath.Join(testdatapath, "orasfake.yml"), "subref", "rev"},
			nil,
			false,
		},
		{
			"Oras_func_return_error",
			[]string{filepath.Join(testdatapath, "orasfake.yml"), "subref", "rev"},
			eputils.GetError("errOras"),
			true,
		},
	}

	authcfg := &types.AuthConfig{
		Username:      "test",
		Password:      "test123",
		ServerAddress: "10.10.10.10",
	}

	err := OrasNewClient(authcfg, "")
	if err != nil {
		t.Error(err)
		return
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockOrasInterface := mock_orasutils.NewMockOrasInterface(ctrl)
			patch, err := mpatch.PatchMethod(oras.Push, mockOrasInterface.Push)
			defer unpatch(t, patch)
			if err != nil {
				t.Fatal(err)
			}

			mockOrasInterface.EXPECT().Push(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(ocispec.Descriptor{}, tc.retOrasError)

			if _, result := OrasCli.OrasPushFile(tc.input[0], tc.input[1], tc.input[2]); result != nil {
				if tc.expectError {
					t.Log("Done")
					return
				} else {
					t.Logf("Test case %s failed.", tc.name)
					t.Error(result)
				}
			}
		})
	}
}

func TestOrasPullFile(t *testing.T) {
	cases := []struct {
		name         string
		input        []string
		retOrasError error
		expectError  bool
	}{
		{
			"No_filename",
			[]string{"", "regref"},
			nil,
			true,
		},
		{
			"Valid_filename_without_regref",
			[]string{filepath.Join(testdatapath, "oraspull.test"), ""},
			nil,
			true,
		},
		{
			"Valid_filename_with_invalid_regref",
			[]string{filepath.Join(testdatapath, "oraspull.test"), "regref"},
			nil,
			true,
		},
		{
			"Valid_filename_with_valid_regref",
			[]string{filepath.Join(testdatapath, "oraspull.test"), "oci://10.10.10.10/oraspull"},
			nil,
			false,
		},
		{
			"Oras_func_return_error",
			[]string{filepath.Join(testdatapath, "oraspull.test"), "oci://10.10.10.10/oraspull"},
			eputils.GetError("errOras"),
			true,
		},
	}

	authcfg := &types.AuthConfig{
		Username:      "test",
		Password:      "test123",
		ServerAddress: "10.10.10.10",
	}

	err := OrasNewClient(authcfg, "")
	if err != nil {
		t.Error(err)
		return
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockOrasInterface := mock_orasutils.NewMockOrasInterface(ctrl)
			patch, err := mpatch.PatchMethod(oras.Pull, mockOrasInterface.Pull)
			defer unpatch(t, patch)
			if err != nil {
				t.Fatal(err)
			}

			mockOrasInterface.EXPECT().Pull(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(ocispec.Descriptor{}, []ocispec.Descriptor{}, tc.retOrasError)

			if result := OrasCli.OrasPullFile(tc.input[0], tc.input[1]); result != nil {
				if tc.expectError {
					t.Log("Done")
					return
				} else {
					t.Logf("Test case %s failed.", tc.name)
					t.Error(result)
				}
			}
		})
	}
}

func TestMain(m *testing.M) {
	_, pwdpath, _, _ := runtime.Caller(0)
	testdatapath = filepath.Join(filepath.Dir(pwdpath), "testdata")
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}
