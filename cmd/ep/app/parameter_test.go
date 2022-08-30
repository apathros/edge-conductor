/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	"errors"
	cmapi "github.com/intel/edge-conductor/pkg/api/certmgr"
	epapiplugins "github.com/intel/edge-conductor/pkg/api/plugins"
	"github.com/intel/edge-conductor/pkg/eputils"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/foomo/htpasswd"
	"github.com/undefinedlabs/go-mpatch"
	"sigs.k8s.io/yaml"
)

var (
	errEmpty      = errors.New("")
	errClose      = errors.New("close.error")
	errConvStruct = errors.New("convertschemastruct.error")
	errMkDir      = errors.New("makedir.error")
	errMarshal    = errors.New("marshal.error")
	errReadfile   = errors.New("readfile.error")
	errRemove     = errors.New("remove.error")
	errSetPwd     = errors.New("setpassword.error")
	errTempFile   = errors.New("tempfile.error")
)

//nolint:unparam
func patchOsStat(t *testing.T, fileInfo fs.FileInfo, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Stat, func(_ string) (fs.FileInfo, error) {
		unpatch(t, patch)
		return fileInfo, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

//nolint:unparam
func patchReadFile(t *testing.T, data []byte, err error) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.ReadFile, func(name string) ([]byte, error) {
		unpatch(t, patch)
		return data, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchtempfile(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(ioutil.TempFile, func(dir, pattern string) (f *os.File, err error) {
		unpatch(t, patch)
		if ok {
			return &os.File{}, nil
		} else {
			return nil, errTempFile
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchclose(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(&os.File{}), "Close", func(*os.File) error {
		unpatch(t, patch)
		if ok {
			return nil
		} else {
			return errClose
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

}

func patchremove(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Remove, func(name string) error {
		unpatch(t, patch)
		if ok {
			return nil
		} else {
			return errRemove
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchsetpassword(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(htpasswd.SetPassword, func(file, name, password string, hashAlgorithm htpasswd.HashAlgorithm) error {
		unpatch(t, patch)
		if ok {
			return nil
		} else {
			return errSetPwd
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func patchmarshal(t *testing.T, ok bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(yaml.Marshal, func(o interface{}) ([]byte, error) {
		unpatch(t, patch)
		if ok {
			return []byte{}, nil
		} else {
			return nil, errMarshal
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func getepparams(t *testing.T, configfile string) *epapiplugins.EpParams {
	_, pwdpath, _, _ := runtime.Caller(0)
	testdatapath := filepath.Join(filepath.Dir(pwdpath), "testdata")

	b, err := os.ReadFile(testdatapath + "/" + configfile)
	if err != nil {
		t.Errorf("Can't read %v", configfile)
	}

	epparams := &epapiplugins.EpParams{}
	err = epparams.UnmarshalBinary(b)
	if err != nil {
		t.Errorf("Can't UnmarshalBinary %v", configfile)
	}
	return epparams

}

func getrealpath(file string) string {
	_, pwdpath, _, _ := runtime.Caller(0)
	testdatapath := filepath.Join(filepath.Dir(pwdpath), "testdata")
	return testdatapath + "/" + file

}

func patchisnotexist(t *testing.T, yes bool) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.IsNotExist, func(err error) bool {
		unpatch(t, patch)
		if yes {
			return true
		} else {
			return false
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func Test_setupcustomconfig(t *testing.T) {
	cases := []struct {
		name          string
		expectError   error
		in_kitcfgPath string
		in_epp        *epapiplugins.EpParams
		beforetest    func()
		teardown      func()
	}{
		{
			name:          "kitcfgpath is not valid",
			expectError:   errEmpty,
			in_kitcfgPath: "",
			in_epp:        getepparams(t, "epparams_withoutctmcfg.json"),
		},
		{
			name:          "kitcfgpath not provided",
			expectError:   eputils.GetError("errConfigPath"),
			in_kitcfgPath: "",
			in_epp:        getepparams(t, "epparams_withctmcfg.json"),
		},
		{
			name:          "need user password",
			expectError:   eputils.GetError("errRegistryPw"),
			in_kitcfgPath: getrealpath("ctmcfg_withloginfail.yml"),
			in_epp:        getepparams(t, "epparams_withreg.json"),
			beforetest: func() {
				patchOsStat(t, nil, nil)
				patchisnotexist(t, false)
			},
		},
		{
			name:          "with_registry_password",
			expectError:   nil,
			in_kitcfgPath: getrealpath("ctmcfg_withloginok.yml"),
			in_epp:        getepparams(t, "epparams_withadminpass.json"),
			beforetest: func() {
				patchOsStat(t, nil, nil)
				patchisnotexist(t, false)
			},
		},
		{
			name:          "without_registry_password",
			expectError:   eputils.GetError("errRegistryPw"),
			in_kitcfgPath: getrealpath("ctmcfg_withloginfail.yml"),
			in_epp:        getepparams(t, "epparams_withadminpass.json"),
			beforetest: func() {
				patchOsStat(t, nil, nil)
				patchisnotexist(t, false)
			},
		},
		{
			name:          "with_registry_password_components",
			expectError:   nil,
			in_kitcfgPath: getrealpath("ctmcfg_withloginok_components.yml"),
			in_epp:        getepparams(t, "epparams_withadminpass.json"),
			beforetest: func() {
				patchOsStat(t, nil, nil)
				patchisnotexist(t, false)
			},
		},
		{
			name:          "testchant1",
			expectError:   nil,
			in_kitcfgPath: getrealpath("ctmcfg_withloginok_components.yml"),
			in_epp:        getepparams(t, "epparams_withdistro.json"),
			beforetest: func() {
				patchOsStat(t, nil, nil)
				patchisnotexist(t, false)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}
				err := setupCustomConfig(tc.in_kitcfgPath, tc.in_epp)

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

func patchconvertschemastruct(t *testing.T, registrycertok bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(eputils.ConvertSchemaStruct, func(from interface{}, to interface{}) error {
		switch i := from.(type) {
		default:
			t.Errorf("patch error: type %v mismatch", i)
			return errConvStruct
		case *cmapi.Certificate:
			if registrycertok {
				return nil
			} else {
				return errConvStruct
			}
		case cmapi.Certificate:
			if registrycertok {
				return nil
			} else {
				return errConvStruct
			}

		}

	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

//nolint:unparam
func getkitconfig(t *testing.T, configfile string) epapiplugins.Kitconfig {
	_, pwdpath, _, _ := runtime.Caller(0)
	testdatapath := filepath.Join(filepath.Dir(pwdpath), "testdata")

	b, err := os.ReadFile(testdatapath + "/" + configfile)
	if err != nil {
		t.Errorf("Can't read %v", configfile)
	}

	kitconfig := epapiplugins.Kitconfig{}
	err = kitconfig.UnmarshalBinary(b)
	if err != nil {
		t.Errorf("Can't UnmarshalBinary %v", configfile)
	}
	return kitconfig
}

func patchmakedir(t *testing.T, runtimebinok bool, runtimedataok bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(MakeDir, func(path string) error {
		if strings.Contains(path, "bin") {
			if runtimebinok {
				return nil
			} else {
				return errMkDir
			}
		} else if strings.Contains(path, "data") {
			if runtimedataok {
				return nil
			} else {
				return errMkDir
			}
		} else {
			return errMkDir
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func Test_InitEpParams(t *testing.T) {
	var patchHandler *mpatch.Patch
	var patchHandler_MakeDir *mpatch.Patch
	cases := []struct {
		name            string
		expectNilOutput bool
		in_eptopconfg   epapiplugins.Kitconfig
		beforetest      func()
		teardown        func()
	}{
		{
			name:            "Convert schema struct fail for registrycert",
			expectNilOutput: true,
			in_eptopconfg:   getkitconfig(t, "kitconfig_sample.json"),
			beforetest: func() {
				patchHandler = patchconvertschemastruct(t, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
			},
		},
		{
			name:            "MakeDir fail for runtimebin",
			expectNilOutput: true,
			in_eptopconfg:   getkitconfig(t, "kitconfig_sample.json"),
			beforetest: func() {
				patchHandler = patchconvertschemastruct(t, true)
				patchHandler_MakeDir = patchmakedir(t, false, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
				unpatch(t, patchHandler_MakeDir)
			},
		},
		{
			name:            "MakeDir fail for runtimedata",
			expectNilOutput: true,
			in_eptopconfg:   getkitconfig(t, "kitconfig_sample.json"),
			beforetest: func() {
				patchHandler = patchconvertschemastruct(t, true)
				patchHandler_MakeDir = patchmakedir(t, true, false)
			},
			teardown: func() {
				unpatch(t, patchHandler)
				unpatch(t, patchHandler_MakeDir)
			},
		},
		{
			name:            "InitEpParams ok",
			expectNilOutput: false,
			in_eptopconfg:   getkitconfig(t, "kitconfig_sample.json"),
			beforetest: func() {
				patchHandler = patchconvertschemastruct(t, true)
				patchHandler_MakeDir = patchmakedir(t, true, true)
			},
			teardown: func() {
				unpatch(t, patchHandler)
				unpatch(t, patchHandler_MakeDir)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				patchUserCurrent := patchUserCurrent(t, nil)
				if tc.beforetest != nil {
					tc.beforetest()
				}

				epparams := InitEpParams(tc.in_eptopconfg)

				if tc.expectNilOutput && epparams != nil {
					t.Errorf("Unexpected output: Epparams is not nil")
				} else if !tc.expectNilOutput && epparams == nil {
					t.Errorf("Unexpected output: Epparams is nil")
				}

				if tc.teardown != nil {
					tc.teardown()
				}
				unpatch(t, patchUserCurrent)

			}

		})
	}
}

func Test_readfile(t *testing.T) {
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
	}{
		{
			name:        "readfile failed",
			expectError: nil,
			beforetest: func() {
				patchReadFile(t, nil, errReadfile)
			},
		},
		{
			name:        "readfile ok",
			expectError: nil,
			beforetest: func() {
				patchReadFile(t, nil, nil)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				_, err := funcs["readfile"].(func(string) (string, error))("")

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

			}

		})
	}
}

func Test_structtoyaml(t *testing.T) {
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

				_, err := funcs["structtoyaml"].(func(interface{}) (string, error))("")

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

			}

		})
	}
}

func patchexit(t *testing.T) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(os.Exit, func(code int) {
		unpatch(t, patch)
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

func Test_htpasswd(t *testing.T) {
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
	}{
		{
			name:        "TempFile failed",
			expectError: nil,
			beforetest: func() {
				patchtempfile(t, false)
			},
		},
		{
			name:        "Readfile failed",
			expectError: nil,
			beforetest: func() {
				patchremove(t, true)
				patchsetpassword(t, true)
				patchReadFile(t, nil, errReadfile)
			},
		},
		{
			name:        "funcs_htpasswd ok",
			expectError: nil,
			beforetest: func() {
				patchremove(t, true)
				patchsetpassword(t, true)
				patchReadFile(t, nil, nil)
				patchclose(t, true)
			},
		},
		{
			name:        "tempfile close fail",
			expectError: nil,
			beforetest: func() {
				patchremove(t, true)
				patchsetpassword(t, true)
				patchReadFile(t, nil, nil)
				patchclose(t, false)
				patchexit(t)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				_, err := funcs["htpasswd"].(func(usr string, pwd string) (string, error))("", "")

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}

			}

		})
	}

}

func Test_base64(t *testing.T) {
	cases := []struct {
		name        string
		expectError error
		beforetest  func()
	}{
		{
			name:        "Base64 ok",
			expectError: nil,
			beforetest:  nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				b64, err := funcs["base64"].(func(string) (string, error))("admin")
				if b64 != "YWRtaW4=" {
					t.Errorf("Unexpected value")
				}

				if !isExpectedError(err, tc.expectError) {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func Test_get_files(t *testing.T) {
	_, pwdpath, _, _ := runtime.Caller(0)
	testdatapath := filepath.Join(filepath.Dir(pwdpath), "testdata")

	files, err := get_files(testdatapath)
	if err != nil || len(files) <= 0 {
		t.Errorf("Failed to get files under %s.", testdatapath)
	}

	files, err = get_files("not-exist")
	if err != nil || len(files) > 0 {
		t.Errorf("Error get files from a non-exist folder.")
	}
}

func Test_schemastruct_to_yaml_item(t *testing.T) {
	patch, patchErr := mpatch.PatchMethod(eputils.SchemaStructToYaml, func(v eputils.SchemaStruct) (string, error) {
		return "test", nil
	})
	if patchErr != nil {
		t.Errorf("Failed to patch function.")
	}
	defer unpatch(t, patch)

	v := &epapiplugins.EpParams{}
	if result, err := schemastruct_to_yaml_item(v); err != nil || result != "- test\n" {
		t.Errorf("Wrong result: %s (error: %s)", result, err)
	}
}

func Test_wf_include(t *testing.T) {
	_, pwdpath, _, _ := runtime.Caller(0)
	testwffile := filepath.Join(filepath.Dir(pwdpath), "testdata", "workflow")

	if result, err := wf_include_data(testwffile); err != nil || strings.Contains("test-data", result) {
		t.Errorf("Unexpected value \"%s\" (error:%s)", result, err)
	}
	if result, err := wf_include_workflows(testwffile); err != nil || strings.Contains("test-workflow", result) {
		t.Errorf("Unexpected value \"%s\" (error:%s)", result, err)
	}
	if result, err := wf_include_containers(testwffile); err != nil || strings.Contains("test-container", result) {
		t.Errorf("Unexpected value \"%s\" (error:%s)", result, err)
	}
	if result, err := wf_include_plugins(testwffile); err != nil || strings.Contains("test-plugin", result) {
		t.Errorf("Unexpected value \"%s\" (error:%s)", result, err)
	}
}

func isExpectedError(returnErr error, wantError error) bool {
	if !errors.Is(returnErr, wantError) &&
		(returnErr == nil || wantError == nil || !strings.Contains(returnErr.Error(), wantError.Error())) {
		return false
	}
	return true
}
