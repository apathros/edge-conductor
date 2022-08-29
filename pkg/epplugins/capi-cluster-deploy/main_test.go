/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package capiclusterdeploy

import (
	"errors"
	"os/exec"
	"reflect"
	"testing"
	"time"

	pluginapi "ep/pkg/api/plugins"
	"ep/pkg/eputils"
	"ep/pkg/eputils/capiutils"
	"ep/pkg/eputils/kubeutils"
	kubemock "ep/pkg/eputils/kubeutils/mock"
	serviceutils "ep/pkg/eputils/service"
	servicemock "ep/pkg/eputils/service/mock"
	"ep/pkg/eputils/test/fakekubeutils"
	"ep/pkg/eputils/test/fakeserviceutils"

	"github.com/golang/mock/gomock"
	"github.com/undefinedlabs/go-mpatch"
)

var (
	capiClusterDeployError = errors.New("capi cluster deploy fail")

	machineStatus = "metal3-cluster-control-plane-qjbfs   metal3    metal3-cluster-control-plan  Running\n"
	kubeconfig    = map[string][]byte{
		"data": []byte("kubeconfig"),
	}

	epParamNoProvider = []byte(`{"kitconfig": {"Parameters": {"global_settings": {}, "extensions": [""]}}, "runtimedir": "", "workspace": ""}`)
	epParam           = []byte(`{"kitconfig": {"Parameters": {"global_settings": {}, "extensions": ["capi-metal3", ""]}, "Cluster": {"config": "aa"}}, "runtimedir": "testruntime", "workspace": "testworkspace"}`)
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPluginMain(t *testing.T) {

	func_capi_cluster_config_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
			return capiClusterDeployError
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}

		return []*mpatch.Patch{p1}
	}

	func_get_capi_template_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p2, err := mpatch.PatchMethod(capiutils.GetCapiSetting, func(epparams *pluginapi.EpParams, clusterManifest *pluginapi.Clustermanifest, clusterConfig *pluginapi.CapiClusterConfig, setting *pluginapi.CapiSetting) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p3, err := mpatch.PatchMethod(capiutils.GetCapiTemplate, func(*pluginapi.EpParams, pluginapi.CapiSetting, *capiutils.CapiTemplate) error {
			return capiClusterDeployError
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}

		return []*mpatch.Patch{p1, p2, p3}
	}

	func_cluster_rendering_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p2, err := mpatch.PatchMethod(capiutils.GetCapiSetting, func(epparams *pluginapi.EpParams, clusterManifest *pluginapi.Clustermanifest, clusterConfig *pluginapi.CapiClusterConfig, setting *pluginapi.CapiSetting) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p3, err := mpatch.PatchMethod(capiutils.GetCapiTemplate, func(*pluginapi.EpParams, pluginapi.CapiSetting, *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p4, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(*capiutils.CapiTemplate, string, string, string) error {
			return capiClusterDeployError
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}

		return []*mpatch.Patch{p1, p2, p3, p4}
	}

	func_cluster_apply_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p2, err := mpatch.PatchMethod(capiutils.GetCapiSetting, func(epparams *pluginapi.EpParams, clusterManifest *pluginapi.Clustermanifest, clusterConfig *pluginapi.CapiClusterConfig, setting *pluginapi.CapiSetting) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p3, err := mpatch.PatchMethod(capiutils.GetCapiTemplate, func(*pluginapi.EpParams, pluginapi.CapiSetting, *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p4, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(*capiutils.CapiTemplate, string, string, string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}

		fakeYamlDeployer := fakeserviceutils.FakeYamlDeployer{}
		mockServiceDeployer := servicemock.NewMockServiceDeployer(ctrl)
		p5, err := mpatch.PatchMethod(serviceutils.NewYamlDeployer, mockServiceDeployer.NewYamlDeployer)
		if err != nil {
			t.Fatal(err)
		}
		mockServiceDeployer.EXPECT().
			NewYamlDeployer(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeYamlDeployer)
		p6, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeYamlDeployer), "YamlInstall", func(*fakeserviceutils.FakeYamlDeployer, string) error {
			return capiClusterDeployError
		})
		if err != nil {
			t.Fatal(err)
		}

		return []*mpatch.Patch{p1, p2, p3, p4, p5, p6}
	}

	func_get_machine_exec := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p2, err := mpatch.PatchMethod(capiutils.GetCapiSetting, func(epparams *pluginapi.EpParams, clusterManifest *pluginapi.Clustermanifest, clusterConfig *pluginapi.CapiClusterConfig, setting *pluginapi.CapiSetting) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p3, err := mpatch.PatchMethod(capiutils.GetCapiTemplate, func(*pluginapi.EpParams, pluginapi.CapiSetting, *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p4, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(*capiutils.CapiTemplate, string, string, string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}

		fakeYamlDeployer := fakeserviceutils.FakeYamlDeployer{}
		mockServiceDeployer := servicemock.NewMockServiceDeployer(ctrl)
		p5, err := mpatch.PatchMethod(serviceutils.NewYamlDeployer, mockServiceDeployer.NewYamlDeployer)
		if err != nil {
			t.Fatal(err)
		}
		mockServiceDeployer.EXPECT().
			NewYamlDeployer(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeYamlDeployer)
		p6, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeYamlDeployer), "YamlInstall", func(*fakeserviceutils.FakeYamlDeployer, string) error {
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		p7, err := mpatch.PatchMethod(eputils.RunCMD, func(*exec.Cmd) (string, error) { return "", capiClusterDeployError })
		if err != nil {
			t.Fatal(err)
		}

		return []*mpatch.Patch{p1, p2, p3, p4, p5, p6, p7}
	}

	func_no_machine_ready := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p2, err := mpatch.PatchMethod(capiutils.GetCapiSetting, func(epparams *pluginapi.EpParams, clusterManifest *pluginapi.Clustermanifest, clusterConfig *pluginapi.CapiClusterConfig, setting *pluginapi.CapiSetting) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p3, err := mpatch.PatchMethod(capiutils.GetCapiTemplate, func(*pluginapi.EpParams, pluginapi.CapiSetting, *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p4, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(*capiutils.CapiTemplate, string, string, string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}

		fakeYamlDeployer := fakeserviceutils.FakeYamlDeployer{}
		mockServiceDeployer := servicemock.NewMockServiceDeployer(ctrl)
		p5, err := mpatch.PatchMethod(serviceutils.NewYamlDeployer, mockServiceDeployer.NewYamlDeployer)
		if err != nil {
			t.Fatal(err)
		}
		mockServiceDeployer.EXPECT().
			NewYamlDeployer(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeYamlDeployer)
		p6, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeYamlDeployer), "YamlInstall", func(*fakeserviceutils.FakeYamlDeployer, string) error {
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		p7, err := mpatch.PatchMethod(eputils.RunCMD, func(*exec.Cmd) (string, error) { return "", nil })
		if err != nil {
			t.Fatal(err)
		}
		p8, err := mpatch.PatchMethod(time.Sleep, func(d time.Duration) {})
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p1, p2, p3, p4, p5, p6, p7, p8}
	}

	func_new_secret_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p2, err := mpatch.PatchMethod(capiutils.GetCapiSetting, func(epparams *pluginapi.EpParams, clusterManifest *pluginapi.Clustermanifest, clusterConfig *pluginapi.CapiClusterConfig, setting *pluginapi.CapiSetting) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p3, err := mpatch.PatchMethod(capiutils.GetCapiTemplate, func(*pluginapi.EpParams, pluginapi.CapiSetting, *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p4, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(*capiutils.CapiTemplate, string, string, string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}

		fakeYamlDeployer := fakeserviceutils.FakeYamlDeployer{}
		mockServiceDeployer := servicemock.NewMockServiceDeployer(ctrl)
		p5, err := mpatch.PatchMethod(serviceutils.NewYamlDeployer, mockServiceDeployer.NewYamlDeployer)
		if err != nil {
			t.Fatal(err)
		}
		mockServiceDeployer.EXPECT().
			NewYamlDeployer(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeYamlDeployer)
		p6, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeYamlDeployer), "YamlInstall", func(*fakeserviceutils.FakeYamlDeployer, string) error {
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		p7, err := mpatch.PatchMethod(eputils.RunCMD, func(*exec.Cmd) (string, error) { return machineStatus, nil })
		if err != nil {
			t.Fatal(err)
		}
		p8, err := mpatch.PatchMethod(time.Sleep, func(d time.Duration) {})
		if err != nil {
			t.Fatal(err)
		}
		p9, err := mpatch.PatchMethod(kubeutils.NewSecret, func(string, string, string, string) (kubeutils.SecretWrapper, error) {
			return nil, capiClusterDeployError
		})
		if err != nil {
			t.Fatal(err)
		}

		return []*mpatch.Patch{p1, p2, p3, p4, p5, p6, p7, p8, p9}
	}

	func_kubeconfig_get_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p2, err := mpatch.PatchMethod(capiutils.GetCapiSetting, func(epparams *pluginapi.EpParams, clusterManifest *pluginapi.Clustermanifest, clusterConfig *pluginapi.CapiClusterConfig, setting *pluginapi.CapiSetting) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p3, err := mpatch.PatchMethod(capiutils.GetCapiTemplate, func(*pluginapi.EpParams, pluginapi.CapiSetting, *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p4, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(*capiutils.CapiTemplate, string, string, string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}

		fakeYamlDeployer := fakeserviceutils.FakeYamlDeployer{}
		mockServiceDeployer := servicemock.NewMockServiceDeployer(ctrl)
		p5, err := mpatch.PatchMethod(serviceutils.NewYamlDeployer, mockServiceDeployer.NewYamlDeployer)
		if err != nil {
			t.Fatal(err)
		}
		mockServiceDeployer.EXPECT().
			NewYamlDeployer(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeYamlDeployer)
		p6, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeYamlDeployer), "YamlInstall", func(*fakeserviceutils.FakeYamlDeployer, string) error {
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		p7, err := mpatch.PatchMethod(eputils.RunCMD, func(*exec.Cmd) (string, error) { return machineStatus, nil })
		if err != nil {
			t.Fatal(err)
		}
		p8, err := mpatch.PatchMethod(time.Sleep, func(d time.Duration) {})
		if err != nil {
			t.Fatal(err)
		}

		fakeSecret := fakekubeutils.FakeSecret{}
		mockKubeClientWrapper := kubemock.NewMockKubeClientWrapper(ctrl)
		p9, err := mpatch.PatchMethod(kubeutils.NewSecret, mockKubeClientWrapper.NewSecret)
		if err != nil {
			t.Fatal(err)
		}
		mockKubeClientWrapper.EXPECT().NewSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&fakeSecret, nil)

		p10, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeSecret), "Get", func(*fakekubeutils.FakeSecret) error {
			return capiClusterDeployError
		})
		if err != nil {
			t.Fatal(err)
		}

		return []*mpatch.Patch{p1, p2, p3, p4, p5, p6, p7, p8, p9, p10}
	}

	func_cluster_deply_succeed := func(ctrl *gomock.Controller) []*mpatch.Patch {
		p1, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p2, err := mpatch.PatchMethod(capiutils.GetCapiSetting, func(epparams *pluginapi.EpParams, clusterManifest *pluginapi.Clustermanifest, clusterConfig *pluginapi.CapiClusterConfig, setting *pluginapi.CapiSetting) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p3, err := mpatch.PatchMethod(capiutils.GetCapiTemplate, func(*pluginapi.EpParams, pluginapi.CapiSetting, *capiutils.CapiTemplate) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}
		p4, err := mpatch.PatchMethod(capiutils.TmplFileRendering, func(*capiutils.CapiTemplate, string, string, string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
			return nil
		}

		fakeYamlDeployer := fakeserviceutils.FakeYamlDeployer{}
		mockServiceDeployer := servicemock.NewMockServiceDeployer(ctrl)
		p5, err := mpatch.PatchMethod(serviceutils.NewYamlDeployer, mockServiceDeployer.NewYamlDeployer)
		if err != nil {
			t.Fatal(err)
		}
		mockServiceDeployer.EXPECT().
			NewYamlDeployer(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeYamlDeployer)
		p6, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeYamlDeployer), "YamlInstall", func(*fakeserviceutils.FakeYamlDeployer, string) error {
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		p7, err := mpatch.PatchMethod(eputils.RunCMD, func(*exec.Cmd) (string, error) { return machineStatus, nil })
		if err != nil {
			t.Fatal(err)
		}
		p8, err := mpatch.PatchMethod(time.Sleep, func(d time.Duration) {})
		if err != nil {
			t.Fatal(err)
		}

		fakeSecret := fakekubeutils.FakeSecret{}
		mockKubeClientWrapper := kubemock.NewMockKubeClientWrapper(ctrl)
		p9, err := mpatch.PatchMethod(kubeutils.NewSecret, mockKubeClientWrapper.NewSecret)
		if err != nil {
			t.Fatal(err)
		}
		mockKubeClientWrapper.EXPECT().NewSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&fakeSecret, nil)

		p10, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeSecret), "Get", func(*fakekubeutils.FakeSecret) error {
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		p11, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakeSecret), "GetData", func(*fakekubeutils.FakeSecret) map[string][]byte {
			return kubeconfig
		})
		if err != nil {
			t.Fatal(err)
		}

		return []*mpatch.Patch{p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11}
	}

	cases := []struct {
		name                  string
		input, expectedOutput map[string][]byte
		expectError           bool
		funcBeforeTest        func(*gomock.Controller) []*mpatch.Patch
	}{
		// TODO: Add the values to complete your test cases.
		// Add the values for input and expectedoutput with particular struct marshal data in json format.
		// They will be used to generate "SchemaMapData" as inputs and expected outputs of plugins under test.
		// if the inputs in the Plugin Input List is not required in your test case, keep the value as nil.
		{
			name: "No legal extension",
			input: map[string][]byte{
				"ep-params":        epParamNoProvider,
				"cluster-manifest": nil,
			},

			expectedOutput: map[string][]byte{

				"cluster": nil,
			},
			expectError:    true,
			funcBeforeTest: nil,
		},
		{
			name: "Load capi cluster config fail",
			input: map[string][]byte{
				"ep-params":        epParam,
				"cluster-manifest": nil,
			},

			expectedOutput: map[string][]byte{

				"cluster": nil,
			},
			expectError:    true,
			funcBeforeTest: func_capi_cluster_config_fail,
		},
		{
			name: "Get capi template fail",
			input: map[string][]byte{
				"ep-params":        epParam,
				"cluster-manifest": nil,
			},

			expectedOutput: map[string][]byte{

				"cluster": nil,
			},
			expectError:    true,
			funcBeforeTest: func_get_capi_template_fail,
		},
		{
			name: "Cluster rendering fail",
			input: map[string][]byte{
				"ep-params":        epParam,
				"cluster-manifest": nil,
			},

			expectedOutput: map[string][]byte{

				"cluster": nil,
			},
			expectError:    true,
			funcBeforeTest: func_cluster_rendering_fail,
		},
		{
			name: "Cluster apply fail",
			input: map[string][]byte{
				"ep-params":        epParam,
				"cluster-manifest": nil,
			},

			expectedOutput: map[string][]byte{

				"cluster": nil,
			},
			expectError:    true,
			funcBeforeTest: func_cluster_apply_fail,
		},
		{
			name: "Execute get machine",
			input: map[string][]byte{
				"ep-params":        epParam,
				"cluster-manifest": nil,
			},

			expectedOutput: map[string][]byte{

				"cluster": nil,
			},
			expectError:    true,
			funcBeforeTest: func_get_machine_exec,
		},
		{
			name: "No machine available",
			input: map[string][]byte{
				"ep-params":        epParam,
				"cluster-manifest": nil,
			},

			expectedOutput: map[string][]byte{

				"cluster": nil,
			},
			expectError:    true,
			funcBeforeTest: func_no_machine_ready,
		},
		{
			name: "New secret fail",
			input: map[string][]byte{
				"ep-params":        epParam,
				"cluster-manifest": nil,
			},

			expectedOutput: map[string][]byte{

				"cluster": nil,
			},
			expectError:    true,
			funcBeforeTest: func_new_secret_fail,
		},
		{
			name: "Kubeconfig_get_fail",
			input: map[string][]byte{
				"ep-params":        epParam,
				"cluster-manifest": nil,
			},

			expectedOutput: map[string][]byte{

				"cluster": nil,
			},
			expectError:    true,
			funcBeforeTest: func_kubeconfig_get_fail,
		},
		{
			name: "Cluster deploy",
			input: map[string][]byte{
				"ep-params":        epParam,
				"cluster-manifest": nil,
			},

			expectedOutput: map[string][]byte{

				"cluster": []byte(`{"content": "kubeconfig"}`),
			},
			expectError:    false,
			funcBeforeTest: func_cluster_deply_succeed,
		},
	}
	// Optional: add setup for the test series
	for _, tc := range cases {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var plist []*mpatch.Patch
		if tc.funcBeforeTest != nil {
			plist = tc.funcBeforeTest(ctrl)
		}

		t.Run(tc.name, func(t *testing.T) {
			// Run test cases in parallel if necessary.
			// t.Parallel()

			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}

			testOutput := generateOutput(nil)
			expectedOutput := generateOutput(tc.expectedOutput)

			if result := PluginMain(input, &testOutput); result != nil {
				if tc.expectError {
					t.Log("Error expected.")
					return
				} else {
					t.Logf("Failed to run PluginMain when input is %s.", tc.input)
					t.Error(result)
				}
			}

			if testOutput.EqualWith(expectedOutput) {
				t.Log("Output expected.")
			} else {
				t.Errorf("Failed to get expected output when input is %s.", tc.input)
			}
		})

		for _, p := range plist {
			unpatch(t, p)
		}
	}

	// Optional: add teardown for the test series
}

func TestSchemaStruct(t *testing.T) {
	cases := []struct {
		name, input string
	}{
		{
			name:  "ep-params",
			input: __name("ep-params"),
		},
		{name: "cluster-manifest",
			input: __name("cluster-manifest")}, {
			name:  "kubeconfig",
			input: __name("kubeconfig"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eputils.SchemaStructNew(tc.input)
		})

	}
}
