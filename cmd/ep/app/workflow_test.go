/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
//nolint: dupl
package app

import (
	"errors"
	"fmt"
	"github.com/intel/edge-conductor/pkg/api/plugins"
	epapiplugins "github.com/intel/edge-conductor/pkg/api/plugins"
	"github.com/intel/edge-conductor/pkg/eputils"
	docker "github.com/intel/edge-conductor/pkg/eputils/docker"
	"github.com/intel/edge-conductor/pkg/eputils/orasutils"
	wf "github.com/intel/edge-conductor/pkg/workflow"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	//	errTopConf = errors.New("Top Config Lost!")
	errTest = errors.New("test_error")
)

func patchSetHostIptoNoProxy(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(setHostIptoNoProxy, func(input_ep_params *epapiplugins.EpParams) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchSetupCustomConfig(t *testing.T, err error, checkParamFunc func(ctmcfgPath string, epp *epapiplugins.EpParams)) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(setupCustomConfig, func(ctmcfgPath string, epp *epapiplugins.EpParams) error {
		if checkParamFunc != nil {
			checkParamFunc(ctmcfgPath, epp)
		}
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchGetAuthConf(t *testing.T, authConfig *types.AuthConfig, err error, checkParamFunc func(server string, port string, user string, password string)) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(docker.GetAuthConf, func(server string, port string, user string, password string) (*types.AuthConfig, error) {
		if checkParamFunc != nil {
			checkParamFunc(server, port, user, password)
		}
		return authConfig, err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}
func patchOrasNewClient(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(orasutils.OrasNewClient, func(authConf *types.AuthConfig, cacert string) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchSaveSchemaStructToYamlFile(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(eputils.SaveSchemaStructToYamlFile, func(v eputils.SchemaStruct, file string) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func TestEpWfStart(t *testing.T) {
	isFunctionCorrectlyFunc := func(err, wantError error) {
		if !isWantedError(err, wantError) {
			t.Errorf("expected error: %v, but function returned error: %v", wantError, err)
		}
	}
	cases := []struct {
		funcBeforeTest      func() []*mpatch.Patch
		wantError           error
		epParams            *epapiplugins.EpParams
		name                string
		isFunctionCorrectly func(err, wantError error)
	}{
		{
			name:                "GetRuntimeTopConfig fail",
			wantError:           eputils.GetError("errKitConfig"),
			isFunctionCorrectly: isFunctionCorrectlyFunc,
			funcBeforeTest: func() []*mpatch.Patch {
				patch, patchErr := mpatch.PatchMethod(GetRuntimeTopConfig, func(epParams *epapiplugins.EpParams) *epapiplugins.Kitconfig {
					return nil
				})

				if patchErr != nil {
					t.Errorf("patch error: %v", patchErr)
				}

				return []*mpatch.Patch{patch}
			},
			epParams: &epapiplugins.EpParams{
				Kitconfig: &epapiplugins.Kitconfig{
					Parameters: &epapiplugins.KitconfigParameters{
						GlobalSettings: &epapiplugins.KitconfigParametersGlobalSettings{
							WorkflowPort: "8228",
						},
					},
				},
			},
		},
		{
			wantError: testError,
			epParams: &epapiplugins.EpParams{
				Kitconfig: &epapiplugins.Kitconfig{
					Parameters: &epapiplugins.KitconfigParameters{
						GlobalSettings: &epapiplugins.KitconfigParametersGlobalSettings{
							WorkflowPort: "8228",
						},
					},
				},
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
			funcBeforeTest: func() []*mpatch.Patch {
				patch, patchErr := mpatch.PatchMethod(wf.Start, func(name string, address string, configFile string) error {
					return testError
				})
				if patchErr != nil {
					t.Errorf("patch error: %v", patchErr)
					return nil
				}
				return []*mpatch.Patch{patch}
			},
		},
		{
			wantError: nil,
			epParams: &epapiplugins.EpParams{
				Kitconfig: &epapiplugins.Kitconfig{
					Parameters: &epapiplugins.KitconfigParameters{
						GlobalSettings: &epapiplugins.KitconfigParametersGlobalSettings{
							ProviderIP:   "localhost",
							WorkflowPort: "8228",
						},
					},
				},
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
			name:                "test_name",
			funcBeforeTest: func() []*mpatch.Patch {
				patch, patchErr := mpatch.PatchMethod(wf.Start, func(name string, address string, configFile string) error {
					if name != "test_name" || address != "localhost:8228" || configFile != WfConfig {
						t.Errorf("The parameters of the workflow.Start function are not expected")
					}
					return nil
				})
				if patchErr != nil {
					t.Errorf("patch error: %v", patchErr)
					return nil
				}
				return []*mpatch.Patch{patch}
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
			err := EpWfStart(testCase.epParams, testCase.name)
			testCase.isFunctionCorrectly(err, testCase.wantError)
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}
	t.Log("Done")
}

func getepparamswithtopconfighost(t *testing.T, filefullpath string) *epapiplugins.EpParams {
	b, err := os.ReadFile(filefullpath)
	if err != nil {
		t.Errorf("Can't read %v", filefullpath)
	}

	epparams := &epapiplugins.EpParams{}
	err = epparams.UnmarshalBinary(b)
	if err != nil {
		t.Errorf("Can't unmarshalbinary %v", filefullpath)
	}

	return epparams
}

func gettestfilefullpath(filename string) string {
	_, pwdpath, _, _ := runtime.Caller(0)
	testdatapath := filepath.Join(filepath.Dir(pwdpath), "testdata")
	return testdatapath + "/" + filename
}

func TestSetHostIptoNoProxy(t *testing.T) {
	isFunctionCorrectlyFunc := func(err, wantError error) {
		if !isWantedError(err, wantError) {
			t.Errorf("expected error: %v, but function returned error: %v", wantError, err)
		}
	}
	cases := []struct {
		funcBeforeTest      func() []*mpatch.Patch
		wantError           error
		input_ep_params     *epapiplugins.EpParams
		isFunctionCorrectly func(err, wantError error)
	}{
		{
			wantError:           eputils.GetError("errHost"),
			input_ep_params:     nil,
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError:           testError,
			input_ep_params:     getepparamswithtopconfighost(t, gettestfilefullpath("workflow_epparams_withhost.json")),
			isFunctionCorrectly: isFunctionCorrectlyFunc,
			funcBeforeTest: func() []*mpatch.Patch {
				patch, patchErr := mpatch.PatchMethod(os.Setenv, func(key, value string) error {
					return testError
				})
				if patchErr != nil {
					t.Errorf("patch error: %v", patchErr)
				}
				return []*mpatch.Patch{patch}
			},
		},
		{
			wantError:       nil,
			input_ep_params: getepparamswithtopconfighost(t, gettestfilefullpath("workflow_epparams_withhost.json")),
			isFunctionCorrectly: func(err, wantError error) {
				isFunctionCorrectlyFunc(err, wantError)
				noProxy := os.Getenv("no_proxy")
				if noProxy != "127.0.0.1/8,192.168.1.1" {
					t.Errorf("Unexpected environment variable no_proxy: %v", noProxy)
				}

			},
			funcBeforeTest: func() []*mpatch.Patch {
				os.Setenv("no_proxy", "127.0.0.1/8")
				return nil
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
			err := setHostIptoNoProxy(testCase.input_ep_params)
			testCase.isFunctionCorrectly(err, testCase.wantError)
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}
	t.Log("Done")
}

func TestEpWfPreInit(t *testing.T) {
	isFunctionCorrectlyFunc := func(output *epapiplugins.EpParams, err, wantError error) {
		if !isWantedError(err, wantError) {
			t.Errorf("expected error: %v, but function returned error: %v", wantError, err)
		}
	}
	cases := []struct {
		funcBeforeTest      func() []*mpatch.Patch
		wantError           error
		epPms               *epapiplugins.EpParams
		p                   map[string]string
		isFunctionCorrectly func(output *epapiplugins.EpParams, err, wantError error)
	}{
		{
			wantError: eputils.GetError("errParameter"),
			epPms:     nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patch := patchFileNameofRuntime(t, "", eputils.GetError("errParameter"))
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: eputils.GetError("errParameter"),
			epPms:     nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patchFileNameofRuntime := patchFileNameofRuntime(t, "test_file_name", eputils.GetError("errParameter"))
				return []*mpatch.Patch{patchFileNameofRuntime}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: errFileDir,
			epPms:     nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patchFileNameofRuntime := patchFileNameofRuntime(t, "test_file_name", nil)
				patchStatOnce(t, nil, nil)
				return []*mpatch.Patch{patchFileNameofRuntime}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: testError,
			epPms: &epapiplugins.EpParams{
				Kitconfig: &epapiplugins.Kitconfig{
					Parameters: &epapiplugins.KitconfigParameters{
						Customconfig: &epapiplugins.Customconfig{},
					},
				},
			},
			p: map[string]string{
				Epcmdline:     "cmd_line",
				Epkubeconfig:  "kube_config",
				KitConfigPath: "config_path",
				"test":        "test",
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patch := patchSetupCustomConfig(t, testError, nil)
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: testError,
			epPms: &epapiplugins.EpParams{
				Kitconfigpath: "custom_config",
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patchSetupCustomConfig := patchSetupCustomConfig(t, nil, func(ctmcfgPath string, epp *epapiplugins.EpParams) {
					if ctmcfgPath != "custom_config" {
						t.Errorf("Unexpected parameter of setupCustomConfig ctmcfgPath: %v, wanted: %v", ctmcfgPath, "custom_config")
					}
				})
				patchSetHostIptoNoProxy := patchSetHostIptoNoProxy(t, testError)
				return []*mpatch.Patch{patchSetupCustomConfig, patchSetHostIptoNoProxy}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: testError,
			epPms: &epapiplugins.EpParams{
				Registrycert: &epapiplugins.Certificate{
					Ca: &epapiplugins.CertificateCa{
						Cert: "test_cert",
					},
				},
				Kitconfig: &epapiplugins.Kitconfig{
					Parameters: &epapiplugins.KitconfigParameters{
						Customconfig: &epapiplugins.Customconfig{
							Registry: &epapiplugins.CustomconfigRegistry{
								User:     "test_user",
								Password: "test_password",
							},
						},
						GlobalSettings: &epapiplugins.KitconfigParametersGlobalSettings{
							ProviderIP:   "test_server",
							RegistryPort: "8080",
						},
					},
				},
				Kitconfigpath: "testkitconfigpath",
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patchSetupCustomConfig := patchSetupCustomConfig(t, nil, nil)
				patchSetHostIptoNoProxy := patchSetHostIptoNoProxy(t, nil)
				patchGetAuthConf := patchGetAuthConf(t, nil, testError, func(server string, port string, user string, password string) {
					if server != "test_server" ||
						port != "8080" ||
						user != "test_user" ||
						password != "test_password" {
						t.Errorf("Unexpected parameter of setupCustomConfig server: %v, port: %v, user:%v, password: %v\nwanted server: %v, port: %v, user:%v, password: %v\nwanted",
							server, port, user, password,
							"test_server", "8080", "test_user", "test_password")
					}
				})
				return []*mpatch.Patch{patchSetupCustomConfig, patchSetHostIptoNoProxy, patchGetAuthConf}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: testError,
			epPms: &epapiplugins.EpParams{
				Registrycert: &epapiplugins.Certificate{
					Ca: &epapiplugins.CertificateCa{},
				},
				Kitconfig: &epapiplugins.Kitconfig{
					Parameters: &epapiplugins.KitconfigParameters{
						Customconfig: &epapiplugins.Customconfig{
							Registry: &epapiplugins.CustomconfigRegistry{},
						},
						GlobalSettings: &epapiplugins.KitconfigParametersGlobalSettings{},
					},
				},
				Kitconfigpath: "testkitconfigpath",
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patchSetupCustomConfig := patchSetupCustomConfig(t, nil, nil)
				patchSetHostIptoNoProxy := patchSetHostIptoNoProxy(t, nil)
				patchGetAuthConf := patchGetAuthConf(t, nil, nil, nil)
				patchOrasNewClient := patchOrasNewClient(t, testError)
				return []*mpatch.Patch{patchSetupCustomConfig, patchSetHostIptoNoProxy, patchGetAuthConf, patchOrasNewClient}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: nil,
			epPms: &epapiplugins.EpParams{
				Registrycert: &epapiplugins.Certificate{
					Ca: &epapiplugins.CertificateCa{},
				},
				Kitconfig: &epapiplugins.Kitconfig{
					Parameters: &epapiplugins.KitconfigParameters{
						Customconfig: &epapiplugins.Customconfig{
							Registry: &epapiplugins.CustomconfigRegistry{},
						},
						GlobalSettings: &epapiplugins.KitconfigParametersGlobalSettings{},
					},
				},
				Kitconfigpath: "testkitconfigpath",
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patchSetupCustomConfig := patchSetupCustomConfig(t, nil, nil)
				patchSetHostIptoNoProxy := patchSetHostIptoNoProxy(t, nil)
				patchGetAuthConf := patchGetAuthConf(t, nil, nil, nil)
				patchOrasNewClient := patchOrasNewClient(t, nil)
				return []*mpatch.Patch{patchSetupCustomConfig, patchSetHostIptoNoProxy, patchGetAuthConf, patchOrasNewClient}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
	}

	for n, testCase := range cases {
		t.Logf("%s case %d start", getFuncName(), n)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}
			epPms, err := EpWfPreInit(testCase.epPms, testCase.p)
			testCase.isFunctionCorrectly(epPms, err, testCase.wantError)
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}
	t.Log("Done")
}

func TestEpWfTearDown(t *testing.T) {
	isFunctionCorrectlyFunc := func(err, wantError error) {
		if !isWantedError(err, wantError) {
			t.Errorf("expected error: %v, but function returned error: %v", wantError, err)
		}
	}
	cases := []struct {
		funcBeforeTest      func() []*mpatch.Patch
		wantError           error
		epParams            *epapiplugins.EpParams
		rfile               string
		isFunctionCorrectly func(err, wantError error)
	}{
		{
			wantError: testError,
			epParams: &epapiplugins.EpParams{
				Kitconfig: &epapiplugins.Kitconfig{
					Parameters: &epapiplugins.KitconfigParameters{
						Customconfig: &epapiplugins.Customconfig{},
					},
				},
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patch := patchSaveSchemaStructToYamlFile(t, testError)
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
		{
			wantError: nil,
			epParams: &epapiplugins.EpParams{
				Kitconfig: &epapiplugins.Kitconfig{
					Parameters: &epapiplugins.KitconfigParameters{
						Customconfig: &epapiplugins.Customconfig{},
					},
				},
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patch := patchSaveSchemaStructToYamlFile(t, nil)
				return []*mpatch.Patch{patch}
			},
			isFunctionCorrectly: isFunctionCorrectlyFunc,
		},
	}

	for n, testCase := range cases {
		t.Logf("%s case %d start", getFuncName(), n)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}
			err := EpWfTearDown(testCase.epParams, testCase.rfile)
			testCase.isFunctionCorrectly(err, testCase.wantError)
		}()
		t.Logf("%s case %d End", getFuncName(), n)
	}
	t.Log("Done")
}

func TestEpwfLoadServices(t *testing.T) {
	func_load_services_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchLoadEkServices, err := mpatch.PatchMethod(load_kit_services, func(kitconfig *epapiplugins.Kitconfig, kitfile string) error {
			return errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchLoadEkServices}
	}

	func_FileNameofRuntime_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchLoadEkServices, err := mpatch.PatchMethod(load_kit_services, func(kitconfig *epapiplugins.Kitconfig, kitfile string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchFileNameofRuntime, err := mpatch.PatchMethod(FileNameofRuntime, func(target_name string) (string, error) {
			return "", errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		return []*mpatch.Patch{pathchLoadEkServices, pathchFileNameofRuntime}
	}

	func_SaveSchemaStructToYamlFile_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchLoadEkServices, err := mpatch.PatchMethod(load_kit_services, func(kitconfig *epapiplugins.Kitconfig, kitfile string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchFileNameofRuntime, err := mpatch.PatchMethod(FileNameofRuntime, func(target_name string) (string, error) {
			return "testPath", nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchSaveSchemaStructToYamlFile, err := mpatch.PatchMethod(eputils.SaveSchemaStructToYamlFile, func(v eputils.SchemaStruct, file string) error {
			return errTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchLoadEkServices, pathchFileNameofRuntime, pathchSaveSchemaStructToYamlFile}
	}

	func_everyFunc_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchLoadEkServices, err := mpatch.PatchMethod(load_kit_services, func(kitconfig *epapiplugins.Kitconfig, kitfile string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchFileNameofRuntime, err := mpatch.PatchMethod(FileNameofRuntime, func(target_name string) (string, error) {
			return "testPath", nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchSaveSchemaStructToYamlFile, err := mpatch.PatchMethod(eputils.SaveSchemaStructToYamlFile, func(v eputils.SchemaStruct, file string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchLoadEkServices, pathchFileNameofRuntime, pathchSaveSchemaStructToYamlFile}
	}

	type args struct {
		epParams *epapiplugins.EpParams
	}
	tests := []struct {
		name               string
		args               args
		expectErrorContent error
		funcBeforeTest     func(*gomock.Controller) []*mpatch.Patch
	}{
		// TODO: Add test cases.
		{
			name: "load_kit_services_fail",
			args: args{
				&epapiplugins.EpParams{
					Kitconfig: &epapiplugins.Kitconfig{
						Components: &epapiplugins.KitconfigComponents{
							Manifests: []string{"test_component_manifest.yml"},
							Selector: []*plugins.KitconfigComponentsSelectorItems0{
								{
									Name:         "test",
									OverrideYaml: "test",
								},
							},
						},
					},
					Kitconfigpath: "/home/user/testPath",
				},
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_load_services_fail,
		},
		{
			name: "FileNameofRuntime_fail",
			args: args{
				&epapiplugins.EpParams{
					Kitconfig: &epapiplugins.Kitconfig{
						Components: &epapiplugins.KitconfigComponents{
							Manifests: []string{"test_component_manifest.yml"},
							Selector: []*plugins.KitconfigComponentsSelectorItems0{
								{
									Name:         "test",
									OverrideYaml: "test",
								},
							},
						},
					},
					Kitconfigpath: "/home/user/testPath",
				},
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_FileNameofRuntime_fail,
		},
		{
			name: "SaveSchemaStructToYamlFile_fail",
			args: args{
				&epapiplugins.EpParams{
					Kitconfig: &epapiplugins.Kitconfig{
						Components: &epapiplugins.KitconfigComponents{
							Manifests: []string{"test_component_manifest.yml"},
							Selector: []*plugins.KitconfigComponentsSelectorItems0{
								{
									Name:         "test",
									OverrideYaml: "test",
								},
							},
						},
					},
					Kitconfigpath: "/home/user/testPath",
				},
			},
			expectErrorContent: errTest,
			funcBeforeTest:     func_SaveSchemaStructToYamlFile_fail,
		},
		{
			name: "everyFunc_ok",
			args: args{
				&epapiplugins.EpParams{
					Kitconfig: &epapiplugins.Kitconfig{
						Components: &epapiplugins.KitconfigComponents{
							Manifests: []string{"test_component_manifest.yml"},
							Selector: []*plugins.KitconfigComponentsSelectorItems0{
								{
									Name:         "test",
									OverrideYaml: "test",
								},
							},
						},
					},
					Kitconfigpath: "/home/user/testPath",
				},
			},
			expectErrorContent: nil,
			funcBeforeTest:     func_everyFunc_ok,
		},
	}
	for _, tc := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var plist []*mpatch.Patch
		if tc.funcBeforeTest != nil {
			plist = tc.funcBeforeTest(ctrl)
		}

		t.Run(tc.name, func(t *testing.T) {
			// Run test cases in parallel if necessary.
			if result := EpwfLoadServices(tc.args.epParams); result != nil {
				if fmt.Sprint(result) == fmt.Sprint(tc.expectErrorContent) {
					t.Log(tc.name, "Done")
				} else {
					t.Errorf("%s error %s", tc.name, result)
				}
			} else {
				t.Log(tc.name, "Done")
			}
		})

		for _, p := range plist {
			unpatch(t, p)
		}
	}
}
