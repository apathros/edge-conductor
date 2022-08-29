/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

//nolint: dupl
package servicedeployer

import (
	epplugins "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	kubeutils "ep/pkg/eputils/kubeutils"
	kubemock "ep/pkg/eputils/kubeutils/mock"
	repoutils "ep/pkg/eputils/repoutils"
	repomock "ep/pkg/eputils/repoutils/mock"
	serviceutil "ep/pkg/eputils/service"
	servicemock "ep/pkg/eputils/service/mock"
	fakekubeutils "ep/pkg/eputils/test/fakekubeutils"
	"errors"
	"fmt"
	"reflect"
	"testing"

	gomock "github.com/golang/mock/gomock"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	kubeerr     = errors.New("kubernetes error")
	errNotFound = errors.New("NotFound")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

type fakeDeployer struct {
	name string
	rev  int
}

func (h *fakeDeployer) GetName() string {
	return h.name
}
func (h *fakeDeployer) YamlInstall(loc_kubeconfig string) error {
	return nil
}
func (h *fakeDeployer) YamlUninstall(loc_kubeconfig string) error {
	return errNotFound
}

func (h *fakeDeployer) HelmStatus(loc_kubeconfig string) (string, int) {
	//log.Infoln(h.name, h.rev)
	return h.name, h.rev
}
func (h *fakeDeployer) HelmInstall(loc_kubeconfig string, arg ...serviceutil.InstallOpt) error {
	h.rev = h.rev + 1
	return nil
}
func (h *fakeDeployer) HelmUpgrade(loc_kubeconfig string) error {
	h.rev = h.rev + 1
	return nil
}
func (h *fakeDeployer) HelmUninstall(loc_kubeconfig string) error {
	h.name = serviceutil.HELM_STATUS_NOT_DEPLOYED
	h.rev = 0
	return nil
}

func TestPluginMain(t *testing.T) {

	func_configmap_getYamlData := func(_ *fakekubeutils.FakeConfigMap) map[string]string {
		return map[string]string{
			"testyaml":  `{"supported-clusters":["default"],"name":"testyaml","type":"yaml","url":"file://u","namespace":"ns"}`,
			"testyaml1": `{"supported-clusters":["default"],"name":"testyaml1","type":"yaml"}`,
		}
	}
	func_configmap_getHelmData := func(_ *fakekubeutils.FakeConfigMap) map[string]string {
		return map[string]string{
			"testhelm":  `{"supported-clusters":["default"],"name":"testhelm","type":"helm","chartoverride":"file://o","url":"file://u","revision":"1","namespace":"ns"}`,
			"testhelm2": `{"supported-clusters":["default"],"name":"testhelm2","type":"helm","chartoverride":"file://o","url":"file://u","revision":"1"}`,
			"testhelm3": `{"supported-clusters":["default"],"name":"testhelm3","type":"helm","chartoverride":"file://o","url":"file://u","revision":"0"}`,
			"testhelm4": `{"supported-clusters":["default"],"name":"testhelm4","type":"helm","chartoverride":"","url":"file://u","revision":"2"}`,
			"testhelm5": `{"supported-clusters":["default"],"name":"testhelm5","type":"helm","chartoverride":"file://o","url":"","revision":"1"}`,
		}
	}

	func_configmap_get := func(_ *fakekubeutils.FakeConfigMap) error {
		return errNotFound
	}

	func_err_CreateNamespace := func(ctrl *gomock.Controller, i int) []*mpatch.Patch {
		p, err := mpatch.PatchMethod(kubeutils.CreateNamespace, func(string, string) error { return kubeerr })
		if err != nil {
			t.Fatal(err)
		}
		return []*mpatch.Patch{p}
	}

	func_err_NewConfigMap := func(ctrl *gomock.Controller, i int) []*mpatch.Patch {
		mockKubeClientWrapper := kubemock.NewMockKubeClientWrapper(ctrl)
		patchCreateNamespace, err := mpatch.PatchMethod(kubeutils.CreateNamespace, mockKubeClientWrapper.CreateNamespace)
		if err != nil {
			t.Fatal(err)
		}
		patchNewConfigMap, err := mpatch.PatchMethod(kubeutils.NewConfigMap, mockKubeClientWrapper.NewConfigMap)
		if err != nil {
			t.Fatal(err)
		}
		// CreateNamespace always ok
		mockKubeClientWrapper.EXPECT().
			CreateNamespace(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		// NewConfigMap return error
		mockKubeClientWrapper.EXPECT().
			NewConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, kubeerr)

		return []*mpatch.Patch{patchCreateNamespace, patchNewConfigMap}
	}

	func_kubeClient_Yaml_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		mockKubeClientWrapper := kubemock.NewMockKubeClientWrapper(ctrl)
		p1, err := mpatch.PatchMethod(kubeutils.CreateNamespace, mockKubeClientWrapper.CreateNamespace)
		if err != nil {
			t.Fatal(err)
		}
		p2, err := mpatch.PatchMethod(kubeutils.NewConfigMap, mockKubeClientWrapper.NewConfigMap)
		if err != nil {
			t.Fatal(err)
		}
		// CreateNamespace always ok
		mockKubeClientWrapper.EXPECT().
			CreateNamespace(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		// NewConfigMap return error
		fakecm := fakekubeutils.FakeConfigMap{}
		p3, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakecm), "GetData", func_configmap_getYamlData)
		if err != nil {
			t.Fatal(err)
		}
		p4, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakecm), "Get", func_configmap_get)
		if err != nil {
			t.Fatal(err)
		}
		mockKubeClientWrapper.EXPECT().
			NewConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&fakecm, nil)

		return []*mpatch.Patch{p1, p2, p3, p4}
	}

	func_kubeClient_Helm_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		mockKubeClientWrapper := kubemock.NewMockKubeClientWrapper(ctrl)
		p1, err := mpatch.PatchMethod(kubeutils.CreateNamespace, mockKubeClientWrapper.CreateNamespace)
		if err != nil {
			t.Fatal(err)
		}
		p2, err := mpatch.PatchMethod(kubeutils.NewConfigMap, mockKubeClientWrapper.NewConfigMap)
		if err != nil {
			t.Fatal(err)
		}
		// CreateNamespace always ok
		mockKubeClientWrapper.EXPECT().
			CreateNamespace(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		// NewConfigMap return error
		fakecm := fakekubeutils.FakeConfigMap{}
		p3, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakecm), "GetData", func_configmap_getHelmData)
		if err != nil {
			t.Fatal(err)
		}
		p4, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakecm), "Get", func_configmap_get)
		if err != nil {
			t.Fatal(err)
		}
		mockKubeClientWrapper.EXPECT().
			NewConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&fakecm, nil)

		return []*mpatch.Patch{p1, p2, p3, p4}
	}

	func_err_NewServiceConfigMap := func(ctrl *gomock.Controller, i int) []*mpatch.Patch {
		plist := func_kubeClient_Yaml_ok(ctrl)
		fakecm := fakekubeutils.FakeConfigMap{}
		patchNewConfigMap, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakecm), "New",
			func(*fakekubeutils.FakeConfigMap) error { return kubeerr })
		if err != nil {
			t.Fatal("patch error:", err)
		}
		plist = append(plist, patchNewConfigMap)
		return plist
	}

	func_err_LoadSchemaStructFromYaml := func(ctrl *gomock.Controller, i int) []*mpatch.Patch {
		plist := func_kubeClient_Yaml_ok(ctrl)

		patchLoadSchemaStructFromYaml, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYaml, func(eputils.SchemaStruct, string) error {
			return kubeerr
		})
		if err != nil {
			t.Fatal("patch error:", err)
		}
		plist = append(plist, patchLoadSchemaStructFromYaml)
		return plist
	}

	func_err_YamlPullFileFromRepo := func(ctrl *gomock.Controller, i int) []*mpatch.Patch {
		plist := func_kubeClient_Yaml_ok(ctrl)

		// Repo Utils
		p1, err := mpatch.PatchMethod(repoutils.PullFileFromRepo, func(string, string) error { return kubeerr })
		if err != nil {
			t.Fatal(err)
		}
		plist = append(plist, p1)
		return plist
	}

	func_yaml_success_case := func(ctrl *gomock.Controller, i int) []*mpatch.Patch {
		plist := func_kubeClient_Yaml_ok(ctrl)

		// Repo Utils
		p1, err := mpatch.PatchMethod(repoutils.PullFileFromRepo, func(string, string) error { return nil })
		if err != nil {
			t.Fatal(err)
		}

		// Service Utils -  Helm Deployer
		mockServiceDeployer := servicemock.NewMockServiceDeployer(ctrl)
		p2, err := mpatch.PatchMethod(serviceutil.NewYamlDeployer, mockServiceDeployer.NewYamlDeployer)
		if err != nil {
			t.Fatal(err)
		}
		mockServiceDeployer.EXPECT().
			NewYamlDeployer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeDeployer{name: "yamlservice"})

		plist = append(plist, p1)
		plist = append(plist, p2)
		return plist
	}

	func_err_HelmPullFileFromRepo := func(ctrl *gomock.Controller, i int) []*mpatch.Patch {
		plist := func_kubeClient_Helm_ok(ctrl)
		// Repo Utils
		mockRepoWrapper := repomock.NewMockRepoUtilsInterface(ctrl)
		patchPullFileFromRepo, err := mpatch.PatchMethod(repoutils.PullFileFromRepo, mockRepoWrapper.PullFileFromRepo)
		if err != nil {
			t.Fatal("Patch err:", err)
		}
		for j := 0; j < i; j++ {
			mockRepoWrapper.EXPECT().PullFileFromRepo(gomock.Any(), gomock.Any()).Return(nil)
			patchGenSHA256, err := mpatch.PatchMethod(eputils.GenFileSHA256, func(string) (string, error) { return "", nil })
			if err != nil {
				t.Fatal("Patch err:", err)
			}
			plist = append(plist, patchGenSHA256)
		}
		mockRepoWrapper.EXPECT().PullFileFromRepo(gomock.Any(), gomock.Any()).Return(kubeerr)

		plist = append(plist, patchPullFileFromRepo)
		return plist
	}

	func_err_HelmStatus := func(ctrl *gomock.Controller, i int) []*mpatch.Patch {
		plist := func_kubeClient_Helm_ok(ctrl)

		// Repo Utils
		patchPullFileFromRepo, err := mpatch.PatchMethod(repoutils.PullFileFromRepo, func(string, string) error { return nil })
		if err != nil {
			t.Fatal("patch error:", err)
		}

		// Gen service hash
		patchGenSHA256, err := mpatch.PatchMethod(eputils.GenFileSHA256, func(string) (string, error) { return "", nil })
		if err != nil {
			t.Fatal("patch error:", err)
		}

		// Service Utils -  Helm Deployer
		mockServiceDeployer := servicemock.NewMockServiceDeployer(ctrl)
		patchNewHelmDeployer, err := mpatch.PatchMethod(serviceutil.NewHelmDeployer, mockServiceDeployer.NewHelmDeployer)
		if err != nil {
			t.Fatal("patch error:", err)
		}
		mockServiceDeployer.EXPECT().
			NewHelmDeployer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeDeployer{name: serviceutil.HELM_STATUS_UNKNOWN, rev: 1})

		plist = append(plist, patchPullFileFromRepo)
		plist = append(plist, patchGenSHA256)
		plist = append(plist, patchNewHelmDeployer)
		return plist
	}

	func_err_HelmInstall := func(ctrl *gomock.Controller, i int) []*mpatch.Patch {
		plist := func_kubeClient_Helm_ok(ctrl)

		// Repo Utils
		patchPullFileFromRepo, err := mpatch.PatchMethod(repoutils.PullFileFromRepo, func(string, string) error { return nil })
		if err != nil {
			t.Fatal("patch error:", err)
		}

		// Gen service hash
		patchGenSHA256, err := mpatch.PatchMethod(eputils.GenFileSHA256, func(string) (string, error) { return "", nil })
		if err != nil {
			t.Fatal("patch error:", err)
		}

		// Service Utils -  Helm Deployer
		mockServiceDeployer := servicemock.NewMockServiceDeployer(ctrl)
		patchNewHelmDeployer, err := mpatch.PatchMethod(serviceutil.NewHelmDeployer, mockServiceDeployer.NewHelmDeployer)
		if err != nil {
			t.Fatal("patch error:", err)
		}
		mockServiceDeployer.EXPECT().
			NewHelmDeployer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeDeployer{name: serviceutil.HELM_STATUS_NOT_DEPLOYED, rev: 1})

		fakecm := fakeDeployer{}
		patchHelmInstall, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakecm), "HelmInstall",
			func(*fakeDeployer, string, ...serviceutil.InstallOpt) error { return kubeerr })
		if err != nil {
			t.Fatal("patch error:", err)
		}

		plist = append(plist, patchPullFileFromRepo)
		plist = append(plist, patchGenSHA256)
		plist = append(plist, patchNewHelmDeployer)
		plist = append(plist, patchHelmInstall)
		return plist
	}

	func_helm_success_case := func(ctrl *gomock.Controller, i int) []*mpatch.Patch {
		plist := func_kubeClient_Helm_ok(ctrl)

		// Repo Utils
		patchPullFileFromRepo, err := mpatch.PatchMethod(repoutils.PullFileFromRepo, func(string, string) error {
			return nil
		})
		if err != nil {
			t.Fatal("patch error:", err)
		}

		// Gen service hash
		patchGenSHA256, err := mpatch.PatchMethod(eputils.GenFileSHA256, func(string) (string, error) { return "", nil })
		if err != nil {
			t.Fatal("patch error:", err)
		}

		// Service Utils -  Helm Deployer
		mockServiceDeployer := servicemock.NewMockServiceDeployer(ctrl)
		patchNewHelmDeployer, err := mpatch.PatchMethod(serviceutil.NewHelmDeployer, mockServiceDeployer.NewHelmDeployer)
		if err != nil {
			t.Fatal("patch error:", err)
		}

		mockServiceDeployer.EXPECT().
			NewHelmDeployer("testhelm", gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeDeployer{name: "Deployed", rev: 1})
		mockServiceDeployer.EXPECT().
			NewHelmDeployer("testhelm1", gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeDeployer{name: serviceutil.HELM_STATUS_NOT_DEPLOYED, rev: 0})
		mockServiceDeployer.EXPECT().
			NewHelmDeployer("testhelm2", gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeDeployer{name: "Deployed", rev: 5})
		mockServiceDeployer.EXPECT().
			NewHelmDeployer("testhelm3", gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeDeployer{name: serviceutil.HELM_STATUS_NOT_DEPLOYED, rev: 0})
		mockServiceDeployer.EXPECT().
			NewHelmDeployer("testhelm4", gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeDeployer{name: "Deployed", rev: 5})
		mockServiceDeployer.EXPECT().
			NewHelmDeployer("testhelm5", gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(&fakeDeployer{name: "Deployed", rev: 1})

		plist = append(plist, patchPullFileFromRepo)
		plist = append(plist, patchGenSHA256)
		plist = append(plist, patchNewHelmDeployer)
		return plist
	}

	cases := []struct {
		name           string
		input          map[string][]byte
		expectedErr    error
		funcBeforeTest func(*gomock.Controller, int) []*mpatch.Patch
		argument       int
	}{
		{
			name: "err_test_create_namespace_fail",
			input: map[string][]byte{
				"ep-params":     nil,
				"serviceconfig": nil,
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_CreateNamespace,
		},
		{
			name: "err_test_create_configmap_fail",
			input: map[string][]byte{
				"ep-params":     nil,
				"serviceconfig": nil,
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_NewConfigMap,
		},
		{
			name: "err_test_new_ServiceConfigMap_fail",
			input: map[string][]byte{
				"ep-params":     nil,
				"serviceconfig": nil,
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_NewServiceConfigMap,
		},
		{
			name: "err_test_loadSchemaStructFromYaml_fail",
			input: map[string][]byte{
				"ep-params":     nil,
				"serviceconfig": nil,
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_LoadSchemaStructFromYaml,
		},
		{
			name: "err_test_yaml_pull_file_fail0",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": nil,
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_YamlPullFileFromRepo,
		},
		{
			name: "err_test_yaml_pull_file_fail1",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"Components":[{"name":"testyaml","type":"yaml","url":"file://u","namespace":"ns"},{"name":"testyaml1"}]}`),
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_YamlPullFileFromRepo,
		},
		{
			name: "test_yaml_success",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"Components":[{"supported-clusters":["default"],"name":"testyaml2","type":"yaml","url":"file://u","namespace":"ns"},{"supported-clusters":["default"],"name":"testyaml3","type":"yaml"}]}`),
			},
			expectedErr:    nil,
			funcBeforeTest: func_yaml_success_case,
		},
		// Helm
		{
			name: "err_test_helm_pull_file_fail0",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"Components":[{"name":"testhelm2"},{"name":"testhelm3"},{"name":"testhelm4"},{"name":"testhelm5"}]}`),
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_HelmPullFileFromRepo,
			argument:       0,
		},
		{
			name: "err_test_helm_pull_file_fail1",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"Components":[{"name":"testhelm2"},{"name":"testhelm3"},{"name":"testhelm4"},{"name":"testhelm5"}]}`),
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_HelmPullFileFromRepo,
			argument:       1,
		},
		{
			name: "err_test_helm_pull_file_fail2",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"Components":[{"supported-clusters":["default"],"name":"testhelm","type":"helm","chartoverride":"o","url":"u"},{"name":"testhelm2"},{"name":"testhelm3"},{"name":"testhelm4"},{"name":"testhelm5"}]}`),
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_HelmPullFileFromRepo,
			argument:       0,
		},
		{
			name: "err_test_helm_pull_file_fail3",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"Components":[{"supported-clusters":["default"],"name":"testhelm","type":"helm","chartoverride":"o","url":"u"},{"name":"testhelm2"},{"name":"testhelm3"},{"name":"testhelm4"},{"name":"testhelm5"}]}`),
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_HelmPullFileFromRepo,
			argument:       1,
		},
		{
			name: "err_test_wrong_helm_status_remove",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimedir":"","kubeconfig":""}`),
			},
			expectedErr:    eputils.GetError("errUnknownStatus"),
			funcBeforeTest: func_err_HelmStatus,
		},
		{
			name: "err_test_wrong_helm_status_install_fail",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"Components":[{"supported-clusters":["default"],"name":"testhelm","type":"helm","chartoverride":"o","url":"u"},{"name":"testhelm2"},{"name":"testhelm3"},{"name":"testhelm4"},{"name":"testhelm5"}]}`),
			},
			expectedErr:    eputils.GetError("errUnknownStatus"),
			funcBeforeTest: func_err_HelmStatus,
		},
		{
			name: "err_test_wrong_helm_status_install_fail1",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"Components":[{"supported-clusters":["default"],"name":"testhelm","type":"helm","chartoverride":"o","url":"u"},{"name":"testhelm2"},{"name":"testhelm3"},{"name":"testhelm4"},{"name":"testhelm5"}]}`),
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_HelmInstall,
		},
		{
			name: "test_helm_success",
			input: map[string][]byte{
				"ep-params": []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"Components":[{"supported-clusters":["default"],
				"name":"testhelm","chartoverride":"","type":"helm","url":"u","namespace":"ns"},
				{"supported-clusters":["default"],"name":"testhelm1","chartoverride":"o","type":"helm","url":""},
				{"supported-clusters":["default"],"name":"testhelm2","chartoverride":"o","type":"helm","url":"u"},
				{"supported-clusters":["default"],"name":"testrepo","type":"repo","url":"u"}]}`),
			},
			expectedErr:    nil,
			funcBeforeTest: func_helm_success_case,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest(ctrl, tc.argument)
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}

			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			testOutput := generateOutput(nil)
			err := PluginMain(input, &testOutput)

			if tc.expectedErr != nil {
				if err == nil {
					t.Error("Expected error but no error found.")
				} else {
					if fmt.Sprint(err) == fmt.Sprint(tc.expectedErr) {
						t.Log("Error expected.")
					} else {
						t.Error("Expect:", tc.expectedErr, "; But found:", err)
					}
				}
			} else {
				if err != nil {
					t.Error("Unexpected Error:", err)
				}
			}
		})
	}

}

func Test_findService(t *testing.T) {
	type args struct {
		serviceName   string
		serviceConfig *epplugins.Serviceconfig
	}
	tests := []struct {
		name        string
		args        args
		expectedErr *epplugins.Component
	}{
		{
			name:        "err_test_serviceConfig_fail",
			args:        args{"", nil},
			expectedErr: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := findService(tc.args.serviceName, tc.args.serviceConfig); !reflect.DeepEqual(got, tc.expectedErr) {
				t.Errorf("find service: %v, want %v", got, tc.expectedErr)
			} else {
				t.Log("Error expected.")
			}
		})
	}
}

func Test_getExpectedRevision(t *testing.T) {
	cases := []struct {
		name        string
		expectedErr string
	}{
		{
			name:        "err_test_getExpectedRevision_fail",
			expectedErr: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fakecm := fakekubeutils.FakeConfigMap{}
			mockConfigMapWrapper := kubemock.NewMockConfigMapWrapper(ctrl)
			patchGetData, err := mpatch.PatchMethod(fakecm.GetData, mockConfigMapWrapper.GetData)
			if err != nil {
				t.Fatal("patch error:", err)
			}
			defer unpatch(t, patchGetData)
			mockConfigMapWrapper.EXPECT().GetData().AnyTimes().Return(nil)

			patchLoadSchemaStructFromYaml, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYaml, func(eputils.SchemaStruct, string) error {
				return kubeerr
			})
			if err != nil {
				t.Fatal("patch error:", err)
			}
			defer unpatch(t, patchLoadSchemaStructFromYaml)

			if got := getExpectedRevision(&fakecm, ""); !reflect.DeepEqual(got, tc.expectedErr) {
				t.Errorf("get expected revision %v, want %v", got, tc.expectedErr)
			} else {
				t.Log("Error expected.")
			}
		})
	}
}

func Test_getExpectedChartHash(t *testing.T) {
	cases := []struct {
		name        string
		expectedErr string
	}{
		{
			name:        "err_test_getExpectedChartHash_fail",
			expectedErr: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fakecm := fakekubeutils.FakeConfigMap{}
			mockConfigMapWrapper := kubemock.NewMockConfigMapWrapper(ctrl)
			patchGetData, err := mpatch.PatchMethod(fakecm.GetData, mockConfigMapWrapper.GetData)
			if err != nil {
				t.Fatal("patch error:", err)
			}
			defer unpatch(t, patchGetData)
			mockConfigMapWrapper.EXPECT().GetData().AnyTimes().Return(nil)

			patch3, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYaml, func(eputils.SchemaStruct, string) error {
				return kubeerr
			})
			if err != nil {
				t.Fatal("patch error:", err)
			}
			defer unpatch(t, patch3)

			if got := getExpectedChartHash(&fakecm, ""); !reflect.DeepEqual(got, tc.expectedErr) {
				t.Errorf("get chart hash: %v, want %v", got, tc.expectedErr)
			} else {
				t.Log("Error expected.")
			}
		})
	}
}

func Test_getExpectedOverrideHash(t *testing.T) {
	cases := []struct {
		name        string
		expectedErr string
	}{
		{
			name:        "err_test_getExpectedOverrideHash_fail",
			expectedErr: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fakecm := fakekubeutils.FakeConfigMap{}
			mockConfigMapWrapper := kubemock.NewMockConfigMapWrapper(ctrl)
			patchGetData, err := mpatch.PatchMethod(fakecm.GetData, mockConfigMapWrapper.GetData)
			if err != nil {
				t.Fatal("patch error:", err)
			}
			defer unpatch(t, patchGetData)
			mockConfigMapWrapper.EXPECT().GetData().AnyTimes().Return(nil)

			patchLoadSchemaStructFromYaml, err := mpatch.PatchMethod(eputils.LoadSchemaStructFromYaml, func(eputils.SchemaStruct, string) error {
				return kubeerr
			})
			if err != nil {
				t.Fatal("patch error:", err)
			}
			defer unpatch(t, patchLoadSchemaStructFromYaml)

			if got := getExpectedOverrideHash(&fakecm, ""); !reflect.DeepEqual(got, tc.expectedErr) {
				t.Errorf("get service value hash: %v, want %v", got, tc.expectedErr)
			} else {
				t.Log("Error expected.")
			}
		})
	}
}

func Test_getRevision(t *testing.T) {
	cases := []struct {
		name        string
		expectedErr error
	}{
		{
			name:        "err_test_getRevision_fail",
			expectedErr: eputils.GetError("errWrongStatus"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fakecm := fakeDeployer{}
			patchHelmStatus, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakecm), "HelmStatus",
				func(*fakeDeployer, string) (string, int) { return "wrong", 0 })
			if err != nil {
				t.Fatal("patch error:", err)
			}
			defer unpatch(t, patchHelmStatus)

			if _, err := getRevision(&fakecm, "", "testhelm"); !reflect.DeepEqual(err, tc.expectedErr) {
				t.Errorf("get service revision:  %v, want %v", err, tc.expectedErr)
			} else {
				t.Log("Error expected.")
			}
		})
	}
}

func Test_updateConfigmap(t *testing.T) {
	func_err_SchemaStructToYaml := func(ctrl *gomock.Controller) []*mpatch.Patch {
		patchSchemaStructToYaml, err := mpatch.PatchMethod(eputils.SchemaStructToYaml, func(eputils.SchemaStruct) (string, error) {
			return "", kubeerr
		})
		if err != nil {
			t.Fatal("patch error:", err)
		}
		return []*mpatch.Patch{patchSchemaStructToYaml}
	}

	func_err_RenewData := func(ctrl *gomock.Controller) []*mpatch.Patch {
		fakecm := fakekubeutils.FakeConfigMap{}
		patchSchemaStructToYaml, err := mpatch.PatchMethod(eputils.SchemaStructToYaml, func(eputils.SchemaStruct) (string, error) {
			return "", nil
		})
		if err != nil {
			t.Fatal("patch error:", err)
		}
		patchRenewData, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&fakecm), "RenewData",
			func(*fakekubeutils.FakeConfigMap, string, string) error { return kubeerr })
		if err != nil {
			t.Fatal("Patch err:", err)
		}
		return []*mpatch.Patch{patchSchemaStructToYaml, patchRenewData}
	}
	cases := []struct {
		name           string
		input          map[string][]byte
		expectedErr    error
		funcBeforeTest func(*gomock.Controller) []*mpatch.Patch
	}{
		{
			name: "err_test_SchemaStructToYaml_fail",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"components":[{"clients":["default"],"name":"testhelm","type":"helm","chartoverride":"o","url":"u"}]}`),
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_SchemaStructToYaml,
		},
		{
			name: "err_test_updateConfigmap_fail",
			input: map[string][]byte{
				"ep-params":     []byte(`{"runtimedir":"","kubeconfig":""}`),
				"serviceconfig": []byte(`{"components":[{"clients":["default"],"name":"testhelm","type":"helm","chartoverride":"o","url":"u"}]}`),
			},
			expectedErr:    kubeerr,
			funcBeforeTest: func_err_RenewData,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if tc.funcBeforeTest != nil {
				plist := tc.funcBeforeTest(ctrl)
				for _, p := range plist {
					defer unpatch(t, p)
				}
			}
			fakecm := fakekubeutils.FakeConfigMap{}
			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			service := input_serviceconfig(input).Components[0]
			if service == nil {
				t.Fatalf("Failed to generate service")
			}

			if err := updateConfigmap(service, &fakecm, ""); !reflect.DeepEqual(err, tc.expectedErr) {
				t.Errorf("update config map: %v, want %v", err, tc.expectedErr)
			} else {
				t.Log("Error expected.")
			}
		})
	}
}
