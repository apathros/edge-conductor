/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	cmapi "ep/pkg/api/certmgr"
	"ep/pkg/api/plugins"
	epapiplugins "ep/pkg/api/plugins"
	certmgr "ep/pkg/certmgr"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	testError  = fmt.Errorf("test error")
	errFile    = fmt.Errorf("file does not exist")
	errFileDir = fmt.Errorf("no such file or directory")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func patchInitTopConfig(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(init_kit_config, func() error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchCheckInitCmd(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(check_init_cmd, func() error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchGenCertAndConfig(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(certmgr.GenCertAndConfig, func(certbundle cmapi.Certificate, hosts string) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func multiplePatchGenCertAndConfig(t *testing.T, err error, nextPatch func()) {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchMethod(certmgr.GenCertAndConfig, func(certbundle cmapi.Certificate, hosts string) error {
		unpatch(t, patch)
		if nextPatch != nil {
			nextPatch()
		}
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}
}

//nolint:unparam
func patchUserCurrent(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(user.Current, func() (*user.User, error) {
		return &user.User{Username: "test"}, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchMakeDir(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(MakeDir, func(path string) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchFileNameofRuntime(t *testing.T, retrunValue string, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(FileNameofRuntime, func(_ string) (string, error) {
		return retrunValue, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchEpWfPreInit(t *testing.T, returnValue *epapiplugins.EpParams, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(EpWfPreInit, func(epPms *epapiplugins.EpParams, p map[string]string) (*epapiplugins.EpParams, error) {
		return returnValue, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchEpWfLoadServices(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(EpwfLoadServices, func(epParams *epapiplugins.EpParams) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchEpWfStart(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(EpWfStart, func(epParams *epapiplugins.EpParams, name string) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchEpWfTearDown(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(EpWfTearDown, func(epParams *epapiplugins.EpParams, rfile string) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchCopyCaRuntimeDataDir(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(copyCaRuntimeDataDir, func(registry string, workspace string, runtimedir string, certpath string) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

//nolint:unparam
func patchStat(t *testing.T, info fs.FileInfo, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(os.Stat, func(name string) (fs.FileInfo, error) {
		return info, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchStatOnce(t *testing.T, info fs.FileInfo, err error) {
	var patch *mpatch.Patch
	patch, _ = mpatch.PatchMethod(os.Stat, func(name string) (fs.FileInfo, error) {
		unpatch(t, patch)
		return info, err
	})
}

func patch_load_kit_bcfg(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(load_kit_bcfg, func(file string) (*KitBaseConfig, error) {
		return &KitBaseConfig{}, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchKitconfigValidate(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&epapiplugins.Kitconfig{}), "Validate", func(_ *epapiplugins.Kitconfig, formats strfmt.Registry) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func isWantedError(returnErr error, wantError error) bool {
	if !errors.Is(returnErr, wantError) &&
		(returnErr == nil || wantError == nil || !strings.Contains(returnErr.Error(), wantError.Error())) {
		return false
	}
	return true
}

func getFuncName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function
}

func getTemplateCfg() epapiplugins.Kitconfig {
	return epapiplugins.Kitconfig{
		Parameters: &epapiplugins.KitconfigParameters{
			Customconfig: &epapiplugins.Customconfig{
				Registry: &epapiplugins.CustomconfigRegistry{
					User:     "test",
					Password: "test",
				},
			},
			GlobalSettings: &epapiplugins.KitconfigParametersGlobalSettings{
				RegistryPort: "9000",
				WorkflowPort: "50088",
				ProviderIP:   "10.10.10.1",
			},
		},
		Cluster: &epapiplugins.KitconfigCluster{
			Manifests: []string{DefaultClusterManifests},
			Provider:  "kind",
			Config:    "config/cluster-provider/kind_cluster.yml",
		},
		OS: &epapiplugins.KitconfigOS{
			Manifests: []string{DefaultOSManifests},
			Provider:  "esp",
			Config:    "test",
		},
		Components: &epapiplugins.KitconfigComponents{
			Manifests: []string{"config/manifests/component_manifest.yml"},
			Selector: []*plugins.KitconfigComponentsSelectorItems0{
				{Name: "portainer"},
				{Name: "test", OverrideYaml: "test"},
			},
		},
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

func TestCheckInitCmd(t *testing.T) {
	cases := []struct {
		funcBeforeTest func() []*mpatch.Patch
		wantError      error
		funcAfterTest  func()
	}{
		{
			funcBeforeTest: func() []*mpatch.Patch {
				kitcfg = epapiplugins.Kitconfig{}
				patch := patchStat(t, nil, os.ErrNotExist)
				kitcfg.Cluster = &plugins.KitconfigCluster{}
				kitcfg.Components = &plugins.KitconfigComponents{}
				return []*mpatch.Patch{patch}
			},
			wantError: errFile,
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				kitcfg = epapiplugins.Kitconfig{}
				fileName := "tmp/test.TestCheckInitCmd.data"
				cfgFile = fileName
				patch1 := patchStat(t, nil, nil)
				patch2 := patchKitconfigValidate(t, testError)
				kitcfg.Cluster = &plugins.KitconfigCluster{Config: fileName}
				kitcfg.Components = &plugins.KitconfigComponents{}
				return []*mpatch.Patch{patch1, patch2}
			},
			wantError: testError,
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				kitcfg = epapiplugins.Kitconfig{}
				fileName := "tmp/test.TestCheckInitCmd.data"
				cfgFile = fileName
				patch := patchStat(t, nil, nil)
				kitcfg.Cluster = &plugins.KitconfigCluster{Config: fileName}
				kitcfg.Components = &plugins.KitconfigComponents{}
				return []*mpatch.Patch{patch}
			},
			wantError: nil,
		},
	}

	for n, testCase := range cases {
		t.Logf("%s case %d start", getFuncName(), n)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := check_init_cmd()
			if !isWantedError(err, testCase.wantError) {
				t.Errorf("Unexpected error: %v", err)
			}
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}

	t.Log("Done")
}

func TestInitCmd(t *testing.T) {
	cases := []struct {
		funcBeforeTest func() []*mpatch.Patch
		wantError      error
	}{
		{
			wantError: errFileDir,
		},
		{
			wantError: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				kitcfg = getTemplateCfg()
				patchInitTopConfig := patchInitTopConfig(t, nil)
				patchCheckInitCmd := patchCheckInitCmd(t, nil)
				patchGenCertAndConfig := patchGenCertAndConfig(t, nil)
				patchMakeDir := patchMakeDir(t, nil)
				patchUserCurrent := patchUserCurrent(t, nil)
				epparams := InitEpParams(kitcfg)
				patchEpWfPreInit := patchEpWfPreInit(t, epparams, nil)
				patchEpWfStart := patchEpWfStart(t, nil)
				patchCopyCaRuntimeDataDir := patchCopyCaRuntimeDataDir(t, nil)
				patchEpWfTearDown := patchEpWfTearDown(t, testError)
				return []*mpatch.Patch{patchInitTopConfig, patchCheckInitCmd, patchUserCurrent, patchGenCertAndConfig, patchMakeDir, patchEpWfPreInit, patchEpWfStart, patchCopyCaRuntimeDataDir, patchEpWfTearDown}
			},
		},
	}

	for n, testCase := range cases {
		t.Logf("%s case %d start", getFuncName(), n)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := initCmd.RunE(&cobra.Command{}, []string{})
			if !isWantedError(err, testCase.wantError) {
				t.Errorf("Unexpected error: %v", err)
			}
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}

	t.Log("Done")
}

func TestInitTopConfig(t *testing.T) {
	cases := []struct {
		funcBeforeTest      func() []*mpatch.Patch
		isFunctionCorrectly func(err error)
	}{
		{
			funcBeforeTest: func() []*mpatch.Patch {
				kitcfg = epapiplugins.Kitconfig{}
				patch := patch_load_kit_bcfg(t, testError)
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
	}

	for n, testCase := range cases {
		t.Logf("%s case %d start", getFuncName(), n)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}
			testCase.isFunctionCorrectly(init_kit_config())
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}

	t.Log("Done")
}

func TestEpInit(t *testing.T) {
	cases := []struct {
		funcBeforeTest      func() []*mpatch.Patch
		isFunctionCorrectly func(err error)
	}{
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patch := patchInitTopConfig(t, testError)
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				patchInitTopConfig := patchInitTopConfig(t, nil)
				patchCheckInitCmd := patchCheckInitCmd(t, testError)
				return []*mpatch.Patch{patchInitTopConfig, patchCheckInitCmd}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				kitcfg = getTemplateCfg()
				patchInitTopConfig := patchInitTopConfig(t, nil)
				patchCheckInitCmd := patchCheckInitCmd(t, nil)
				patchGenCertAndConfig := patchGenCertAndConfig(t, testError)
				return []*mpatch.Patch{patchInitTopConfig, patchCheckInitCmd, patchGenCertAndConfig}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				kitcfg = getTemplateCfg()
				patchInitTopConfig := patchInitTopConfig(t, nil)
				patchCheckInitCmd := patchCheckInitCmd(t, nil)
				multiplePatchGenCertAndConfig(t, nil, func() {
					multiplePatchGenCertAndConfig(t, testError, nil)
				})
				return []*mpatch.Patch{patchInitTopConfig, patchCheckInitCmd}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				kitcfg = getTemplateCfg()
				patchInitTopConfig := patchInitTopConfig(t, nil)
				patchCheckInitCmd := patchCheckInitCmd(t, nil)
				patchGenCertAndConfig := patchGenCertAndConfig(t, nil)
				patchMakeDir := patchMakeDir(t, nil)
				patchUserCurrent := patchUserCurrent(t, nil)
				patchFileNameofRuntime := patchFileNameofRuntime(t, "", testError)
				return []*mpatch.Patch{patchInitTopConfig, patchCheckInitCmd, patchUserCurrent, patchGenCertAndConfig, patchMakeDir, patchFileNameofRuntime}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				kitcfg = getTemplateCfg()
				patchInitTopConfig := patchInitTopConfig(t, nil)
				patchCheckInitCmd := patchCheckInitCmd(t, nil)
				patchGenCertAndConfig := patchGenCertAndConfig(t, nil)
				patchMakeDir := patchMakeDir(t, nil)
				patchUserCurrent := patchUserCurrent(t, nil)
				patchEpWfPreInit := patchEpWfPreInit(t, nil, testError)
				return []*mpatch.Patch{patchInitTopConfig, patchCheckInitCmd, patchUserCurrent, patchGenCertAndConfig, patchMakeDir, patchEpWfPreInit}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				kitcfg = getTemplateCfg()
				patchInitTopConfig := patchInitTopConfig(t, nil)
				patchCheckInitCmd := patchCheckInitCmd(t, nil)
				patchGenCertAndConfig := patchGenCertAndConfig(t, nil)
				patchMakeDir := patchMakeDir(t, nil)
				patchUserCurrent := patchUserCurrent(t, nil)
				epparams := InitEpParams(kitcfg)
				patchEpWfPreInit := patchEpWfPreInit(t, epparams, nil)
				patchEpWfStart := patchEpWfStart(t, testError)
				return []*mpatch.Patch{patchInitTopConfig, patchCheckInitCmd, patchUserCurrent, patchGenCertAndConfig, patchMakeDir, patchEpWfPreInit, patchEpWfStart}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				kitcfg = getTemplateCfg()
				patchInitTopConfig := patchInitTopConfig(t, nil)
				patchCheckInitCmd := patchCheckInitCmd(t, nil)
				patchGenCertAndConfig := patchGenCertAndConfig(t, nil)
				patchMakeDir := patchMakeDir(t, nil)
				patchUserCurrent := patchUserCurrent(t, nil)
				epparams := InitEpParams(kitcfg)
				patchEpWfPreInit := patchEpWfPreInit(t, epparams, nil)
				patchEpWfStart := patchEpWfStart(t, nil)
				patchCopyCaRuntimeDataDir := patchCopyCaRuntimeDataDir(t, nil)
				patchEpWfTearDown := patchEpWfTearDown(t, testError)
				return []*mpatch.Patch{patchInitTopConfig, patchCheckInitCmd, patchUserCurrent, patchGenCertAndConfig, patchMakeDir, patchEpWfPreInit, patchEpWfStart, patchCopyCaRuntimeDataDir, patchEpWfTearDown}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, nil) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
		{
			funcBeforeTest: func() []*mpatch.Patch {
				kitcfg = getTemplateCfg()
				kitcfg = getTemplateCfg()
				patchInitTopConfig := patchInitTopConfig(t, nil)
				patchCheckInitCmd := patchCheckInitCmd(t, nil)
				patchGenCertAndConfig := patchGenCertAndConfig(t, nil)
				patchMakeDir := patchMakeDir(t, nil)
				patchUserCurrent := patchUserCurrent(t, nil)
				epparams := InitEpParams(kitcfg)
				patchEpWfPreInit := patchEpWfPreInit(t, epparams, nil)
				patchEpWfStart := patchEpWfStart(t, nil)
				patchCopyCaRuntimeDataDir := patchCopyCaRuntimeDataDir(t, testError)

				return []*mpatch.Patch{patchInitTopConfig, patchCheckInitCmd, patchUserCurrent, patchGenCertAndConfig, patchMakeDir, patchEpWfPreInit, patchEpWfStart, patchCopyCaRuntimeDataDir}
			},
			isFunctionCorrectly: func(err error) {
				if !isWantedError(err, testError) {
					t.Errorf("Unexpected error: %v", err)
				}
			},
		},
	}

	for n, testCase := range cases {
		t.Logf("%s case %d start", getFuncName(), n)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}
			testCase.isFunctionCorrectly(ep_init())
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}

	t.Log("Done")
}
