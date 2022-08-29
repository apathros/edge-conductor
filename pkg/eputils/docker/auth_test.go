/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package docker

import (
	eputils "ep/pkg/eputils"
	"errors"
	"github.com/docker/docker/api/types"
	mpatch "github.com/undefinedlabs/go-mpatch"
	"io/ioutil"
	"testing"
)

var (
	errFakeFunc = errors.New("fake func return error.")
)

func TestLoadDockerCliCredentials(t *testing.T) {
	func_with_login := func() []*mpatch.Patch {
		patch, _ := mpatch.PatchMethod(getCliConfigFilePath, func() string { return "testdata/auth_login/config.json" })
		return []*mpatch.Patch{patch}
	}

	func_with_logout := func() []*mpatch.Patch {
		patch, _ := mpatch.PatchMethod(getCliConfigFilePath, func() string { return "testdata/auth_logout/config.json" })
		return []*mpatch.Patch{patch}
	}

	func_fail_read_config_file := func() []*mpatch.Patch {
		patch1, _ := mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) {
			_ = filename
			return []byte(""), errFakeFunc
		})
		patch2, _ := mpatch.PatchMethod(eputils.FileExists, func(filename string) bool { _ = filename; return true })
		return []*mpatch.Patch{patch1, patch2}
	}

	func_with_invalid_config_format := func() []*mpatch.Patch {
		content := `{
	"auths": {
		}
	}
}`
		patch1, _ := mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) { _ = filename; return []byte(content), nil })
		patch2, _ := mpatch.PatchMethod(eputils.FileExists, func(filename string) bool { _ = filename; return true })
		return []*mpatch.Patch{patch1, patch2}
	}

	func_fail_decode_authstr := func() []*mpatch.Patch {
		content := `{
	"auths": {
		"https://index.docker.io/v1/": {
			"auth": "dGDp0ZXN0Cg=="
		}
	},
	"HttpHeaders": {
                "User-Agent": "Docker-Client/18.06.1-ce (linux)"
        }
}`
		patch1, _ := mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) { _ = filename; return []byte(content), nil })
		patch2, _ := mpatch.PatchMethod(eputils.FileExists, func(filename string) bool { _ = filename; return true })
		return []*mpatch.Patch{patch1, patch2}

	}

	func_with_invalid_authstr := func() []*mpatch.Patch {
		content := `{
        "auths": {
                "https://index.docker.io/v1/": {
                        "auth": "dGVzdHRlc3QK"
                }
        },
        "HttpHeaders": {
                "User-Agent": "Docker-Client/18.06.1-ce (linux)"
        }
}`
		patch1, _ := mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) { _ = filename; return []byte(content), nil })
		patch2, _ := mpatch.PatchMethod(eputils.FileExists, func(filename string) bool { _ = filename; return true })
		return []*mpatch.Patch{patch1, patch2}
	}

	func_password_include_colon := func() []*mpatch.Patch {
		content := `{
        "auths": {
                "https://index.docker.io/v1/": {
                        "auth": "dGVzdDpUZXN0OnRlc3Q6dGVzdA=="
                }
        },
        "HttpHeaders": {
                "User-Agent": "Docker-Client/18.06.1-ce (linux)"
        }
}`
		patch1, _ := mpatch.PatchMethod(ioutil.ReadFile, func(filename string) ([]byte, error) { _ = filename; return []byte(content), nil })
		patch2, _ := mpatch.PatchMethod(eputils.FileExists, func(filename string) bool { _ = filename; return true })
		return []*mpatch.Patch{patch1, patch2}
	}

	cases := []struct {
		name           string
		inputImgName   string
		expectUser     string
		expectPassWord string
		expectAuth     *types.AuthConfig
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:           "login with docker cli",
			inputImgName:   "hello-world:test",
			expectAuth:     &types.AuthConfig{Username: "test", Password: "test"},
			expectUser:     "test",
			expectPassWord: "test",
			funcBeforeTest: func_with_login,
		},
		{
			name:           "logout with docker cli",
			inputImgName:   "hello-world:test",
			expectAuth:     nil,
			expectUser:     "",
			expectPassWord: "",
			funcBeforeTest: func_with_logout,
		},
		{
			name:           "fail to read cli config file",
			inputImgName:   "hello-world:test",
			expectAuth:     nil,
			expectUser:     "",
			expectPassWord: "",
			funcBeforeTest: func_fail_read_config_file,
		},
		{
			name:           "invalid format of cli config file",
			inputImgName:   "hello-world:test",
			expectAuth:     nil,
			expectUser:     "",
			expectPassWord: "",
			funcBeforeTest: func_with_invalid_config_format,
		},
		{
			name:           "auth string in cli config file failed decode by base64",
			inputImgName:   "hello-world:test",
			expectAuth:     nil,
			expectUser:     "",
			expectPassWord: "",
			funcBeforeTest: func_fail_decode_authstr,
		},
		{
			name:           "the format of auth string in cli config file is invalid.",
			inputImgName:   "hello-world:test",
			expectAuth:     nil,
			expectUser:     "",
			expectPassWord: "",
			funcBeforeTest: func_with_invalid_authstr,
		},
		{
			name:           "password with special characters",
			inputImgName:   "hello-world:test",
			expectAuth:     &types.AuthConfig{Username: "test", Password: "Test:test:test"},
			expectUser:     "test",
			expectPassWord: "Test:test:test",
			funcBeforeTest: func_password_include_colon,
		},
		{
			name:           "pull images from unofficial registry",
			inputImgName:   "quay.io/prometheus/test:test",
			expectAuth:     nil,
			expectUser:     "",
			expectPassWord: "",
			funcBeforeTest: func_with_login,
		},
		{
			name:           "pull images from private registry",
			inputImgName:   "128.0.0.1:5000/testreg:test",
			expectAuth:     nil,
			expectUser:     "",
			expectPassWord: "",
			funcBeforeTest: func_with_login,
		},
		{
			name:           "empty image name",
			inputImgName:   "",
			expectAuth:     nil,
			expectUser:     "",
			expectPassWord: "",
			funcBeforeTest: func_with_login,
		},
		{
			name:           "wrong image name formate",
			inputImgName:   "128.0.0.1:8080/testReg:test",
			expectAuth:     nil,
			expectUser:     "",
			expectPassWord: "",
			funcBeforeTest: func_with_login,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest()
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}

			auth, _ := LoadDockerCliCredentials(tc.inputImgName)

			if tc.expectAuth == nil && auth == nil {
				t.Logf("Done")
			} else if tc.expectAuth != nil && auth != nil {
				if auth.Username == tc.expectAuth.Username && auth.Password == tc.expectAuth.Password {
					t.Logf("Done")
				} else {
					t.Errorf("get auth %s , passwd %s is not expected.", auth.Username, auth.Password)
				}
			} else {
				t.Errorf("get auth unequal to expect value.")
			}

		})
	}

}
